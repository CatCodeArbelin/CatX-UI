package forkext

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRuntimeActivationPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("XUI_TEST_PG_DSN"))
	if dsn == "" {
		t.Skip("set XUI_TEST_PG_DSN to a disposable Postgres database to run this test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	disableRuntime()
	t.Cleanup(disableRuntime)
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	if err := NewSettings(db).UpdateFeatures(allManagedFeatures(false)); err != nil {
		t.Fatalf("reset test feature flags: %v", err)
	}
	if err := RegisterMigrations(db); err != nil {
		t.Fatalf("register migrations: %v", err)
	}
	if err := ConfigureRuntimeFromSettings(db); err != nil {
		t.Fatalf("configure disabled runtime: %v", err)
	}
	if err := NewSettings(db).UpdateFeatures(allManagedFeatures(true)); err != nil {
		t.Fatalf("enable feature flags: %v", err)
	}
	if err := ReloadRuntimeFromSettings(db); err == nil {
		t.Fatal("postgres reload activated an unprepared schema")
	}
	if err := PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare postgres runtime: %v", err)
	}
	for name, target := range map[string]any{
		"analytics": &analytics.DestinationObservation{},
		"policies":  &policy.Policy{},
	} {
		if !db.Migrator().HasTable(target) {
			t.Errorf("postgres preparation did not create %s schema", name)
		}
	}
	if err := ReloadRuntimeFromSettings(db); err != nil {
		t.Fatalf("activate postgres runtime: %v", err)
	}
	assertRuntimeState(t, true)

	failures := []struct {
		name string
		seam *func(*gorm.DB) error
	}{
		{name: "audit", seam: &migrateAuditSchema},
		{name: "analytics", seam: &migrateAnalyticsSchema},
		{name: "policy", seam: &migratePolicySchema},
	}
	for _, failure := range failures {
		t.Run(failure.name, func(t *testing.T) {
			original := *failure.seam
			defer func() { *failure.seam = original }()
			*failure.seam = func(*gorm.DB) error { return errors.New("injected postgres " + failure.name + " DDL failure") }

			err := PrepareRuntimeFromSettings(db)
			if err == nil || !strings.Contains(err.Error(), "injected postgres "+failure.name+" DDL failure") {
				t.Fatalf("preparation error = %v, want injected %s failure", err, failure.name)
			}
			assertRuntimeState(t, true)
			*failure.seam = original
			if err := PrepareRuntimeFromSettings(db); err != nil {
				t.Fatalf("retry preparation after %s failure: %v", failure.name, err)
			}
			if err := ReloadRuntimeFromSettings(db); err != nil {
				t.Fatalf("retry activation after %s failure: %v", failure.name, err)
			}
			assertRuntimeState(t, true)
		})
	}
}
