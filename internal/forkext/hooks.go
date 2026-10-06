// Package forkext contains the fixed, first-party extension boundaries used by
// CatX-UI. Feature modules stay behind these hooks so upstream execution paths
// remain authoritative when a fork feature is disabled.
package forkext

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/fleetupdate"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/groupquota"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/portal"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/risk"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/sponsors"
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

type runtimeConfig struct {
	analytics      bool
	dns            bool
	policies       bool
	trafficControl bool
	security       bool
	audit          bool
	selfService    bool
	fleetUpdates   bool
	fleetMutation  bool
	sponsors       bool
}

func loadRuntimeConfig(db *gorm.DB) (runtimeConfig, error) {
	if db == nil {
		return runtimeConfig{}, nil
	}
	values := make(map[Flag]bool, len(managedFeatureFlags))
	settings := NewSettings(db)
	for _, feature := range managedFeatureFlags {
		enabled, err := settings.Enabled(feature.flag)
		if err != nil {
			return runtimeConfig{}, fmt.Errorf("read fork feature flag %s: %w", feature.flag, err)
		}
		values[feature.flag] = enabled
	}
	if err := validateFeatureState(values); err != nil {
		return runtimeConfig{}, fmt.Errorf("validate fork feature flags: %w", err)
	}
	return runtimeConfig{
		analytics:      values[FlagAnalytics],
		dns:            values[FlagDNSIntelligence],
		policies:       values[FlagPolicies],
		trafficControl: values[FlagTrafficControl],
		security:       values[FlagSecurityAnomaly],
		audit:          values[FlagAudit],
		selfService:    values[FlagSelfService],
		fleetUpdates:   values[FlagFleetUpdates],
		fleetMutation:  values[FlagFleetMutation],
		sponsors:       values[FlagSponsors],
	}, nil
}

func prepareRuntimeSchema(db *gorm.DB, cfg runtimeConfig) error {
	if db == nil {
		return nil
	}
	// The portal schema has historically been part of every CatX database.
	// Preserve that compatibility while keeping all other optional schemas
	// gated by the feature that owns them.
	if err := portal.Migrate(db); err != nil {
		return fmt.Errorf("prepare self-service schema: %w", err)
	}
	if cfg.fleetUpdates {
		if err := fleetupdate.Migrate(db); err != nil {
			return fmt.Errorf("prepare fleet-update schema: %w", err)
		}
	}
	if cfg.audit {
		if err := audit.Migrate(db); err != nil {
			return fmt.Errorf("prepare audit schema: %w", err)
		}
	}
	if cfg.trafficControl {
		if err := groupquota.Migrate(db); err != nil {
			return fmt.Errorf("prepare group-quota schema: %w", err)
		}
		if err := trafficpolicy.Migrate(db); err != nil {
			return fmt.Errorf("prepare traffic-policy schema: %w", err)
		}
	}
	if cfg.analytics {
		if err := analytics.Migrate(db); err != nil {
			return fmt.Errorf("prepare analytics schema: %w", err)
		}
	}
	if cfg.policies {
		if err := policy.Migrate(db); err != nil {
			return fmt.Errorf("prepare policy schema: %w", err)
		}
	}
	if cfg.security {
		if err := risk.Migrate(db); err != nil {
			return fmt.Errorf("prepare security-anomaly schema: %w", err)
		}
	}
	return nil
}

func disableRuntime() {
	disableRuntimeWithError("")
}

func disableRuntimeWithError(runtimeErr string) {
	portal.Configure(nil, false)
	fleetupdate.Configure(nil, false)
	audit.Configure(nil, false)
	groupquota.Configure(nil, false)
	trafficpolicy.Configure(nil, false)
	trafficcontrol.Configure(false)
	analytics.Configure(analytics.NoopRepository{}, false)
	analytics.SetEvidenceEnabled(false)
	policy.Configure(nil, false)
	risk.Configure(nil, false)
	sponsors.Configure(nil, false)
	publishRuntimeSnapshot(runtimeConfig{}, true, runtimeErr)
}

func applyRuntimeConfig(db *gorm.DB, cfg runtimeConfig) {
	portal.Configure(db, cfg.selfService)
	fleetupdate.Configure(db, cfg.fleetUpdates)
	if service := fleetupdate.Current(); service != nil {
		service.SetMutationEnabled(cfg.fleetMutation)
	}
	audit.Configure(db, cfg.audit)
	groupquota.Configure(db, cfg.trafficControl)
	trafficpolicy.Configure(db, cfg.trafficControl)
	trafficcontrol.Configure(cfg.trafficControl)
	if cfg.analytics {
		analytics.Configure(analytics.NewRepository(db, true), true)
		analytics.SetRetentionPolicy(retentionPolicyFromDB(db))
	} else {
		analytics.Configure(analytics.NoopRepository{}, false)
	}
	analytics.SetEvidenceEnabled(cfg.analytics && cfg.dns)
	policy.Configure(db, cfg.policies)
	risk.Configure(db, cfg.security)
	sponsors.Configure(db, cfg.sponsors)
	publishRuntimeSnapshot(cfg, true, "")
}

// RegisterMigrations is the single database integration point for initial
// fork-owned schema preparation. It deliberately does not configure runtime
// state; database startup performs that as a separate lifecycle step.
func RegisterMigrations(db *gorm.DB) error {
	setSettingsDB(db)
	if db != nil {
		if err := migrateSponsorsSchema(db); err != nil {
			disableRuntimeWithError(err.Error())
			return fmt.Errorf("migrate sponsors schema: %w", err)
		}
	}
	cfg, err := loadRuntimeConfig(db)
	if err != nil {
		disableRuntimeWithError(err.Error())
		return err
	}
	if err := prepareRuntimeSchema(db, cfg); err != nil {
		disableRuntimeWithError(err.Error())
		return err
	}
	return nil
}

var migrateSponsorsSchema = sponsors.Migrate

// ConfigureRuntimeFromSettings configures in-process fork services only. The
// caller must prepare schema first; startup owns the initial migration pass.
func ConfigureRuntimeFromSettings(db *gorm.DB) error {
	cfg, err := loadRuntimeConfig(db)
	if err != nil {
		disableRuntimeWithError(err.Error())
		return err
	}
	applyRuntimeConfig(db, cfg)
	return nil
}

// ReloadRuntimeFromSettings is the panel-restart boundary. It prepares the
// schemas for the newly persisted fork configuration through the same
// idempotent migration boundary used at startup, then publishes in-process
// state. This is required when an operator enables a module on an already
// running panel: a process restart is not guaranteed to follow this call.
func ReloadRuntimeFromSettings(db *gorm.DB) error {
	setSettingsDB(db)
	cfg, err := loadRuntimeConfig(db)
	if err != nil {
		disableRuntimeWithError(err.Error())
		return err
	}
	if err := prepareRuntimeSchema(db, cfg); err != nil {
		disableRuntimeWithError(err.Error())
		return err
	}
	applyRuntimeConfig(db, cfg)
	return nil
}

// RegisterRoutes is the protected API integration point for fork endpoints.
// The controller calls it only after establishing the authenticated panel API
// group, so modules do not create public admin surfaces accidentally.
func RegisterRoutes(api *gin.RouterGroup) {
	portal.RegisterAdminRoutes(api)
	audit.RegisterRoutes(api)
	analytics.RegisterActivityRoutes(api)
	registerAnalyticsSettingsRoutes(api)
	registerFeatureSettingsRoutes(api)
	policy.RegisterRoutes(api)
	policysim.RegisterRoutes(api)
	groupquota.RegisterRoutes(api)
	trafficcontrol.RegisterRoutes(api)
	trafficpolicy.RegisterRoutes(api)
	fleetupdate.RegisterRoutes(api)
	risk.RegisterRoutes(api)
	sponsors.RegisterRoutes(api)
}

// RegisterPortalRoutes installs the separate client portal surface at the
// application root without sharing the admin session namespace.
func RegisterPortalRoutes(g *gin.RouterGroup, secret []byte, basePath string, secure bool) {
	portal.RegisterPortalRoutes(g, secret, basePath, secure)
}

// SetDeviceRevoker wires the portal to the existing ownership-safe client
// device service without importing web/service into the fork package.
func SetDeviceRevoker(fn func(email string, deviceID int) error) {
	portal.SetDeviceRevoker(fn)
}

// RegisterMiddleware installs request correlation and post-success auditing
// before upstream routes are registered.
func RegisterMiddleware(api *gin.RouterGroup) {
	if api != nil {
		api.Use(audit.Middleware())
	}
}

// RegisterJobs is the fixed scheduler integration point for fork jobs.
func RegisterJobs(_ context.Context, scheduler *cron.Cron) {
	audit.RegisterJobs(scheduler)
	groupquota.RegisterJobs(scheduler)
	trafficcontrol.RegisterJobs(scheduler)
	trafficpolicy.RegisterJobs(scheduler)
	risk.RegisterJobs(scheduler)
	fleetupdate.RegisterJobs(scheduler)
}

func SetTrafficControlRestartCallback(fn func()) { groupquota.SetRestartCallback(fn) }

// RecordClientIPHistory is the single fork handoff used by the existing
// CheckClientIpJob and node attribution merge path. It never collects data on
// its own and is a no-op when security analytics is disabled.
func RecordClientIPHistory(ctx context.Context, nodeGuid string, observations map[string][]model.ClientIpEntry, source string) error {
	return risk.RecordIPHistory(ctx, nodeGuid, observations, source)
}

func RecordLogin(ctx context.Context, username, sourceIP, outcome, reason string, userID int) error {
	return audit.RecordLogin(ctx, username, sourceIP, outcome, reason, userID)
}

// RegisterEventSubscribers is the fixed event-bus integration point for fork
// subscribers.
func RegisterEventSubscribers(bus *eventbus.Bus) { audit.RegisterEventSubscribers(bus) }

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
	return append(append(append(append(groupquota.Models(), trafficpolicy.Models()...), &portal.Credential{}, &portal.HostGrant{}), &fleetupdate.Campaign{}, &fleetupdate.Target{}), sponsors.Models()...)
}

// Start is the lifecycle integration point for fork-owned goroutines. The
// returned closer is always safe to call in the no-op foundation state.
func Start(ctx context.Context) (func(), error) {
	closer, err := analytics.Start(ctx)
	if err != nil {
		return nil, err
	}
	auditCloser, err := audit.Start(ctx)
	if err != nil {
		closer()
		return nil, err
	}
	return func() { auditCloser(); closer() }, nil
}

// Stop is the matching lifecycle shutdown hook.
func Stop() { audit.Stop() }

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
