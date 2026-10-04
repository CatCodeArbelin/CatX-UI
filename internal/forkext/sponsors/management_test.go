package sponsors

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
)

func managementRouter() *gin.Engine {
	router := gin.New()
	RegisterRoutes(router.Group("/panel/api"))
	return router
}

func performJSON(t *testing.T, router http.Handler, method, path string, value any) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if value != nil {
		if err := json.NewEncoder(&body).Encode(value); err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, path, &body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestLocalSponsorsMigrationCRUDRestartAndFeatureOff(t *testing.T) {
	db := sponsorsTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasTable(&Record{}) {
		t.Fatal("local sponsors table was not migrated")
	}
	saveSettingsForTest(t, db, ProviderLocal)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })

	router := managementRouter()
	created := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", sponsorRecordInput{
		ID:             "local-one",
		Enabled:        boolPointer(true),
		Name:           "Local One",
		Priority:       4,
		Slots:          []string{"sidebar", "page"},
		DestinationURL: "https://example.com/sponsor",
		Title:          map[string]string{"en-US": "Local One"},
		Text:           map[string]string{"en-US": "A local sponsor"},
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", created.Code, created.Body.String())
	}

	list, err := currentSponsors(time.Now().UTC())
	if err != nil || len(list.Sponsors) != 1 || list.Sponsors[0].ID != "local-one" {
		t.Fatalf("local list = %#v, err = %v", list, err)
	}

	updated := performJSON(t, router, http.MethodPut, "/panel/api/fork/sponsors/local-one", sponsorRecordInput{
		Enabled:        boolPointer(false),
		Name:           "Local One",
		Priority:       4,
		Slots:          []string{"page"},
		DestinationURL: "https://example.com/sponsor",
		Title:          map[string]string{"en-US": "Local One"},
	})
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updated.Code, updated.Body.String())
	}
	list, err = currentSponsors(time.Now().UTC())
	if err != nil || len(list.Sponsors) != 0 {
		t.Fatalf("disabled local list = %#v, err = %v", list, err)
	}

	Configure(nil, false)
	Configure(db, true)
	list, err = currentSponsors(time.Now().UTC())
	if err != nil || len(list.Sponsors) != 0 {
		t.Fatalf("restarted local list = %#v, err = %v", list, err)
	}
	Configure(db, false)
	response := performJSON(t, router, http.MethodGet, "/panel/api/fork/sponsors/manage", nil)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "featureDisabled") {
		t.Fatalf("feature-off response = %d/%s", response.Code, response.Body.String())
	}
	Configure(db, true)

	deleted := performJSON(t, router, http.MethodDelete, "/panel/api/fork/sponsors/local-one", nil)
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", deleted.Code, deleted.Body.String())
	}
	var count int64
	if err := db.Model(&Record{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("local record count = %d, err = %v", count, err)
	}
}

func TestRemoteProviderMakesLocalCRUDReadOnly(t *testing.T) {
	db := sponsorsTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	saveSettingsForTest(t, db, ProviderRemote)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })

	response := performJSON(t, managementRouter(), http.MethodPost, "/panel/api/fork/sponsors", sponsorRecordInput{
		Enabled:        boolPointer(true),
		Name:           "Remote blocked",
		Slots:          []string{"page"},
		DestinationURL: "https://example.com/",
		Title:          map[string]string{"en-US": "Remote blocked"},
	})
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "switch sponsors provider") {
		t.Fatalf("remote CRUD response = %d/%s", response.Code, response.Body.String())
	}
}

func TestSponsorManagementWritesMetadataOnlyAuditEvents(t *testing.T) {
	db := sponsorsTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := audit.Migrate(db); err != nil {
		t.Fatal(err)
	}
	audit.Configure(db, true)
	t.Cleanup(func() { audit.Configure(nil, false) })
	saveSettingsForTest(t, db, ProviderLocal)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })

	router := managementRouter()
	response := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", sponsorRecordInput{
		ID:             "audited",
		Enabled:        boolPointer(true),
		Name:           "Audited",
		Slots:          []string{"page"},
		DestinationURL: "https://example.com/",
		Title:          map[string]string{"en-US": "Audited"},
		Text:           map[string]string{"en-US": "Safe metadata"},
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", response.Code, response.Body.String())
	}
	response = performJSON(t, router, http.MethodPut, "/panel/api/fork/sponsors/settings", map[string]any{
		"providerMode": ProviderLocal,
		"sourceUrl":    "",
		"contactUrl":   "",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("settings status = %d, body = %s", response.Code, response.Body.String())
	}

	var events []audit.AuditEvent
	if err := db.Order("created_at ASC").Find(&events).Error; err != nil {
		t.Fatal(err)
	}
	if len(events) < 2 {
		t.Fatalf("audit events = %d, want create and provider config", len(events))
	}
	for _, event := range events {
		if event.Metadata != `{"source":"catx_sponsors"}` {
			t.Fatalf("audit metadata = %q, contains sponsor payload", event.Metadata)
		}
		if strings.Contains(event.Metadata, "example.com") || strings.Contains(event.Metadata, "Safe metadata") {
			t.Fatalf("audit metadata leaked sponsor content: %q", event.Metadata)
		}
	}
}

func TestSponsorsSchemaPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("XUI_TEST_PG_DSN"))
	if dsn == "" {
		t.Skip("set XUI_TEST_PG_DSN to a reachable Postgres to run this test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasTable(&Record{}) {
		t.Fatal("sponsors table was not migrated in PostgreSQL")
	}
	saveSettingsForTest(t, db, ProviderLocal)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })
	input := sponsorRecordInput{
		ID:             "postgres-duplicate",
		Enabled:        boolPointer(true),
		Name:           "Postgres duplicate",
		Slots:          []string{"page"},
		DestinationURL: "https://example.com/",
	}
	router := managementRouter()
	if response := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", input); response.Code != http.StatusCreated {
		t.Fatalf("first PostgreSQL create status = %d, body = %s", response.Code, response.Body.String())
	}
	if response := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", input); response.Code != http.StatusConflict {
		t.Fatalf("duplicate PostgreSQL create status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestDuplicateSponsorIDSQLiteReturnsConflict(t *testing.T) {
	db := sponsorsTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	saveSettingsForTest(t, db, ProviderLocal)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })
	input := sponsorRecordInput{
		ID:             "sqlite-duplicate",
		Enabled:        boolPointer(true),
		Name:           "SQLite duplicate",
		Slots:          []string{"page"},
		DestinationURL: "https://example.com/",
	}
	router := managementRouter()
	if response := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", input); response.Code != http.StatusCreated {
		t.Fatalf("first SQLite create status = %d, body = %s", response.Code, response.Body.String())
	}
	if response := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", input); response.Code != http.StatusConflict {
		t.Fatalf("duplicate SQLite create status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestSponsorPUTIsFullReplacementAndPathIDWins(t *testing.T) {
	db := sponsorsTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	saveSettingsForTest(t, db, ProviderLocal)
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })
	router := managementRouter()
	initial := sponsorRecordInput{
		ID:             "replace-me",
		Enabled:        boolPointer(true),
		Name:           "Before replacement",
		Priority:       5,
		Slots:          []string{"dashboard", "sidebar"},
		StartAt:        "2026-01-01T00:00:00Z",
		EndAt:          "2026-12-31T00:00:00Z",
		DestinationURL: "https://example.com/old",
		LogoURL:        "https://example.com/old.png",
		Title:          map[string]string{"en-US": "Old title"},
		Text:           map[string]string{"en-US": "Old text"},
	}
	if response := performJSON(t, router, http.MethodPost, "/panel/api/fork/sponsors", initial); response.Code != http.StatusCreated {
		t.Fatalf("initial create status = %d, body = %s", response.Code, response.Body.String())
	}
	replacement := sponsorRecordInput{
		ID:             "must-be-ignored",
		Enabled:        boolPointer(false),
		Name:           "After replacement",
		Priority:       0,
		Slots:          []string{"page"},
		DestinationURL: "https://example.com/new",
	}
	response := performJSON(t, router, http.MethodPut, "/panel/api/fork/sponsors/replace-me", replacement)
	if response.Code != http.StatusOK {
		t.Fatalf("replacement status = %d, body = %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Obj managedSponsor `json:"obj"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	got := envelope.Obj
	if got.ID != "replace-me" || got.Enabled || got.Name != "After replacement" || got.Priority != 0 || strings.Join(got.Slots, ",") != "page" {
		t.Fatalf("replacement retained identity/state: %+v", got)
	}
	if got.StartAt != "" || got.EndAt != "" || got.LogoURL != "" || got.DestinationURL != "https://example.com/new" {
		t.Fatalf("replacement retained cleared fields: %+v", got)
	}
	if got.Title["en-US"] != "After replacement" || len(got.Text) != 0 {
		t.Fatalf("replacement localized fields = title=%#v text=%#v", got.Title, got.Text)
	}
}

func TestLocalSponsorsCapSidebarSlots(t *testing.T) {
	db := sponsorsTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < maxSidebarSlots+1; index++ {
		record := Record{
			ID:             "sidebar-" + string(rune('a'+index)),
			Enabled:        true,
			Name:           "Sidebar",
			Priority:       index,
			SlotsJSON:      `["sidebar"]`,
			DestinationURL: "https://example.com/",
			TitleJSON:      `{"en-US":"Sidebar"}`,
			TextJSON:       `{}`,
		}
		if err := db.Create(&record).Error; err != nil {
			t.Fatal(err)
		}
	}
	list, err := localSponsorList(db, "", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Sponsors) != maxSidebarSlots {
		t.Fatalf("local sidebar sponsors = %d, want %d", len(list.Sponsors), maxSidebarSlots)
	}
}

func TestNormalizeRecordInputRejectsUnsafeOrUnsupportedFields(t *testing.T) {
	base := sponsorRecordInput{
		Enabled:        boolPointer(true),
		Name:           "Valid sponsor",
		Slots:          []string{"page", "dashboard"},
		DestinationURL: "https://example.com/",
		Title:          map[string]string{"en-US": "Valid sponsor"},
	}
	if record, err := normalizeRecordInput(base, nil); err != nil || record.ID == "" {
		t.Fatalf("valid sponsor = %#v, err = %v", record, err)
	}
	for name, input := range map[string]sponsorRecordInput{
		"missing enabled":  {Name: "x", Slots: []string{"page"}, DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "x"}},
		"missing name":     {Enabled: boolPointer(true), Slots: []string{"page"}, DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "x"}},
		"empty slots":      {Enabled: boolPointer(true), Name: "x", DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "x"}},
		"login slot":       {Enabled: boolPointer(true), Name: "x", Slots: []string{"login"}, DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "x"}},
		"http destination": {Enabled: boolPointer(true), Name: "x", Slots: []string{"page"}, DestinationURL: "http://example.com/", Title: map[string]string{"en-US": "x"}},
		"credential URL":   {Enabled: boolPointer(true), Name: "x", Slots: []string{"page"}, DestinationURL: "https://u:p@example.com/", Title: map[string]string{"en-US": "x"}},
		"HTML title":       {Enabled: boolPointer(true), Name: "x", Slots: []string{"page"}, DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "<script>"}},
		"bad schedule":     {Enabled: boolPointer(true), Name: "x", Slots: []string{"page"}, StartAt: "2026-01-02T00:00:00Z", EndAt: "2026-01-01T00:00:00Z", DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "x"}},
		"invalid time":     {Enabled: boolPointer(true), Name: "x", Slots: []string{"page"}, StartAt: "not-a-time", DestinationURL: "https://example.com/", Title: map[string]string{"en-US": "x"}},
	} {
		if _, err := normalizeRecordInput(input, nil); err == nil {
			t.Errorf("%s unexpectedly succeeded", name)
		}
	}
}

func TestNormalizeSlotsAreDeterministicAndOpenEndedDatesAreJSONSafe(t *testing.T) {
	record, err := normalizeRecordInput(sponsorRecordInput{
		Enabled:        boolPointer(true),
		Name:           "Ordered",
		Slots:          []string{"page", "sidebar", "dashboard", "sidebar"},
		DestinationURL: "https://example.com/",
		Title:          map[string]string{"en-US": "Ordered"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	view, err := toPublicSponsor(*record, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(view.Slots, ",") != "dashboard,sidebar,page" {
		t.Fatalf("normalized slots = %#v", view.Slots)
	}
	if _, err := json.Marshal(view); err != nil {
		t.Fatalf("open-ended sponsor is not JSON-safe: %v", err)
	}
}

func saveSettingsForTest(t *testing.T, db *gorm.DB, mode ProviderMode) {
	t.Helper()
	if err := saveSettings(db, settings{providerMode: mode}); err != nil {
		t.Fatal(err)
	}
}

func boolPointer(value bool) *bool { return &value }
