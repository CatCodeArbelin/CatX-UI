package sponsors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
)

type ProviderMode string

const (
	ProviderLocal  ProviderMode = "local"
	ProviderRemote ProviderMode = "remote"
)

var localeRE = regexp.MustCompile(`^[A-Za-z]{2,8}(?:-[A-Za-z0-9]{2,8})?$`)

var supportedSlots = []string{"dashboard", "sidebar", "page"}

type sponsorRecordInput struct {
	ID             string            `json:"id"`
	Enabled        *bool             `json:"enabled"`
	Name           string            `json:"name"`
	Priority       int               `json:"priority"`
	Slots          []string          `json:"slots"`
	StartAt        string            `json:"startAt"`
	EndAt          string            `json:"endAt"`
	DestinationURL string            `json:"destinationUrl"`
	LogoURL        string            `json:"logoUrl"`
	Title          map[string]string `json:"title"`
	Text           map[string]string `json:"text"`
}

type managedSponsor struct {
	ID             string            `json:"id"`
	Enabled        bool              `json:"enabled"`
	Name           string            `json:"name"`
	Priority       int               `json:"priority"`
	Slots          []string          `json:"slots"`
	StartAt        string            `json:"startAt,omitempty"`
	EndAt          string            `json:"endAt,omitempty"`
	DestinationURL string            `json:"destinationUrl"`
	LogoURL        string            `json:"logoUrl,omitempty"`
	Title          map[string]string `json:"title"`
	Text           map[string]string `json:"text"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

type sponsorStatus struct {
	Enabled                  bool         `json:"enabled"`
	ProviderMode             ProviderMode `json:"providerMode"`
	LocalSponsorCount        int64        `json:"localSponsorCount"`
	ActiveSponsorCount       int64        `json:"activeSponsorCount"`
	RemoteProviderConfigured bool         `json:"remoteProviderConfigured"`
	LastSuccessfulFetch      *time.Time   `json:"lastSuccessfulFetch,omitempty"`
	LastError                string       `json:"lastError,omitempty"`
	CacheState               string       `json:"cacheState"`
}

func registerManagementRoutes(group *gin.RouterGroup) {
	group.GET("/status", statusHandler)
	group.GET("/manage", managedListHandler)
	group.POST("", createManagedHandler)
	group.GET("/:id", getManagedHandler)
	group.PUT("/:id", updateManagedHandler)
	group.DELETE("/:id", deleteManagedHandler)
}

func statusHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors status unavailable"})
		return
	}
	config, err := readSettings(db)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors status unavailable"})
		return
	}
	var localCount int64
	if err := db.Model(&Record{}).Count(&localCount).Error; err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors status unavailable"})
		return
	}
	activeCount := int64(0)
	if enabled && config.providerMode == ProviderLocal {
		list, listErr := localSponsorList(db, config.contactURL, time.Now())
		if listErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors status unavailable"})
			return
		}
		activeCount = int64(len(list.Sponsors))
	}
	runtimeState.RLock()
	lastFetch := runtimeState.lastFetch
	lastError := runtimeState.lastError
	cacheState := "empty"
	if config.providerMode == ProviderLocal {
		cacheState = "local"
	} else if !lastFetch.IsZero() {
		cacheState = "fresh"
	} else if lastError != "" {
		cacheState = "error"
	}
	runtimeState.RUnlock()
	if !enabled {
		cacheState = "disabled"
	}
	var fetched *time.Time
	if !lastFetch.IsZero() {
		value := lastFetch
		fetched = &value
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": sponsorStatus{
		Enabled: enabled, ProviderMode: config.providerMode, LocalSponsorCount: localCount,
		ActiveSponsorCount: activeCount, RemoteProviderConfigured: config.sourceURL != "",
		LastSuccessfulFetch: fetched, LastError: lastError, CacheState: cacheState,
	}})
}

func managedListHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil || !enabled {
		featureDisabledResponse(c)
		return
	}
	items, err := listManaged(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsors list failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": items})
}

func createManagedHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil || !enabled {
		featureDisabledResponse(c)
		return
	}
	if !requireLocalProvider(c, db) {
		return
	}
	var input sponsorRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid sponsor"})
		return
	}
	record, err := normalizeRecordInput(input, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		return recordAudit(c, tx, "create", record.ID)
	}); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "msg": "sponsor ID already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor create failed"})
		return
	}
	view, err := toManagedSponsor(*record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor response failed"})
		return
	}
	invalidateCache()
	c.JSON(http.StatusCreated, gin.H{"success": true, "msg": "sponsor created", "obj": view})
}

func getManagedHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil || !enabled {
		featureDisabledResponse(c)
		return
	}
	var record Record
	if err := db.Where("id = ?", c.Param("id")).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "sponsor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor read failed"})
		return
	}
	view, err := toManagedSponsor(record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor response failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": view})
}

func updateManagedHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil || !enabled {
		featureDisabledResponse(c)
		return
	}
	if !requireLocalProvider(c, db) {
		return
	}
	var input sponsorRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid sponsor"})
		return
	}
	var existing Record
	if err := db.Where("id = ?", c.Param("id")).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "sponsor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor read failed"})
		return
	}
	input.ID = existing.ID
	record, err := normalizeRecordInput(input, &existing)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	action := "update"
	if existing.Enabled != record.Enabled {
		if record.Enabled {
			action = "enable"
		} else {
			action = "disable"
		}
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(record).Error; err != nil {
			return err
		}
		return recordAudit(c, tx, action, record.ID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor update failed"})
		return
	}
	view, err := toManagedSponsor(*record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor response failed"})
		return
	}
	invalidateCache()
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "sponsor updated", "obj": view})
}

func deleteManagedHandler(c *gin.Context) {
	db, enabled := runtime()
	if db == nil || !enabled {
		featureDisabledResponse(c)
		return
	}
	if !requireLocalProvider(c, db) {
		return
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&Record{}, "id = ?", c.Param("id"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return recordAudit(c, tx, "delete", c.Param("id"))
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "sponsor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "sponsor delete failed"})
		return
	}
	invalidateCache()
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "sponsor deleted"})
}

func featureDisabledResponse(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "featureDisabled": true, "msg": "sponsors feature is disabled"})
}

func requireLocalProvider(c *gin.Context, db *gorm.DB) bool {
	config, err := readSettings(db)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "sponsors settings unavailable"})
		return false
	}
	if config.providerMode == ProviderRemote {
		c.JSON(http.StatusConflict, gin.H{"success": false, "msg": "switch sponsors provider to local before changing records"})
		return false
	}
	return true
}

func listManaged(db *gorm.DB) ([]managedSponsor, error) {
	var records []Record
	if err := db.Order("priority ASC").Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]managedSponsor, 0, len(records))
	for _, record := range records {
		view, err := toManagedSponsor(record)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

func localSponsorList(db *gorm.DB, contact string, now time.Time) (SponsorList, error) {
	var records []Record
	if err := db.Where("enabled = ?", true).Order("priority ASC").Order("id ASC").Find(&records).Error; err != nil {
		return SponsorList{}, err
	}
	result := SponsorList{Contact: contact, Sponsors: make([]Sponsor, 0, len(records))}
	sidebarTaken := 0
	for _, record := range records {
		if !recordActive(record, now) {
			continue
		}
		value, err := toPublicSponsor(record, now)
		if err != nil {
			return SponsorList{}, err
		}
		value.Slots = cappedSlots(value.Slots, &sidebarTaken)
		if len(value.Slots) == 0 {
			continue
		}
		result.Sponsors = append(result.Sponsors, value)
		if len(result.Sponsors) >= maxSponsors {
			break
		}
	}
	return result, nil
}

func cappedSlots(values []string, sidebarTaken *int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := allowedSlots[value]; !ok {
			continue
		}
		if _, ok := seen[value]; ok || (value == "sidebar" && *sidebarTaken >= maxSidebarSlots) {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	if contains(result, "sidebar") {
		(*sidebarTaken)++
	}
	return result
}

func localSponsorLogo(db *gorm.DB, id string, now time.Time) ([]byte, string, error) {
	if !idRE.MatchString(id) {
		return nil, "", errUnknownLogo
	}
	var record Record
	if err := db.Where("id = ?", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errUnknownLogo
		}
		return nil, "", err
	}
	if !recordActive(record, now) || record.LogoURL == "" {
		return nil, "", errUnknownLogo
	}
	runtimeState.RLock()
	cached, ok := runtimeState.logos[id]
	runtimeState.RUnlock()
	if ok && cached.source == record.LogoURL && now.Before(cached.retryAt) {
		return cached.data, cached.contentType, cached.err
	}
	data, contentType, fetchErr := fetchLogo(record.LogoURL)
	runtimeState.Lock()
	defer runtimeState.Unlock()
	if fetchErr == nil {
		if runtimeState.logos == nil {
			runtimeState.logos = make(map[string]cachedLogo)
		}
		runtimeState.logos[id] = cachedLogo{source: record.LogoURL, data: data, contentType: contentType, retryAt: now.Add(metadataTTL)}
		return data, contentType, nil
	}
	if ok && cached.source == record.LogoURL && len(cached.data) > 0 {
		cached.retryAt = now.Add(errorTTL)
		runtimeState.logos[id] = cached
		return cached.data, cached.contentType, nil
	}
	if runtimeState.logos == nil {
		runtimeState.logos = make(map[string]cachedLogo)
	}
	runtimeState.logos[id] = cachedLogo{source: record.LogoURL, err: fetchErr, retryAt: now.Add(errorTTL)}
	return nil, "", fetchErr
}

func recordActive(record Record, now time.Time) bool {
	return record.Enabled && (record.StartAt == nil || !now.Before(*record.StartAt)) && (record.EndAt == nil || now.Before(*record.EndAt))
}

func toPublicSponsor(record Record, now time.Time) (Sponsor, error) {
	view, err := toManagedSponsor(record)
	if err != nil {
		return Sponsor{}, err
	}
	var until time.Time
	if record.EndAt != nil {
		until = *record.EndAt
	} else {
		until = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)
	}
	enabled := true
	logo := ""
	if record.LogoURL != "" {
		logo = logoPathPrefix + record.ID
	}
	return Sponsor{ID: record.ID, Name: record.Name, Enable: &enabled, Slots: view.Slots, From: record.StartAt, Until: until, Logo: logo, Title: view.Title, Text: view.Text, Link: record.DestinationURL}, nil
}

func toManagedSponsor(record Record) (managedSponsor, error) {
	var slots []string
	var title, text map[string]string
	if err := json.Unmarshal([]byte(record.SlotsJSON), &slots); err != nil {
		return managedSponsor{}, fmt.Errorf("decode sponsor slots: %w", err)
	}
	if err := json.Unmarshal([]byte(record.TitleJSON), &title); err != nil {
		return managedSponsor{}, fmt.Errorf("decode sponsor title: %w", err)
	}
	if err := json.Unmarshal([]byte(record.TextJSON), &text); err != nil {
		return managedSponsor{}, fmt.Errorf("decode sponsor text: %w", err)
	}
	view := managedSponsor{ID: record.ID, Enabled: record.Enabled, Name: record.Name, Priority: record.Priority, Slots: slots, DestinationURL: record.DestinationURL, LogoURL: record.LogoURL, Title: title, Text: text, CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: record.UpdatedAt.UTC().Format(time.RFC3339)}
	if record.StartAt != nil {
		view.StartAt = record.StartAt.UTC().Format(time.RFC3339)
	}
	if record.EndAt != nil {
		view.EndAt = record.EndAt.UTC().Format(time.RFC3339)
	}
	return view, nil
}

func normalizeRecordInput(input sponsorRecordInput, existing *Record) (*Record, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" && existing != nil {
		id = existing.ID
	}
	if id == "" {
		id = "s-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:24]
	}
	if !idRE.MatchString(id) {
		return nil, errors.New("sponsor ID must contain only letters, digits, dot, underscore, or hyphen")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" && existing != nil {
		name = existing.Name
	}
	if name == "" || len([]rune(name)) > maxNameLength || strings.ContainsAny(name, "<>") {
		return nil, fmt.Errorf("sponsor name must be 1-%d characters without HTML", maxNameLength)
	}
	if input.Priority < -10000 || input.Priority > 10000 {
		return nil, errors.New("sponsor priority must be between -10000 and 10000")
	}
	slots, err := normalizeSlots(input.Slots)
	if err != nil {
		return nil, err
	}
	start, err := parseManagementTime(input.StartAt, "startAt")
	if err != nil {
		return nil, err
	}
	end, err := parseManagementTime(input.EndAt, "endAt")
	if err != nil {
		return nil, err
	}
	if start != nil && end != nil && !start.Before(*end) {
		return nil, errors.New("startAt must be before endAt")
	}
	destination, err := normalizedHTTPSURL(input.DestinationURL)
	if err != nil || destination == nil {
		return nil, errors.New("destinationUrl must be HTTPS without credentials")
	}
	logo := ""
	if strings.TrimSpace(input.LogoURL) != "" {
		parsed, parseErr := normalizedHTTPSURL(input.LogoURL)
		if parseErr != nil || parsed == nil {
			return nil, errors.New("logoUrl must be HTTPS without credentials")
		}
		logo = parsed.String()
	}
	title, err := normalizeLocaleMap(input.Title, name, true)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}
	text, err := normalizeLocaleMap(input.Text, "", false)
	if err != nil {
		return nil, fmt.Errorf("invalid text: %w", err)
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	} else if existing != nil {
		enabled = existing.Enabled
	}
	slotsJSON, _ := json.Marshal(slots)
	titleJSON, _ := json.Marshal(title)
	textJSON, _ := json.Marshal(text)
	record := &Record{ID: id, Enabled: enabled, Name: name, Priority: input.Priority, SlotsJSON: string(slotsJSON), StartAt: start, EndAt: end, DestinationURL: destination.String(), LogoURL: logo, TitleJSON: string(titleJSON), TextJSON: string(textJSON)}
	if existing != nil {
		record.CreatedAt = existing.CreatedAt
	}
	return record, nil
}

func parseManagementTime(raw, field string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be RFC3339", field)
	}
	value = value.UTC()
	return &value, nil
}

func normalizeSlots(values []string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := allowedSlots[value]; !ok {
			return nil, fmt.Errorf("unsupported sponsor slot %q", value)
		}
		seen[value] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for _, value := range supportedSlots {
		if _, ok := seen[value]; ok {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("at least one sponsor slot is required")
	}
	return result, nil
}

func normalizeLocaleMap(values map[string]string, fallback string, required bool) (map[string]string, error) {
	result := make(map[string]string)
	for locale, value := range values {
		locale = strings.TrimSpace(locale)
		value = strings.TrimSpace(value)
		if len(result) >= 16 || !localeRE.MatchString(locale) {
			return nil, errors.New("locale keys are invalid or too numerous")
		}
		if value == "" {
			continue
		}
		if len([]rune(value)) > maxTextLength || strings.ContainsAny(value, "<>") {
			return nil, fmt.Errorf("localized text must be at most %d characters without HTML", maxTextLength)
		}
		result[locale] = value
	}
	if required && len(result) == 0 {
		result["en-US"] = fallback
	}
	keys := make([]string, 0, len(result))
	for key := range result {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]string, len(result))
	for _, key := range keys {
		ordered[key] = result[key]
	}
	return ordered, nil
}

func recordAudit(c *gin.Context, tx *gorm.DB, action, id string) error {
	return recordAuditTarget(c, tx, action, "sponsor", id)
}

func recordAuditTarget(c *gin.Context, tx *gorm.DB, action, targetType, targetRef string) error {
	requestID, _ := c.Get("catx_request_id")
	value, _ := requestID.(string)
	requestContext := context.Background()
	if c.Request != nil && c.Request.Context() != nil {
		requestContext = c.Request.Context()
	}
	return audit.RecordInTx(requestContext, tx, &audit.AuditEvent{EventType: "sponsors." + action, Outcome: audit.EventSuccess, TargetType: targetType, TargetRef: targetRef, RequestID: value, Metadata: `{"source":"catx_sponsors"}`})
}
