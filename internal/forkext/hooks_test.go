package forkext

import (
	"context"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

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
