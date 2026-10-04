// Package sponsors implements the CatX-owned, authenticated sponsor surface.
// It deliberately does not import the upstream panel sponsor service: the
// operator chooses the metadata source and CatX applies its own safety policy.
package sponsors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"gorm.io/gorm"
)

const (
	sourceSettingKey  = "fork.sponsors.source_url"
	contactSettingKey = "fork.sponsors.contact_url"
	logoPathPrefix    = "/panel/api/fork/sponsors/logo/"

	metadataTTL        = time.Hour
	errorTTL           = 10 * time.Minute
	maxMetadataBytes   = 256 << 10
	maxLogoBytes       = 256 << 10
	maxSponsors        = 100
	maxSidebarSlots    = 3
	maxRemoteURLLength = 2048
	maxTextLength      = 500
	maxNameLength      = 160
)

var (
	logoNameRE   = regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9._-]{0,127}\.(png|webp|jpg|jpeg)$`)
	idRE         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	allowedSlots = map[string]struct{}{
		"dashboard": {},
		"sidebar":   {},
		"page":      {},
		"login":     {},
	}
)

// Sponsor is the CatX-compatible public metadata shape. It intentionally
// contains no private operator or traffic information.
type Sponsor struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Enable *bool             `json:"enable,omitempty"`
	Slots  []string          `json:"slots"`
	From   *time.Time        `json:"from,omitempty"`
	Until  time.Time         `json:"until"`
	Logo   string            `json:"logo,omitempty"`
	Title  map[string]string `json:"title"`
	Text   map[string]string `json:"text"`
	Link   string            `json:"link"`
}

type SponsorList struct {
	Contact  string    `json:"contact,omitempty"`
	Sponsors []Sponsor `json:"sponsors"`
}

type settings struct {
	sourceURL  string
	contactURL string
}

type cachedList struct {
	source  string
	list    *SponsorList
	err     error
	retryAt time.Time
}

type cachedLogo struct {
	source      string
	data        []byte
	contentType string
	err         error
	retryAt     time.Time
}

var runtimeState struct {
	sync.RWMutex
	db      *gorm.DB
	enabled bool
	list    cachedList
	logos   map[string]cachedLogo
}

// httpClientFactory is replaceable in package tests so local httptest servers
// can be exercised without weakening the production public-address dialer.
var httpClientFactory = newSafeHTTPClient

// Configure publishes the database and feature state at the fork lifecycle
// boundary. Cache state is cleared when either input changes.
func Configure(db *gorm.DB, enabled bool) {
	runtimeState.Lock()
	if runtimeState.db != db || runtimeState.enabled != enabled {
		runtimeState.list = cachedList{}
		runtimeState.logos = nil
	}
	runtimeState.db = db
	runtimeState.enabled = enabled
	runtimeState.Unlock()
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	group := api.Group("/fork/sponsors")
	group.GET("", listHandler)
	group.GET("/logo/:name", logoHandler)
	group.GET("/settings", settingsHandler)
	group.PUT("/settings", updateSettingsHandler)
}

func listHandler(c *gin.Context) {
	list, err := currentSponsors(time.Now())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": list})
}

func logoHandler(c *gin.Context) {
	data, contentType, err := currentLogo(c.Param("name"), time.Now())
	if errors.Is(err, errUnknownLogo) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "sponsor logo not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsor logo unavailable"})
		return
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, contentType, data)
}

type settingsResponse struct {
	Enabled    bool   `json:"enabled"`
	Configured bool   `json:"configured"`
	SourceURL  string `json:"sourceUrl"`
	ContactURL string `json:"contactUrl"`
}

func settingsHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors settings unavailable"})
		return
	}
	current, err := readSettings(db)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors settings unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": settingsResponse{
		Enabled: enabled, Configured: current.sourceURL != "", SourceURL: current.sourceURL, ContactURL: current.contactURL,
	}})
}

func updateSettingsHandler(c *gin.Context) {
	db, _ := runtime()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors settings unavailable"})
		return
	}
	var input struct {
		SourceURL  string `json:"sourceUrl"`
		ContactURL string `json:"contactUrl"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid sponsors settings"})
		return
	}
	next, err := validateSettings(settings{sourceURL: input.SourceURL, contactURL: input.ContactURL})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	if err := saveSettings(db, next); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsors settings save failed"})
		return
	}
	invalidateCache()
	_, enabled := runtime()
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "sponsors settings saved", "obj": settingsResponse{
		Enabled: enabled, Configured: next.sourceURL != "", SourceURL: next.sourceURL, ContactURL: next.contactURL,
	}})
}

func runtime() (*gorm.DB, bool) {
	runtimeState.RLock()
	defer runtimeState.RUnlock()
	return runtimeState.db, runtimeState.enabled
}

func invalidateCache() {
	runtimeState.Lock()
	runtimeState.list = cachedList{}
	runtimeState.logos = nil
	runtimeState.Unlock()
}

func readSettings(db *gorm.DB) (settings, error) {
	if db == nil {
		return settings{}, errors.New("database unavailable")
	}
	read := func(key string) (string, error) {
		var row model.Setting
		err := db.Where("key = ?", key).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return row.Value, err
	}
	source, err := read(sourceSettingKey)
	if err != nil {
		return settings{}, err
	}
	contact, err := read(contactSettingKey)
	if err != nil {
		return settings{}, err
	}
	return validateSettings(settings{sourceURL: source, contactURL: contact})
}

func saveSettings(db *gorm.DB, value settings) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for key, next := range map[string]string{
			sourceSettingKey:  value.sourceURL,
			contactSettingKey: value.contactURL,
		} {
			var row model.Setting
			err := tx.Where("key = ?", key).First(&row).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if next == "" {
					continue
				}
				if err := tx.Create(&model.Setting{Key: key, Value: next}).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			row.Value = next
			if err := tx.Save(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func validateSettings(value settings) (settings, error) {
	var err error
	if value.sourceURL, err = normalizedRemoteURL(value.sourceURL); err != nil {
		return settings{}, fmt.Errorf("invalid source URL: %w", err)
	}
	if value.contactURL != "" {
		var parsed *url.URL
		if parsed, err = normalizedHTTPSURL(value.contactURL); err != nil {
			return settings{}, fmt.Errorf("invalid contact URL: %w", err)
		}
		value.contactURL = parsed.String()
	}
	return value, nil
}

func currentSponsors(now time.Time) (SponsorList, error) {
	db, enabled := runtime()
	if !enabled {
		return SponsorList{Sponsors: []Sponsor{}}, nil
	}
	config, err := readSettings(db)
	if err != nil {
		return SponsorList{}, err
	}
	if config.sourceURL == "" {
		return SponsorList{Contact: config.contactURL, Sponsors: []Sponsor{}}, nil
	}

	runtimeState.RLock()
	cached := runtimeState.list
	runtimeState.RUnlock()
	if cached.source == config.sourceURL && now.Before(cached.retryAt) {
		if cached.list != nil {
			return activeSponsors(cached.list, config.contactURL, now), nil
		}
		return SponsorList{}, cached.err
	}

	fetched, fetchErr := fetchSponsors(config.sourceURL)
	runtimeState.Lock()
	defer runtimeState.Unlock()
	if fetchErr == nil {
		runtimeState.list = cachedList{source: config.sourceURL, list: fetched, retryAt: now.Add(metadataTTL)}
		return activeSponsors(fetched, config.contactURL, now), nil
	}
	if cached.source == config.sourceURL && cached.list != nil {
		runtimeState.list = cachedList{source: config.sourceURL, list: cached.list, err: fetchErr, retryAt: now.Add(errorTTL)}
		return activeSponsors(cached.list, config.contactURL, now), nil
	}
	runtimeState.list = cachedList{source: config.sourceURL, err: fetchErr, retryAt: now.Add(errorTTL)}
	return SponsorList{}, fetchErr
}

func activeSponsors(raw *SponsorList, contact string, now time.Time) SponsorList {
	out := SponsorList{Contact: contact, Sponsors: []Sponsor{}}
	if raw == nil {
		return out
	}
	sidebarTaken := 0
	for _, sponsor := range raw.Sponsors {
		if len(out.Sponsors) >= maxSponsors || (sponsor.Enable != nil && !*sponsor.Enable) || sponsor.ID == "" || sponsor.Name == "" || sponsor.Until.IsZero() || !now.Before(sponsor.Until) || (sponsor.From != nil && now.Before(*sponsor.From)) {
			continue
		}
		if !idRE.MatchString(sponsor.ID) || len([]rune(sponsor.Name)) > maxNameLength {
			continue
		}
		if _, err := normalizedHTTPSURL(sponsor.Link); err != nil {
			continue
		}
		slots := make([]string, 0, len(sponsor.Slots))
		seen := make(map[string]struct{}, len(sponsor.Slots))
		for _, slot := range sponsor.Slots {
			if _, ok := allowedSlots[slot]; !ok {
				continue
			}
			if _, ok := seen[slot]; ok || (slot == "sidebar" && sidebarTaken >= maxSidebarSlots) {
				continue
			}
			seen[slot] = struct{}{}
			slots = append(slots, slot)
		}
		if len(slots) == 0 {
			continue
		}
		if sponsor.Logo != "" {
			if !logoNameRE.MatchString(sponsor.Logo) {
				sponsor.Logo = ""
			} else {
				sponsor.Logo = logoPathPrefix + sponsor.Logo
			}
		}
		sponsor.Title = boundedLocaleMap(sponsor.Title)
		sponsor.Text = boundedLocaleMap(sponsor.Text)
		sponsor.Slots = slots
		if contains(slots, "sidebar") {
			sidebarTaken++
		}
		out.Sponsors = append(out.Sponsors, sponsor)
	}
	return out
}

func boundedLocaleMap(values map[string]string) map[string]string {
	result := make(map[string]string)
	for locale, value := range values {
		if len(result) >= 16 || len(locale) == 0 || len(locale) > 16 {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || len([]rune(value)) > maxTextLength {
			continue
		}
		result[locale] = value
	}
	return result
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

var errUnknownLogo = errors.New("unknown sponsor logo")

func currentLogo(name string, now time.Time) ([]byte, string, error) {
	if !logoNameRE.MatchString(name) {
		return nil, "", errUnknownLogo
	}
	db, enabled := runtime()
	if !enabled {
		return nil, "", errUnknownLogo
	}
	config, err := readSettings(db)
	if err != nil {
		return nil, "", err
	}
	if config.sourceURL == "" {
		return nil, "", errUnknownLogo
	}
	list, err := currentSponsors(now)
	if err != nil {
		return nil, "", err
	}
	wanted := logoPathPrefix + name
	known := false
	for _, sponsor := range list.Sponsors {
		if sponsor.Logo == wanted {
			known = true
			break
		}
	}
	if !known {
		return nil, "", errUnknownLogo
	}

	runtimeState.RLock()
	cached, ok := runtimeState.logos[name]
	runtimeState.RUnlock()
	if ok && cached.source == config.sourceURL && now.Before(cached.retryAt) {
		return cached.data, cached.contentType, cached.err
	}
	logoURL, err := logoRemoteURL(config.sourceURL, name)
	if err != nil {
		return nil, "", errUnknownLogo
	}
	data, contentType, fetchErr := fetchLogo(logoURL)
	runtimeState.Lock()
	defer runtimeState.Unlock()
	if fetchErr == nil {
		if runtimeState.logos == nil {
			runtimeState.logos = make(map[string]cachedLogo)
		}
		runtimeState.logos[name] = cachedLogo{source: config.sourceURL, data: data, contentType: contentType, retryAt: now.Add(metadataTTL)}
		return data, contentType, nil
	}
	if ok && cached.source == config.sourceURL && len(cached.data) > 0 {
		cached.retryAt = now.Add(errorTTL)
		runtimeState.logos[name] = cached
		return cached.data, cached.contentType, nil
	}
	if runtimeState.logos == nil {
		runtimeState.logos = make(map[string]cachedLogo)
	}
	runtimeState.logos[name] = cachedLogo{source: config.sourceURL, err: fetchErr, retryAt: now.Add(errorTTL)}
	return nil, "", fetchErr
}

func fetchSponsors(source string) (*SponsorList, error) {
	body, err := fetchLimited(source, maxMetadataBytes)
	if err != nil {
		return nil, err
	}
	var list SponsorList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("decode sponsors metadata: %w", err)
	}
	return &list, nil
}

func fetchLogo(source string) ([]byte, string, error) {
	body, err := fetchLimited(source, maxLogoBytes)
	if err != nil {
		return nil, "", err
	}
	contentType := http.DetectContentType(body)
	switch contentType {
	case "image/png", "image/webp", "image/jpeg":
		return body, contentType, nil
	default:
		return nil, "", fmt.Errorf("unsupported sponsor logo content type %s", contentType)
	}
}

func fetchLimited(raw string, limit int) ([]byte, error) {
	client := httpClientFactory()
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote source returned status %d", response.StatusCode)
	}
	if response.ContentLength > int64(limit) {
		return nil, fmt.Errorf("remote source exceeds %d bytes", limit)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > limit {
		return nil, fmt.Errorf("remote source exceeds %d bytes", limit)
	}
	return body, nil
}

func logoRemoteURL(source, name string) (string, error) {
	normalized, err := normalizedRemoteURL(source)
	if err != nil || !logoNameRE.MatchString(name) {
		return "", errUnknownLogo
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", errUnknownLogo
	}
	parsed.Path = path.Join(path.Dir(parsed.Path), "logos", name)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func normalizedRemoteURL(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	parsed, err := normalizedHTTPSURL(raw)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func normalizedHTTPSURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	if len(raw) > maxRemoteURLLength {
		return nil, errors.New("URL is too long")
	}
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return nil, errors.New("URL must be HTTPS without credentials")
	}
	if port := parsed.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 {
			return nil, errors.New("URL has an invalid port")
		}
	}
	return parsed, nil
}

func newSafeHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           safeDialContext,
		MaxIdleConns:          4,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("too many redirects")
			}
			if _, err := normalizedHTTPSURL(request.URL.String()); err != nil {
				return fmt.Errorf("unsafe redirect: %w", err)
			}
			return nil
		},
	}
}

func safeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var addresses []net.IPAddr
	if parsed := net.ParseIP(host); parsed != nil {
		addresses = []net.IPAddr{{IP: parsed}}
	} else {
		addresses, err = net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	for _, candidate := range addresses {
		if !isPublicIP(candidate.IP) {
			continue
		}
		connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		if dialErr == nil {
			return connection, nil
		}
		err = dialErr
	}
	if err == nil {
		err = errors.New("remote host has no public address")
	}
	return nil, err
}

func isPublicIP(raw net.IP) bool {
	address, err := netip.ParseAddr(raw.String())
	if err != nil {
		return false
	}
	address = address.Unmap()
	return !address.IsPrivate() && !address.IsLoopback() && !address.IsLinkLocalUnicast() && !address.IsLinkLocalMulticast() && !address.IsUnspecified() && !address.IsMulticast()
}
