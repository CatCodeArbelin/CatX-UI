package forkext

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// M00 qualification tests intentionally use one process-global lifecycle at a
// time. The production snapshot and settings DB are process-wide by design;
// resetting both around every test prevents a later test from inheriting an
// earlier runtime state.
func beginM00Test(t *testing.T) {
	t.Helper()
	disableRuntime()
	setSettingsDB(nil)
	t.Cleanup(func() {
		setSettingsDB(nil)
		disableRuntime()
	})
}

func newM00SQLite(t *testing.T, name string) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), name+".db")
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "requires cgo") {
			t.Skip("BLOCKED: local SQLite acceptance requires a CGO-enabled Go toolchain")
		}
		t.Fatalf("open SQLite database: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func startM00Runtime(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := RegisterMigrations(db); err != nil {
		t.Fatalf("register M00 migrations: %v", err)
	}
	if err := ConfigureRuntimeFromSettings(db); err != nil {
		t.Fatalf("configure M00 runtime: %v", err)
	}
}

func m00FeatureItem(t *testing.T, db *gorm.DB, flag Flag) FeatureFlagInfo {
	t.Helper()
	items, err := NewSettings(db).FeatureFlags()
	if err != nil {
		t.Fatalf("read feature flags: %v", err)
	}
	for _, item := range items {
		if item.Key == flag {
			return item
		}
	}
	t.Fatalf("feature flag %q is missing from Feature Settings", flag)
	return FeatureFlagInfo{}
}

func assertM00Feature(t *testing.T, db *gorm.DB, flag Flag, enabled, active bool, state RuntimeState, restartRequired bool) {
	t.Helper()
	item := m00FeatureItem(t, db, flag)
	if item.Enabled != enabled || item.Active != active || item.State != state || item.RestartRequired != restartRequired {
		t.Fatalf("feature %q = enabled:%v active:%v state:%q restart:%v, want enabled:%v active:%v state:%q restart:%v", flag, item.Enabled, item.Active, item.State, item.RestartRequired, enabled, active, state, restartRequired)
	}
}

func m00FeatureSettingsGET(t *testing.T, db *gorm.DB) featureSettingsEnvelope {
	t.Helper()
	setSettingsDB(db)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerFeatureSettingsRoutes(router.Group("/panel/api"))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panel/api/fork/settings/features", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("Feature Settings GET status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
	return decodeFeatureSettings(t, recorder)
}

// M00-T01 — all CatX features OFF is an effective runtime no-op.
func TestM00T01AllFeaturesOff(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-test-off")
	startM00Runtime(t, db)

	assertRuntimeState(t, false)
	for _, feature := range managedFeatureFlags {
		assertM00Feature(t, db, feature.flag, false, false, RuntimeStateFeatureOff, false)
	}
	response := m00FeatureSettingsGET(t, db)
	if !response.Success || response.Obj.RestartRequired {
		t.Fatalf("feature-off settings response = %+v, want readable/no pending state", response)
	}
}

// M00-T02 — persisting desired ON does not activate the runtime.
func TestM00T02DesiredOnIsNotActive(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-desired-on")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}

	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateRestartRequired, true)
}

// M00-T03 — explicit preparation followed by apply is the only successful
// activation path asserted by this module.
func TestM00T03SuccessfulActivation(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-activation")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare analytics runtime: %v", err)
	}
	if !db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("explicit preparation did not create analytics schema")
	}
	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateRestartRequired, true)

	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("activate prepared analytics runtime: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, true, true, RuntimeStateActive, false)
	if status := analytics.CurrentStatus(); !status.Enabled {
		t.Fatalf("analytics runtime status = %+v, want enabled", status)
	}
}

// M00-T04 — disable remains pending until the real apply boundary, then
// removes the feature runtime effect.
func TestM00T04SuccessfulDisable(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-disable")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("enable analytics desired state: %v", err)
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare analytics: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("activate analytics: %v", err)
	}

	if err := NewSettings(db).Set(FlagAnalytics, false); err != nil {
		t.Fatalf("disable analytics desired state: %v", err)
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare analytics disable: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, false, true, RuntimeStateRestartRequired, true)

	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("apply analytics disable: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, false, false, RuntimeStateFeatureOff, false)
	if status := analytics.CurrentStatus(); status.Enabled || status.DNSIntelligence {
		t.Fatalf("analytics runtime after disable = %+v, want disabled", status)
	}
}

// M00-T06 — closing and reopening the persisted database reconstructs desired
// state through the same startup registration/configuration boundaries.
func TestM00T06ProcessRestartReconstructsRuntime(t *testing.T) {
	beginM00Test(t)
	path := filepath.Join(t.TempDir(), "catx-m00-test.db")
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "requires cgo") {
			t.Skip("BLOCKED: local SQLite acceptance requires a CGO-enabled Go toolchain")
		}
		t.Fatalf("open initial SQLite database: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate initial settings: %v", err)
	}
	if err := RegisterMigrations(db); err != nil {
		t.Fatalf("register initial M00 migrations: %v", err)
	}
	if err := ConfigureRuntimeFromSettings(db); err != nil {
		t.Fatalf("configure initial M00 runtime: %v", err)
	}
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics: %v", err)
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare analytics before restart: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("activate analytics before restart: %v", err)
	}
	firstSQL, err := db.DB()
	if err != nil {
		t.Fatalf("get initial SQLite pool: %v", err)
	}
	if err := firstSQL.Close(); err != nil {
		t.Fatalf("close initial SQLite pool: %v", err)
	}
	disableRuntime()

	reopened, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatalf("reopen SQLite database: %v", err)
	}
	reopenedSQL, err := reopened.DB()
	if err != nil {
		t.Fatalf("get reopened SQLite pool: %v", err)
	}
	t.Cleanup(func() { _ = reopenedSQL.Close() })
	if err := RegisterMigrations(reopened); err != nil {
		t.Fatalf("register M00 migrations after restart: %v", err)
	}
	if err := ConfigureRuntimeFromSettings(reopened); err != nil {
		t.Fatalf("configure M00 runtime after restart: %v", err)
	}
	assertM00Feature(t, reopened, FlagAnalytics, true, true, RuntimeStateActive, false)
}

// M00-T07 — generic reload does not become an implicit migration boundary.
func TestM00T07GenericReloadDoesNotMigrate(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-reload-no-migrate")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	if db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("analytics schema unexpectedly exists before explicit preparation")
	}
	if err := ReloadRuntimeFromSettings(db); err == nil || !strings.Contains(err.Error(), "schema is not prepared") {
		t.Fatalf("generic reload error = %v, want explicit-preparation error", err)
	}
	if db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("generic reload created analytics schema")
	}
	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateError, false)
}

// M00-T08 — the explicit preparation boundary creates the required schema.
func TestM00T08ExplicitPreparation(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-explicit-preparation")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("explicit preparation: %v", err)
	}
	if !db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("explicit preparation did not create analytics schema")
	}
	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateRestartRequired, true)
}

// M00-T09 — an inactive feature that cannot prepare is never reported active,
// and the settings surface remains readable.
func TestM00T09PreparationFailure(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-preparation-failure")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	original := migrateAnalyticsSchema
	migrateAnalyticsSchema = func(*gorm.DB) error { return errors.New("M00 injected analytics preparation failure") }
	t.Cleanup(func() { migrateAnalyticsSchema = original })
	if err := PrepareRuntimeFromSettings(db); err == nil || !strings.Contains(err.Error(), "M00 injected analytics preparation failure") {
		t.Fatalf("preparation error = %v, want deterministic injected failure", err)
	}
	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateError, false)
	response := m00FeatureSettingsGET(t, db)
	if !response.Success {
		t.Fatalf("Feature Settings after preparation failure = %+v, want readable response", response)
	}
}

// M00-T10 — ordered preparation records partial DDL instead of pretending it
// rolled back the successful first step.
func TestM00T10PartialPreparationFailure(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-partial-preparation")
	startM00Runtime(t, db)
	if err := NewSettings(db).UpdateFeatures(map[Flag]bool{FlagAnalytics: true, FlagPolicies: true}); err != nil {
		t.Fatalf("persist ordered feature state: %v", err)
	}
	original := migratePolicySchema
	migratePolicySchema = func(*gorm.DB) error { return errors.New("M00 injected policy preparation failure") }
	t.Cleanup(func() { migratePolicySchema = original })
	if err := PrepareRuntimeFromSettings(db); err == nil || !strings.Contains(err.Error(), "M00 injected policy preparation failure") {
		t.Fatalf("partial preparation error = %v, want deterministic injected failure", err)
	}
	if !db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("successful first preparation step was rolled back unexpectedly")
	}
	if db.Migrator().HasTable(&policy.Policy{}) {
		t.Fatal("failed second preparation step unexpectedly created policy schema")
	}
	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateError, false)
	assertM00Feature(t, db, FlagPolicies, true, false, RuntimeStateError, false)
}

// M00-T11 — removing the deterministic fault makes the same activation
// retryable through the approved preparation and apply boundaries.
func TestM00T11RetryAfterFailure(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-retry")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	original := migrateAnalyticsSchema
	migrateAnalyticsSchema = func(*gorm.DB) error { return errors.New("M00 one-shot analytics failure") }
	t.Cleanup(func() { migrateAnalyticsSchema = original })
	if err := PrepareRuntimeFromSettings(db); err == nil {
		t.Fatal("first preparation unexpectedly succeeded")
	}
	migrateAnalyticsSchema = original
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("retry preparation: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("retry activation: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, true, true, RuntimeStateActive, false)
}

// M00-T12 — a failed later transition preserves the previous known-good
// runtime while distinguishing the desired failed transition with ERROR.
func TestM00T12PreviousKnownGoodRuntime(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-known-good")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare known-good analytics: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("activate known-good analytics: %v", err)
	}
	if err := NewSettings(db).Set(FlagPolicies, true); err != nil {
		t.Fatalf("persist policy transition: %v", err)
	}
	original := migratePolicySchema
	migratePolicySchema = func(*gorm.DB) error { return errors.New("M00 injected later-transition failure") }
	t.Cleanup(func() { migratePolicySchema = original })
	if err := PrepareRuntimeFromSettings(db); err == nil {
		t.Fatal("later preparation unexpectedly succeeded")
	}
	analyticsItem := m00FeatureItem(t, db, FlagAnalytics)
	if !analyticsItem.Active || analyticsItem.State != RuntimeStateError {
		t.Fatalf("known-good analytics after failed later transition = %+v, want active/error", analyticsItem)
	}
	if status := analytics.CurrentStatus(); !status.Enabled {
		t.Fatalf("known-good analytics runtime = %+v, want preserved enabled runtime", status)
	}
	assertM00Feature(t, db, FlagPolicies, true, false, RuntimeStateError, false)
}

// M00-T13 — malformed persisted input is isolated to its feature item.
func TestM00T13MalformedPersistedFlag(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-malformed")
	if err := db.Create(&model.Setting{Key: settingKey(FlagAnalytics), Value: "not-a-boolean"}).Error; err != nil {
		t.Fatalf("persist malformed analytics flag: %v", err)
	}
	response := m00FeatureSettingsGET(t, db)
	if !response.Success || len(response.Obj.Items) != len(managedFeatureFlags) {
		t.Fatalf("malformed Feature Settings response = %+v, want complete semantic response", response)
	}
	assertM00Feature(t, db, FlagAnalytics, false, false, RuntimeStateError, false)
}

// M00-T14 — dependency validation is enforced at the backend persistence
// boundary, not only by a UI.
func TestM00T14DependencyViolation(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-dependency")
	settings := NewSettings(db)
	if err := settings.UpdateFeatures(map[Flag]bool{FlagDNSIntelligence: true}); err == nil || !strings.Contains(err.Error(), "requires") {
		t.Fatalf("dependency violation error = %v, want rejection", err)
	}
	for _, flag := range []Flag{FlagAnalytics, FlagDNSIntelligence, FlagSecurityAnomaly} {
		if enabled, err := settings.Enabled(flag); err != nil || enabled {
			t.Fatalf("dependency rejection changed %q: enabled=%v err=%v", flag, enabled, err)
		}
	}
}

// M00-T15 — the complete deterministic preparation/failure/retry contract is
// exercised against a file-backed SQLite database.
func TestM00T15SQLiteLifecycle(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-test")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist SQLite analytics state: %v", err)
	}
	original := migrateAnalyticsSchema
	migrateAnalyticsSchema = func(*gorm.DB) error { return errors.New("M00 SQLite injected failure") }
	t.Cleanup(func() { migrateAnalyticsSchema = original })
	if err := PrepareRuntimeFromSettings(db); err == nil {
		t.Fatal("SQLite failure injection unexpectedly succeeded")
	}
	migrateAnalyticsSchema = original
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("SQLite retry preparation: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("SQLite retry activation: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, true, true, RuntimeStateActive, false)
}

// M00-T17 — the Feature Settings API exposes desired, active, pending, and
// final state as distinct semantic fields.
func TestM00T17FeatureSettingsAPISemantics(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-api-semantics")
	startM00Runtime(t, db)
	setSettingsDB(db)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerFeatureSettingsRoutes(router.Group("/panel/api"))

	put := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", strings.NewReader(`{"flags":{"analytics.enabled":true}}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(put, request)
	if put.Code != http.StatusOK {
		t.Fatalf("enable PUT status = %d; body=%s", put.Code, put.Body.String())
	}
	enabled := decodeFeatureSettings(t, put)
	if !enabled.Success || !enabled.Obj.RestartRequired {
		t.Fatalf("enable PUT envelope = %+v, want success with pending runtime", enabled)
	}
	assertM00Feature(t, db, FlagAnalytics, true, false, RuntimeStateRestartRequired, true)

	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("apply API-enabled feature: %v", err)
	}
	active := m00FeatureSettingsGET(t, db)
	if !active.Success {
		t.Fatalf("active Feature Settings response = %+v", active)
	}
	assertM00Feature(t, db, FlagAnalytics, true, true, RuntimeStateActive, false)

	put = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", strings.NewReader(`{"flags":{"analytics.enabled":false}}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(put, request)
	if put.Code != http.StatusOK {
		t.Fatalf("disable PUT status = %d; body=%s", put.Code, put.Body.String())
	}
	disabledPending := decodeFeatureSettings(t, put)
	if !disabledPending.Success || !disabledPending.Obj.RestartRequired {
		t.Fatalf("disable PUT envelope = %+v, want success with pending runtime", disabledPending)
	}
	assertM00Feature(t, db, FlagAnalytics, false, true, RuntimeStateRestartRequired, true)

	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("apply API-disabled feature: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, false, false, RuntimeStateFeatureOff, false)
}

// M00-T18 — successful activation is idempotent and does not duplicate the
// settings record or drift the runtime snapshot.
func TestM00T18IdempotentActivation(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-idempotent-on")
	startM00Runtime(t, db)
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := PrepareRuntimeFromSettings(db); err != nil {
			t.Fatalf("prepare iteration %d: %v", i, err)
		}
		if err := ReloadRuntimeFromSettings(db); err != nil {
			t.Fatalf("activate iteration %d: %v", i, err)
		}
	}
	var settingsRows int64
	if err := db.Model(&model.Setting{}).Where("key = ?", settingKey(FlagAnalytics)).Count(&settingsRows).Error; err != nil {
		t.Fatalf("count analytics settings rows: %v", err)
	}
	if settingsRows != 1 {
		t.Fatalf("analytics settings rows = %d, want exactly one", settingsRows)
	}
	var observations int64
	if err := db.Model(&analytics.DestinationObservation{}).Count(&observations).Error; err != nil {
		t.Fatalf("count analytics observations after repeated activation: %v", err)
	}
	assertM00Feature(t, db, FlagAnalytics, true, true, RuntimeStateActive, false)
}

// M00-T19 — repeated disable is a stable no-op.
func TestM00T19IdempotentDisable(t *testing.T) {
	beginM00Test(t)
	db := newM00SQLite(t, "catx-m00-idempotent-off")
	startM00Runtime(t, db)
	for i := 0; i < 3; i++ {
		if err := PrepareRuntimeFromSettings(db); err != nil {
			t.Fatalf("disabled preparation iteration %d: %v", i, err)
		}
		if err := ReloadRuntimeFromSettings(db); err != nil {
			t.Fatalf("disabled apply iteration %d: %v", i, err)
		}
	}
	assertRuntimeState(t, false)
	assertM00Feature(t, db, FlagAnalytics, false, false, RuntimeStateFeatureOff, false)
}

// M00-T20 — disabled fork hooks preserve the upstream object and are safe to
// invoke through the fixed lifecycle boundaries.
func TestM00T20FeatureOffUpstreamCompatibility(t *testing.T) {
	beginM00Test(t)
	if err := RegisterMigrations((*gorm.DB)(nil)); err != nil {
		t.Fatalf("RegisterMigrations(nil): %v", err)
	}
	if err := ConfigureRuntimeFromSettings((*gorm.DB)(nil)); err != nil {
		t.Fatalf("ConfigureRuntimeFromSettings(nil): %v", err)
	}
	RegisterRoutes((*gin.RouterGroup)(nil))
	RegisterJobs(context.Background(), (*cron.Cron)(nil))
	RegisterEventSubscribers((*eventbus.Bus)(nil))
	closer, err := Start(context.Background())
	if err != nil {
		t.Fatalf("Start(): %v", err)
	}
	closer()
	Stop()

	candidate := &xray.Config{}
	decorated, err := DecorateXrayConfig(context.Background(), candidate)
	if err != nil {
		t.Fatalf("DecorateXrayConfig(): %v", err)
	}
	if decorated != candidate {
		t.Fatal("feature-off Xray decoration replaced the upstream config")
	}
	assertRuntimeState(t, false)
}
