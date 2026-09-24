// Package forkext contains the fixed, first-party extension boundaries used by
// CatX-UI. The initial contracts are deliberately no-ops so the upstream
// execution path remains authoritative until a later work package supplies
// feature implementations.
package forkext

import (
	"context"
	"log"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Install initializes the fixed first-party hooks. It intentionally has no
// side effects while all fork features are disabled.
func Install() {}

// RegisterMigrations is the single database integration point for fork-owned
// migrations. Analytics is opt-in; the default path is an exact no-op.
func RegisterMigrations(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	enabled, err := NewSettings(db).Enabled(FlagAnalytics)
	if err != nil {
		log.Printf("fork analytics disabled: cannot read feature flag: %v", err)
		return nil
	}
	if !enabled {
		return nil
	}
	if err := analytics.Migrate(db); err != nil {
		log.Printf("fork analytics migration skipped after error: %v", err)
	}
	return nil
}

// RegisterRoutes is the protected API integration point for fork endpoints.
// An empty registration preserves the upstream route set exactly.
func RegisterRoutes(_ *gin.RouterGroup) {}

// RegisterJobs is the fixed scheduler integration point for fork jobs.
func RegisterJobs(_ context.Context, _ *cron.Cron) {}

// RegisterEventSubscribers is the fixed event-bus integration point for fork
// subscribers.
func RegisterEventSubscribers(_ *eventbus.Bus) {}

// Start is the lifecycle integration point for fork-owned goroutines. The
// returned closer is always safe to call in the no-op foundation state.
func Start(_ context.Context) (func(), error) { return func() {}, nil }

// Stop is the matching lifecycle shutdown hook.
func Stop() {}

// DecorateXrayConfig is called after upstream has assembled the complete
// candidate configuration. Returning the same pointer is the exact no-op
// behavior required while fork features are disabled.
func DecorateXrayConfig(_ context.Context, cfg *xray.Config) (*xray.Config, error) {
	return cfg, nil
}
