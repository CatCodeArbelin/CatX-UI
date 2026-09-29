package forkext

import (
	"path/filepath"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return openSettingsTestDB(t, filepath.Join(t.TempDir(), "settings.db"))
}

func openSettingsTestDB(t *testing.T, path string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	return db
}

func TestSettingsDefaultOffAndPersisted(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "settings.db")
	db := openSettingsTestDB(t, dbPath)
	if err := db.Create(&model.Setting{Key: "panel.unrelated", Value: "preserve-me"}).Error; err != nil {
		t.Fatalf("seed upstream setting: %v", err)
	}

	settings := NewSettings(db)
	flags, err := settings.Flags()
	if err != nil {
		t.Fatalf("Flags() error = %v", err)
	}
	for flag, enabled := range flags {
		if enabled {
			t.Errorf("default %q = true, want false", flag)
		}
	}
	if err := settings.Set(FlagPolicies, true); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	enabled, err := NewSettings(db).Enabled(FlagPolicies)
	if err != nil {
		t.Fatalf("Enabled() error = %v", err)
	}
	if !enabled {
		t.Fatal("persisted policies flag = false, want true")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite database: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sqlite database: %v", err)
	}
	reopened, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("reopen sqlite: %v", err)
	}
	var rows []model.Setting
	if err := reopened.Find(&rows).Error; err != nil {
		t.Fatalf("read settings: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("unexpected settings rows: %+v", rows)
	}
	var reloaded model.Setting
	if err := reopened.Where("key = ?", "fork.policies.enabled").First(&reloaded).Error; err != nil {
		t.Fatalf("reload persisted fork setting: %v", err)
	}
	if reloaded.Value != "true" {
		t.Fatalf("reloaded policies value = %q, want true", reloaded.Value)
	}
	var unrelated model.Setting
	if err := reopened.Where("key = ?", "panel.unrelated").First(&unrelated).Error; err != nil {
		t.Fatalf("reload upstream setting: %v", err)
	}
	if unrelated.Value != "preserve-me" {
		t.Fatalf("unrelated upstream value = %q, want preserve-me", unrelated.Value)
	}
}

func TestSettingsUnknownFlagRejected(t *testing.T) {
	settings := NewSettings(nil)
	if _, err := settings.Enabled(Flag("unknown.enabled")); err == nil {
		t.Fatal("Enabled() accepted an unknown flag")
	}
	if err := settings.Set(Flag("unknown.enabled"), true); err == nil {
		t.Fatal("Set() accepted an unknown flag")
	}
}

func TestUpdateFeaturesValidatesDependenciesAndPreservesState(t *testing.T) {
	db := newSettingsTestDB(t)
	settings := NewSettings(db)

	if err := settings.UpdateFeatures(map[Flag]bool{FlagDNSIntelligence: true}); err == nil {
		t.Fatal("DNS intelligence enabled without analytics dependency")
	}
	if enabled, err := settings.Enabled(FlagDNSIntelligence); err != nil || enabled {
		t.Fatalf("invalid dependency update changed DNS intelligence: enabled=%v err=%v", enabled, err)
	}

	if err := settings.UpdateFeatures(map[Flag]bool{
		FlagAnalytics:       true,
		FlagDNSIntelligence: true,
		FlagSecurityAnomaly: true,
	}); err != nil {
		t.Fatalf("valid analytics dependency update: %v", err)
	}
	if err := settings.UpdateFeatures(map[Flag]bool{FlagAnalytics: false}); err == nil {
		t.Fatal("analytics disabled while dependent features remained enabled")
	}
	for _, flag := range []Flag{FlagAnalytics, FlagDNSIntelligence, FlagSecurityAnomaly} {
		if enabled, err := settings.Enabled(flag); err != nil || !enabled {
			t.Fatalf("failed dependency update changed %q: enabled=%v err=%v", flag, enabled, err)
		}
	}

	if err := settings.UpdateFeatures(map[Flag]bool{
		FlagDNSIntelligence: false,
		FlagSecurityAnomaly: false,
		FlagAnalytics:       false,
	}); err != nil {
		t.Fatalf("valid dependency disable update: %v", err)
	}

	if err := settings.UpdateFeatures(map[Flag]bool{FlagFleetMutation: true}); err == nil {
		t.Fatal("fleet mutation enabled without fleet updates dependency")
	}
	if err := settings.UpdateFeatures(map[Flag]bool{
		FlagFleetUpdates:  true,
		FlagFleetMutation: true,
	}); err != nil {
		t.Fatalf("valid fleet dependency update: %v", err)
	}
}

func TestUpdateFeaturesRollsBackOnPersistenceFailure(t *testing.T) {
	db := newSettingsTestDB(t)
	if err := db.Exec(`
		CREATE TRIGGER reject_policies_flag
		BEFORE INSERT ON settings
		WHEN NEW.key = 'fork.policies.enabled'
		BEGIN
			SELECT RAISE(ABORT, 'test write failure');
		END`).Error; err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	settings := NewSettings(db)
	if err := settings.UpdateFeatures(map[Flag]bool{
		FlagAnalytics: true,
		FlagPolicies:  true,
	}); err == nil {
		t.Fatal("UpdateFeatures succeeded despite a persistence failure")
	}
	for _, flag := range []Flag{FlagAnalytics, FlagPolicies} {
		if enabled, err := settings.Enabled(flag); err != nil || enabled {
			t.Fatalf("failed transaction left %q enabled=%v err=%v", flag, enabled, err)
		}
	}
}
