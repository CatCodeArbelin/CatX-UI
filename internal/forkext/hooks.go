// Package forkext contains the fixed, first-party extension boundaries used by
// CatX-UI. The initial contracts are deliberately no-ops so the upstream
// execution path remains authoritative until a later work package supplies
// feature implementations.
package forkext

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/groupquota"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficcontrol"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficpolicy"
	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/policycompiler"
	"github.com/mhsanaei/3x-ui/v3/internal/policysim"
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
	setSettingsDB(db)
	if db == nil {
		groupquota.Configure(nil, false)
		trafficpolicy.Configure(nil, false)
		trafficcontrol.Configure(false)
		analytics.Configure(analytics.NoopRepository{}, false)
		policy.Configure(nil, false)
		return nil
	}
	trafficControlEnabled, err := NewSettings(db).Enabled(FlagTrafficControl)
	if err != nil {
		trafficControlEnabled = false
	}
	if trafficControlEnabled {
		if err := groupquota.Migrate(db); err != nil {
			return err
		}
		if err := trafficpolicy.Migrate(db); err != nil {
			return err
		}
	}
	groupquota.Configure(db, trafficControlEnabled)
	trafficpolicy.Configure(db, trafficControlEnabled)
	trafficcontrol.Configure(trafficControlEnabled)
	analyticsEnabled, err := NewSettings(db).Enabled(FlagAnalytics)
	if err != nil {
		log.Printf("fork analytics disabled: cannot read feature flag: %v", err)
		analyticsEnabled = false
	}
	if !analyticsEnabled {
		analytics.Configure(analytics.NoopRepository{}, false)
	} else if err := analytics.Migrate(db); err != nil {
		log.Printf("fork analytics migration skipped after error: %v", err)
		analytics.Configure(analytics.NoopRepository{}, false)
		analyticsEnabled = false
	} else {
		analytics.Configure(analytics.NewRepository(db, true), true)
		analytics.SetRetentionPolicy(retentionPolicyFromDB(db))
	}
	policiesEnabled, err := NewSettings(db).Enabled(FlagPolicies)
	if err != nil {
		log.Printf("fork policies disabled: cannot read feature flag: %v", err)
		policiesEnabled = false
	}
	if policiesEnabled {
		if err := policy.Migrate(db); err != nil {
			log.Printf("fork policy migration skipped after error: %v", err)
			policiesEnabled = false
		}
	}
	policy.Configure(db, policiesEnabled)
	dnsEnabled, err := NewSettings(db).Enabled(FlagDNSIntelligence)
	if err != nil {
		log.Printf("fork DNS intelligence disabled: cannot read feature flag: %v", err)
		dnsEnabled = false
	}
	analytics.SetEvidenceEnabled(analyticsEnabled && dnsEnabled)
	return nil
}

// RegisterRoutes is the protected API integration point for fork endpoints.
// An empty registration preserves the upstream route set exactly.
func RegisterRoutes(api *gin.RouterGroup) {
	analytics.RegisterActivityRoutes(api)
	registerAnalyticsSettingsRoutes(api)
	policy.RegisterRoutes(api)
	policysim.RegisterRoutes(api)
	groupquota.RegisterRoutes(api)
	trafficcontrol.RegisterRoutes(api)
	trafficpolicy.RegisterRoutes(api)
}

// RegisterJobs is the fixed scheduler integration point for fork jobs.
func RegisterJobs(_ context.Context, scheduler *cron.Cron) {
	groupquota.RegisterJobs(scheduler)
	trafficcontrol.RegisterJobs(scheduler)
	trafficpolicy.RegisterJobs(scheduler)
}

func SetTrafficControlRestartCallback(fn func()) { groupquota.SetRestartCallback(fn) }

// RegisterEventSubscribers is the fixed event-bus integration point for fork
// subscribers.
func RegisterEventSubscribers(_ *eventbus.Bus) {}

// GroupQuotaApplyDeltas is the sole accounting handoff from upstream traffic
// writes to the fork quota module.
func ApplyTrafficDeltas(tx *gorm.DB, traffics []*xray.ClientTraffic) error {
	if err := groupquota.ApplyClientDeltas(tx, traffics); err != nil {
		return err
	}
	return trafficpolicy.ApplyDeltas(tx, traffics)
}

func GroupQuotaApplyDeltas(tx *gorm.DB, traffics []*xray.ClientTraffic) error {
	return ApplyTrafficDeltas(tx, traffics)
}

func GroupQuotaDepletedEmails(tx *gorm.DB) ([]string, error) {
	return groupquota.DepletedEmails(tx)
}

func GroupQuotaPreserveReset(tx *gorm.DB, email string, newUp, newDown int64) error {
	return groupquota.PreserveReset(tx, email, newUp, newDown)
}

func GroupQuotaClearOwnership(tx *gorm.DB, email string) error {
	return groupquota.ClearOwnership(tx, email)
}

func GroupQuotaIsBlocked(tx *gorm.DB, email string) (bool, error) {
	return groupquota.IsBlocked(tx, email)
}

func GroupQuotaIsGroupDepleted(tx *gorm.DB, group string) (bool, error) {
	return groupquota.IsGroupDepleted(tx, group)
}

func GroupQuotaReset(tx *gorm.DB, group string) ([]string, error) {
	return groupquota.Reset(tx, group)
}

func GroupQuotaChangeMembership(tx *gorm.DB, email, oldGroup, newGroup string) error {
	return groupquota.ChangeMembership(tx, email, oldGroup, newGroup)
}

func GroupQuotaRemoveMembership(tx *gorm.DB, email string) error {
	return groupquota.RemoveMembershipByEmail(tx, email)
}

func GroupQuotaViews(tx *gorm.DB) ([]groupquota.GroupView, error) {
	return groupquota.ListViews(tx)
}

func GroupQuotaRename(tx *gorm.DB, oldName, newName string) error {
	return groupquota.RenameGroup(tx, oldName, newName)
}
func GroupQuotaReconcileEnabled(tx *gorm.DB) error { return groupquota.ReconcileEnabled(tx) }
func GroupQuotaRebaselineClient(tx *gorm.DB, email string, up, down int64) error {
	return groupquota.RebaselineClient(tx, email, up, down)
}

func MigrationModels() []any {
	return append(groupquota.Models(), trafficpolicy.Models()...)
}

// Start is the lifecycle integration point for fork-owned goroutines. The
// returned closer is always safe to call in the no-op foundation state.
func Start(ctx context.Context) (func(), error) { return analytics.Start(ctx) }

// Stop is the matching lifecycle shutdown hook.
func Stop() {}

// DecorateXrayConfig is called after upstream has assembled the complete
// candidate configuration. Returning the same pointer is the exact no-op
// behavior required while fork features are disabled.
func DecorateXrayConfig(ctx context.Context, cfg *xray.Config) (*xray.Config, error) {
	if cfg == nil || !policy.Enabled() {
		return cfg, nil
	}
	set := map[string]struct{}{}
	for _, inbound := range cfg.InboundConfigs {
		var settings map[string]any
		if len(inbound.Settings) == 0 {
			continue
		}
		if err := json.Unmarshal(inbound.Settings, &settings); err != nil {
			return cfg, err
		}
		clients, ok := settings["clients"].([]any)
		if !ok {
			continue
		}
		for _, raw := range clients {
			if client, ok := raw.(map[string]any); ok {
				if email, ok := client["email"].(string); ok && email != "" {
					set[email] = struct{}{}
				}
			}
		}
	}
	emails := make([]string, 0, len(set))
	for email := range set {
		emails = append(emails, email)
	}
	decisions, err := policy.ResolveCurrent(ctx, emails, time.Now().UnixMilli())
	if err != nil {
		return cfg, err
	}
	if len(decisions) == 0 {
		return cfg, nil
	}
	return policycompiler.Compile(cfg, decisions)
}
