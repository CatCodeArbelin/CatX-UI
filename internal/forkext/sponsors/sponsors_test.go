package sponsors

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t rewriteTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	copyRequest := request.Clone(request.Context())
	copyURL := *request.URL
	copyURL.Scheme = t.target.Scheme
	copyURL.Host = t.target.Host
	copyRequest.URL = &copyURL
	return t.base.RoundTrip(copyRequest)
}

func sponsorsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skipf("sqlite unavailable: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func useTestClient(t *testing.T, serverURL string) {
	t.Helper()
	target, err := url.Parse(serverURL)
	if err != nil {
		t.Fatal(err)
	}
	httpClientFactory = func() *http.Client {
		return &http.Client{Transport: rewriteTransport{target: target, base: http.DefaultTransport}}
	}
	t.Cleanup(func() { httpClientFactory = newSafeHTTPClient })
}

func configureTestRuntime(t *testing.T, db *gorm.DB, source string, enabled bool) {
	t.Helper()
	if source != "" {
		if err := saveSettings(db, settings{sourceURL: source}); err != nil {
			t.Fatal(err)
		}
	}
	Configure(db, enabled)
	t.Cleanup(func() { Configure(nil, false) })
}

func TestDisabledSponsorsDoNotFetchRemoteData(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()
	db := sponsorsTestDB(t)
	configureTestRuntime(t, db, "https://public.example/feed.json", false)
	useTestClient(t, server.URL)

	list, err := currentSponsors(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Sponsors) != 0 || requests.Load() != 0 {
		t.Fatalf("disabled sponsors = %#v, requests = %d", list, requests.Load())
	}
}

func TestUnconfiguredSponsorsDoNotFetchRemoteData(t *testing.T) {
	db := sponsorsTestDB(t)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })
	list, err := currentSponsors(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Sponsors) != 0 {
		t.Fatalf("unconfigured sponsors = %#v", list)
	}
}

func TestActiveSponsorsAreValidatedAndCached(t *testing.T) {
	var requests atomic.Int32
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/feed.json" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(SponsorList{Sponsors: []Sponsor{
			{ID: "active", Name: "Active", Slots: []string{"dashboard", "sidebar", "sidebar"}, Until: now.Add(time.Hour), Logo: "brand.png", Title: map[string]string{"en": "Active"}, Link: "https://brand.example/"},
			{ID: "expired", Name: "Expired", Slots: []string{"page"}, Until: now.Add(-time.Minute), Link: "https://brand.example/"},
			{ID: "bad-link", Name: "Bad link", Slots: []string{"page"}, Until: now.Add(time.Hour), Link: "http://brand.example/"},
		}})
	}))
	defer server.Close()
	useTestClient(t, server.URL)
	db := sponsorsTestDB(t)
	configureTestRuntime(t, db, "https://public.example/feed.json", true)

	list, err := currentSponsors(now)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Sponsors) != 1 || list.Sponsors[0].Logo != logoPathPrefix+"brand.png" {
		t.Fatalf("active list = %#v", list)
	}
	if len(list.Sponsors[0].Slots) != 2 {
		t.Fatalf("duplicate slots were not removed: %#v", list.Sponsors[0].Slots)
	}
	if _, err := currentSponsors(now.Add(30 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatalf("cache requests = %d, want 1", requests.Load())
	}
}

func TestFetchFailureKeepsKnownGoodListDuringRetryWindow(t *testing.T) {
	var requests atomic.Int32
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) > 1 {
			http.Error(w, "temporary failure", http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(SponsorList{Sponsors: []Sponsor{{
			ID: "cached", Name: "Cached", Slots: []string{"page"}, Until: now.Add(2 * time.Hour), Link: "https://brand.example/",
		}}})
	}))
	defer server.Close()
	useTestClient(t, server.URL)
	db := sponsorsTestDB(t)
	configureTestRuntime(t, db, "https://public.example/feed.json", true)
	if _, err := currentSponsors(now); err != nil {
		t.Fatal(err)
	}
	list, err := currentSponsors(now.Add(metadataTTL + time.Second))
	if err != nil || len(list.Sponsors) != 1 || list.Sponsors[0].ID != "cached" {
		t.Fatalf("fallback list = %#v, err = %v", list, err)
	}
	if requests.Load() != 2 {
		t.Fatalf("fetch count = %d, want 2", requests.Load())
	}
}

func TestLogoIsDerivedFromSourceAndBoundToActiveMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/feed.json":
			_ = json.NewEncoder(w).Encode(SponsorList{Sponsors: []Sponsor{{
				ID: "brand", Name: "Brand", Slots: []string{"page"}, Until: time.Now().Add(time.Hour), Logo: "brand.png", Link: "https://brand.example/",
			}}})
		case "/logos/brand.png":
			w.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	useTestClient(t, server.URL)
	db := sponsorsTestDB(t)
	configureTestRuntime(t, db, "https://public.example/feed.json", true)

	data, contentType, err := currentLogo("brand.png", time.Now())
	if err != nil || contentType != "image/png" || len(data) == 0 {
		t.Fatalf("logo = %q/%q, err = %v", data, contentType, err)
	}
	if _, _, err := currentLogo("../secret.png", time.Now()); !errors.Is(err, errUnknownLogo) {
		t.Fatalf("path traversal error = %v", err)
	}
}

func TestSettingsRejectCredentialsAndNonHTTPS(t *testing.T) {
	for _, value := range []settings{
		{sourceURL: "http://public.example/feed.json"},
		{sourceURL: "https://user:pass@public.example/feed.json"},
		{sourceURL: "https://public.example:bad/feed.json"},
		{contactURL: "http://public.example/contact"},
	} {
		if _, err := validateSettings(value); err == nil {
			t.Fatalf("validateSettings(%#v) unexpectedly succeeded", value)
		}
	}
	valid, err := validateSettings(settings{sourceURL: " https://public.example/feed.json ", contactURL: "https://public.example/contact"})
	if err != nil || valid.sourceURL != "https://public.example/feed.json" {
		t.Fatalf("valid settings = %#v, err = %v", valid, err)
	}
}

func TestMetadataFetchRejectsMalformedAndOversizedBodies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/large" {
			_, _ = w.Write([]byte(strings.Repeat("x", maxMetadataBytes+1)))
			return
		}
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()
	useTestClient(t, server.URL)
	if _, err := fetchSponsors("https://public.example/large"); err == nil {
		t.Fatal("oversized metadata unexpectedly succeeded")
	}
	if _, err := fetchSponsors("https://public.example/malformed"); err == nil {
		t.Fatal("malformed metadata unexpectedly succeeded")
	}
}

func TestSafeDialerRejectsPrivateAddresses(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1:443",
		"10.0.0.1:443",
		"100.64.0.1:443",
		"192.0.2.1:443",
		"198.18.0.1:443",
		"198.51.100.1:443",
		"203.0.113.1:443",
		"224.0.0.1:443",
		"240.0.0.1:443",
		"[::1]:443",
		"[fc00::1]:443",
		"[fe80::1]:443",
		"[ff02::1]:443",
		"[2001:db8::1]:443",
		"[::ffff:192.0.2.1]:443",
	} {
		if _, err := safeDialContext(context.Background(), "tcp", address); err == nil {
			t.Fatalf("safeDialContext(%q) unexpectedly succeeded", address)
		}
	}
}

func TestIsPublicIPRejectsSpecialPurposeAddresses(t *testing.T) {
	for _, raw := range []string{
		"0.0.0.1",
		"10.0.0.1",
		"100.64.0.1",
		"127.0.0.1",
		"169.254.1.1",
		"172.16.0.1",
		"192.0.0.1",
		"192.31.196.1",
		"192.52.193.1",
		"192.88.99.1",
		"192.168.0.1",
		"198.18.0.1",
		"198.51.100.1",
		"203.0.113.1",
		"224.0.0.1",
		"240.0.0.1",
		"::1",
		"fc00::1",
		"fe80::1",
		"ff00::1",
		"2001:2::1",
		"2001:10::1",
		"2001:20::1",
		"2001:3::1",
		"2002::1",
		"3fff::1",
		"100::1",
		"::ffff:198.51.100.1",
	} {
		if isPublicIP(net.ParseIP(raw)) {
			t.Errorf("isPublicIP(%q) = true, want false", raw)
		}
	}
	for _, raw := range []string{"1.1.1.1", "2606:4700:4700::1111"} {
		if !isPublicIP(net.ParseIP(raw)) {
			t.Errorf("isPublicIP(%q) = false, want true", raw)
		}
	}
}

func TestSafeClientRedirectPolicyRejectsUnsafeURLForms(t *testing.T) {
	client := newSafeHTTPClient()
	for _, raw := range []string{
		"http://public.example/feed.json",
		"https://user:pass@public.example/feed.json",
	} {
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("parse %q: %v", raw, err)
		}
		request := &http.Request{URL: parsed}
		if err := client.CheckRedirect(request, nil); err == nil {
			t.Fatalf("redirect to %q unexpectedly succeeded", raw)
		}
	}
}

func TestCatxSlotsDoNotAdvertiseUnauthenticatedLogin(t *testing.T) {
	raw := &SponsorList{Sponsors: []Sponsor{{
		ID: "login", Name: "Login", Slots: []string{"login", "page"}, Until: time.Now().Add(time.Hour), Link: "https://brand.example/",
	}}}
	got := activeSponsors(raw, "", time.Now())
	if len(got.Sponsors) != 1 || len(got.Sponsors[0].Slots) != 1 || got.Sponsors[0].Slots[0] != "page" {
		t.Fatalf("active sponsor slots = %#v, want only page", got.Sponsors)
	}
}

func TestLogoRemoteURLDoesNotCarryQueryOrPathInput(t *testing.T) {
	got, err := logoRemoteURL("https://public.example/a/feed.json?token=secret", "brand.png")
	if err != nil || got != "https://public.example/a/logos/brand.png" {
		t.Fatalf("logo URL = %q, err = %v", got, err)
	}
	if strings.Contains(got, "secret") {
		t.Fatal("logo URL leaked source query")
	}
}

func TestRegisterRoutesUsesProtectedForkNamespace(t *testing.T) {
	router := gin.New()
	api := router.Group("/panel/api")
	RegisterRoutes(api)
	paths := make(map[string]bool)
	for _, route := range router.Routes() {
		paths[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"GET /panel/api/fork/sponsors",
		"GET /panel/api/fork/sponsors/logo/:name",
		"GET /panel/api/fork/sponsors/settings",
		"PUT /panel/api/fork/sponsors/settings",
	} {
		if !paths[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
