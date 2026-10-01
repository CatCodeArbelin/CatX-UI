package forkext

import (
	"context"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/fleetupdate"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/groupquota"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/portal"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/risk"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficcontrol"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficpolicy"
	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRuntimeTestDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	return db
}

func assertRuntimeState(t *testing.T, enabled bool) {
	t.Helper()
	status := analytics.CurrentStatus()
	if status.Enabled != enabled || status.DNSIntelligence != enabled {
		t.Fatalf("analytics status = %+v, want enabled=%v and dns=%v", status, enabled, enabled)
	}
	checks := map[string]bool{
		"audit":            audit.Enabled(),
		"fleet updates":    fleetupdate.Current() != nil && fleetupdate.Current().Enabled(),
		"fleet mutation":   fleetupdate.Current() != nil && fleetupdate.Current().MutationEnabled(),
		"group quota":      groupquota.Enabled(),
		"policies":         policy.Enabled(),
		"self service":     portal.Enabled(),
		"security anomaly": risk.Enabled(),
		"traffic control":  trafficcontrol.Enabled(),
		"traffic policy":   trafficpolicy.Enabled(),
	}
	for name, actual := range checks {
		if actual != enabled {
			t.Errorf("%s enabled = %v, want %v", name, actual, enabled)
		}
	}
}

func allManagedFeatures(enabled bool) map[Flag]bool {
	result := make(map[Flag]bool, len(managedFeatureFlags))
	for _, feature := range managedFeatureFlags {
		result[feature.flag] = enabled
	}
	return result
}

func TestRuntimeReloadAppliesManagedFeatureTransitionsAndSchema(t *testing.T) {
	disableRuntime()
	t.Cleanup(disableRuntime)
	db := newRuntimeTestDB(t, "forkext-runtime-transitions")

	if err := RegisterMigrations(db); err != nil {
		t.Fatalf("prepare disabled schema: %v", err)
	}
	assertRuntimeState(t, false)
	if err := ConfigureRuntimeFromSettings(db); err != nil {
		t.Fatalf("configure disabled runtime: %v", err)
	}
	if db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("analytics schema created while disabled")
	}
	if db.Migrator().HasTable(&model.User{}) || db.Migrator().HasTable(&model.HistoryOfSeeders{}) {
		t.Fatal("fork lifecycle hook ran upstream database initialization or seeders")
	}

	if err := NewSettings(db).UpdateFeatures(allManagedFeatures(true)); err != nil {
		t.Fatalf("persist enabled feature set: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("reload enabled runtime: %v", err)
	}
	assertRuntimeState(t, true)
	if !db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("newly enabled analytics schema is missing")
	}
	for name, target := range map[string]any{
		"audit":            &audit.AuditEvent{},
		"fleet updates":    &fleetupdate.Campaign{},
		"group quota":      &groupquota.State{},
		"policies":         &policy.Policy{},
		"security anomaly": &risk.IPHistory{},
		"self service":     &portal.Credential{},
		"traffic policy":   &trafficpolicy.Policy{},
	} {
		if !db.Migrator().HasTable(target) {
			t.Errorf("newly enabled %s schema is missing", name)
		}
	}
	if db.Migrator().HasTable(&model.User{}) || db.Migrator().HasTable(&model.HistoryOfSeeders{}) {
		t.Fatal("runtime reload ran upstream database initialization or seeders")
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("idempotent enabled reload: %v", err)
	}
	assertRuntimeState(t, true)

	if err := NewSettings(db).UpdateFeatures(allManagedFeatures(false)); err != nil {
		t.Fatalf("persist disabled feature set: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("reload disabled runtime: %v", err)
	}
	assertRuntimeState(t, false)
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("idempotent disabled reload: %v", err)
	}
	assertRuntimeState(t, false)

	if err := NewSettings(db).UpdateFeatures(allManagedFeatures(true)); err != nil {
		t.Fatalf("persist re-enabled feature set: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("reload re-enabled runtime: %v", err)
	}
	assertRuntimeState(t, true)
}

func TestRuntimeReloadFailureDoesNotLeaveFeaturesEnabled(t *testing.T) {
	disableRuntime()
	t.Cleanup(disableRuntime)
	db := newRuntimeTestDB(t, "forkext-runtime-failure")
	if err := NewSettings(db).UpdateFeatures(allManagedFeatures(true)); err != nil {
		t.Fatalf("persist enabled feature set: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("initial reload: %v", err)
	}
	assertRuntimeState(t, true)

	if err := db.Model(&model.Setting{}).Where("key = ?", settingKey(FlagAnalytics)).Update("value", "not-a-boolean").Error; err != nil {
		t.Fatalf("corrupt feature setting: %v", err)
	}
	err := ReloadRuntimeFromSettings(db)
	if err == nil || !strings.Contains(err.Error(), "read fork feature flag analytics.enabled") {
		t.Fatalf("reload error = %v, want feature-read failure", err)
	}
	assertRuntimeState(t, false)

	if err := db.Model(&model.Setting{}).Where("key = ?", settingKey(FlagAnalytics)).Update("value", "false").Error; err != nil {
		t.Fatalf("restore analytics feature setting: %v", err)
	}
	err = ReloadRuntimeFromSettings(db)
	if err == nil || !strings.Contains(err.Error(), "dns_intelligence.enabled requires analytics.enabled") {
		t.Fatalf("reload error = %v, want dependency failure", err)
	}
	assertRuntimeState(t, false)
}

func TestDisabledHooksAreExactNoOps(t *testing.T) {
	disableRuntime()
	t.Cleanup(disableRuntime)
	if err := RegisterMigrations((*gorm.DB)(nil)); err != nil {
		t.Fatalf("RegisterMigrations() error = %v", err)
	}
	if err := ConfigureRuntimeFromSettings((*gorm.DB)(nil)); err != nil {
		t.Fatalf("ConfigureRuntimeFromSettings() error = %v", err)
	}
	RegisterRoutes((*gin.RouterGroup)(nil))
	RegisterJobs(context.Background(), (*cron.Cron)(nil))
	RegisterEventSubscribers((*eventbus.Bus)(nil))
	closer, err := Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	closer()
	Stop()

	cfg := &xray.Config{}
	decorated, err := DecorateXrayConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("DecorateXrayConfig() error = %v", err)
	}
	if decorated != cfg {
		t.Fatal("DecorateXrayConfig() replaced the upstream config in no-op mode")
	}
}
