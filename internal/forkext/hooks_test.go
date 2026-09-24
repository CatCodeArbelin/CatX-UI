package forkext

import (
	"context"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAnalyticsMigrationFollowsFeatureFlag(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:forkext-analytics-migration?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	if err := RegisterMigrations(db); err != nil {
		t.Fatalf("disabled migration: %v", err)
	}
	if db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("analytics schema created while disabled")
	}
	if err := NewSettings(db).Set(FlagAnalytics, true); err != nil {
		t.Fatalf("enable analytics: %v", err)
	}
	if err := RegisterMigrations(db); err != nil {
		t.Fatalf("enabled migration: %v", err)
	}
	if !db.Migrator().HasTable(&analytics.DestinationObservation{}) {
		t.Fatal("analytics schema missing while enabled")
	}
}

func TestDisabledHooksAreExactNoOps(t *testing.T) {
	if err := RegisterMigrations((*gorm.DB)(nil)); err != nil {
		t.Fatalf("RegisterMigrations() error = %v", err)
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
