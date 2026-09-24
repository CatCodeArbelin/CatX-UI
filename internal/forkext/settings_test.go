package forkext

import (
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSettingsDefaultOffAndPersisted(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:forkext-settings?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
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
	var rows []model.Setting
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("read settings: %v", err)
	}
	if len(rows) != 1 || rows[0].Key != "fork.policies.enabled" {
		t.Fatalf("unexpected settings rows: %+v", rows)
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
