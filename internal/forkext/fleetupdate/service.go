package fleetupdate

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
	"github.com/mhsanaei/3x-ui/v3/internal/forkrelease"
)

var (
	ErrDisabled         = errors.New("fleet updates are disabled")
	ErrCampaignNotFound = gorm.ErrRecordNotFound
)

type ReleaseResolver interface {
	Resolve(context.Context, string) (ReleaseSnapshot, error)
}
type ProviderResolver struct{ Provider forkrelease.Provider }

func (r ProviderResolver) Resolve(ctx context.Context, channel string) (ReleaseSnapshot, error) {
	var ch forkrelease.Channel
	switch channel {
	case string(forkrelease.ChannelStable):
		ch = forkrelease.ChannelStable
	case string(forkrelease.ChannelDev):
		ch = forkrelease.ChannelDev
	default:
		return ReleaseSnapshot{}, fmt.Errorf("unsupported release channel %q", channel)
	}
	release, err := r.Provider.Fetch(ctx, ch)
	if err != nil {
		return ReleaseSnapshot{}, err
	}
	if _, err := r.Provider.DownloadVerified(ctx, release, forkrelease.Current.UpdateScriptAssetName(), 16<<20); err != nil {
		return ReleaseSnapshot{}, fmt.Errorf("release checksum preflight: %w", err)
	}
	// The trusted update script is staged and verified by the node-local updater;
	// the campaign persists only the release identity and this trust result.
	return ReleaseSnapshot{Channel: channel, Tag: release.TagName, APIURL: release.APIURL, HTMLURL: release.HTMLURL, ChecksumVerified: true}, nil
}

type Service struct {
	db              *gorm.DB
	enabled         bool
	resolver        ReleaseResolver
	executor        UpdateExecutor
	mutationEnabled bool
	mu              sync.Mutex
}

var configured struct {
	sync.RWMutex
	service *Service
}

func Configure(db *gorm.DB, enabled bool) {
	configured.Lock()
	configured.service = New(db, enabled)
	configured.Unlock()
}

func Current() *Service { configured.RLock(); defer configured.RUnlock(); return configured.service }

func New(db *gorm.DB, enabled bool) *Service {
	return &Service{db: db, enabled: enabled, resolver: ProviderResolver{Provider: forkrelease.NewProvider(nil)}, executor: DisabledExecutor{}}
}

func (s *Service) SetResolver(r ReleaseResolver) {
	if r != nil {
		s.resolver = r
	}
}

func (s *Service) SetExecutor(e UpdateExecutor) {
	if e != nil {
		s.executor = e
	}
}

func (s *Service) SetMutationEnabled(enabled bool) { s.mutationEnabled = enabled }

func SetExecutor(e UpdateExecutor) {
	if s := Current(); s != nil && e != nil {
		s.SetExecutor(e)
	}
}
func (s *Service) Enabled() bool { return s != nil && s.enabled && s.db != nil }
func Migrate(db *gorm.DB) error  { return db.AutoMigrate(&Campaign{}, &Target{}) }

func normalizeChannel(c string) string {
	if c == string(forkrelease.ChannelDev) {
		return c
	}
	return string(forkrelease.ChannelStable)
}

func defaults(r *PlanRequest) {
	if r.Name == "" {
		r.Name = "Fleet update"
	}
	r.Channel = normalizeChannel(r.Channel)
	if r.BatchSize <= 0 {
		r.BatchSize = 1
	}
	if r.MaxParallel <= 0 {
		r.MaxParallel = 1
	}
	if r.HealthTimeoutSecs <= 0 {
		r.HealthTimeoutSecs = 300
	}
	if r.SoakSeconds < 0 {
		r.SoakSeconds = 0
	}
	if r.StopOnFailure == nil {
		v := true
		r.StopOnFailure = &v
	}
}

func nodeSnapshot(n model.Node) NodeSnapshot {
	return NodeSnapshot{ID: n.Id, Guid: n.Guid, Name: n.Name, Address: n.Address, Port: n.Port, Scheme: n.Scheme, Enabled: n.Enable, Status: n.Status, LastHeartbeat: n.LastHeartbeat, PanelVersion: n.PanelVersion, XrayState: n.XrayState}
}

func fresh(n NodeSnapshot) bool {
	return n.LastHeartbeat > 0 && time.Since(time.Unix(n.LastHeartbeat, 0)) <= 30*time.Second
}
func normalizedVersion(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }
func sameVersion(a, b string) bool {
	return normalizedVersion(a) != "" && normalizedVersion(a) == normalizedVersion(b)
}

func eligibility(n NodeSnapshot, release ReleaseSnapshot) string {
	if n.ID <= 0 || n.Transitive {
		return ReasonTransitive
	}
	if !n.Enabled {
		return ReasonDisabled
	}
	if n.Status == "offline" {
		return ReasonOffline
	}
	if !fresh(n) {
		return ReasonStaleHeartbeat
	}
	if n.PanelVersion == "" {
		return ReasonUnsupported
	}
	if n.XrayState != "" && n.XrayState != "running" && n.XrayState != "healthy" && n.XrayState != "ok" {
		return ReasonDegraded
	}
	if sameVersion(n.PanelVersion, release.Tag) {
		return ReasonAlreadyCurrent
	}
	return ""
}

func (s *Service) Plan(ctx context.Context, req PlanRequest) (Plan, error) {
	if !s.Enabled() {
		return Plan{}, ErrDisabled
	}
	defaults(&req)
	if !req.DryRun && (!s.mutationEnabled || !req.ConfirmProduction) {
		_ = audit.Record(ctx, &audit.AuditEvent{EventType: "fleet_update.authorization", Outcome: "denied", ActorType: "user", TargetType: "update_campaign", Metadata: "{\"reason\":\"mutation_gate\"}"})
		return Plan{}, ErrMutationDisabled
	}
	if !req.DryRun {
		_ = audit.Record(ctx, &audit.AuditEvent{EventType: "fleet_update.authorization", Outcome: "success", ActorType: "user", TargetType: "update_campaign", Metadata: "{\"confirmed\":true}"})
	}
	release, err := s.resolver.Resolve(ctx, req.Channel)
	if err != nil {
		return Plan{}, err
	}
	q := s.db.WithContext(ctx).Model(&model.Node{}).Where("id > 0")
	if len(req.NodeIDs) > 0 {
		q = q.Where("id IN ?", req.NodeIDs)
	}
	var nodes []model.Node
	if err := q.Order("id asc").Find(&nodes).Error; err != nil {
		return Plan{}, err
	}
	c := Campaign{Name: req.Name, Channel: req.Channel, ReleaseTag: release.Tag, ReleaseAPIURL: release.APIURL, ReleaseHTMLURL: release.HTMLURL, DryRun: req.DryRun, MutationAuthorized: !req.DryRun && s.mutationEnabled && req.ConfirmProduction, CanaryCount: req.CanaryCount, BatchSize: req.BatchSize, MaxParallel: req.MaxParallel, HealthTimeoutSecs: req.HealthTimeoutSecs, SoakSeconds: req.SoakSeconds, StopOnFailure: *req.StopOnFailure, State: StatePreflight, Revision: 1}
	if err := s.db.WithContext(ctx).Create(&c).Error; err != nil {
		return Plan{}, err
	}
	ids := uniqueInts(req.NodeIDs)
	present := map[int]bool{}
	targets := make([]Target, 0, len(nodes)+len(ids))
	for _, n := range nodes {
		present[n.Id] = true
		snap := nodeSnapshot(n)
		reason := eligibility(snap, release)
		if s.concurrentNode(n.Id) {
			reason = ReasonConcurrentCampaign
		}
		state := StateReady
		if reason != "" {
			state = StateBlocked
		}
		targets = append(targets, Target{CampaignID: c.ID, NodeID: n.Id, NodeGuid: n.Guid, NodeName: n.Name, NodeAddress: n.Address, NodePort: n.Port, NodeScheme: n.Scheme, InitialVersion: n.PanelVersion, State: state, BlockedReason: reason, ObservedVersion: n.PanelVersion, ObservedStatus: n.Status, ObservedXray: n.XrayState, DispatchKey: fmt.Sprintf("%d:%d", c.ID, n.Id)})
	}
	for _, id := range ids {
		if !present[id] {
			targets = append(targets, Target{CampaignID: c.ID, NodeID: id, State: StateBlocked, BlockedReason: ReasonDeleted, DispatchKey: fmt.Sprintf("%d:%d", c.ID, id)})
		}
	}
	for i := range targets {
		if err := s.db.WithContext(ctx).Create(&targets[i]).Error; err != nil {
			return Plan{}, err
		}
	}
	if req.DryRun {
		c.State = StateReady
	} else {
		c.State = StateReady
	}
	s.db.Model(&c).Updates(map[string]any{"state": c.State})
	auditCampaign(ctx, "fleet_update.plan", &c, "success")
	return Plan{Campaign: c, Targets: targets}, nil
}

func uniqueInts(in []int) []int {
	seen := map[int]bool{}
	out := make([]int, 0, len(in))
	for _, v := range in {
		if v > 0 && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

func (s *Service) concurrentNode(id int) bool {
	var n int64
	s.db.Model(&Target{}).Where("node_id = ? AND state NOT IN ?", id, []string{StateSucceeded, StateFailed, StateAborted, StateBlocked}).Count(&n)
	return n > 0
}

func (s *Service) Get(ctx context.Context, id uint) (Plan, error) {
	if !s.Enabled() {
		return Plan{}, ErrDisabled
	}
	var c Campaign
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return Plan{}, err
	}
	var ts []Target
	if err := s.db.WithContext(ctx).Where("campaign_id = ?", id).Order("id asc").Find(&ts).Error; err != nil {
		return Plan{}, err
	}
	return Plan{Campaign: c, Targets: ts}, nil
}

func (s *Service) List(ctx context.Context) ([]Campaign, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	var c []Campaign
	err := s.db.WithContext(ctx).Order("id desc").Find(&c).Error
	return c, err
}

func terminal(state string) bool {
	return state == StateSucceeded || state == StateFailed || state == StateAborted || state == StateBlocked
}

func (s *Service) acquireTarget(t *Target, owner string) bool {
	until := time.Now().Add(30 * time.Second)
	result := s.db.Model(&Target{}).Where("id = ? AND (lease_until IS NULL OR lease_until < ? OR lease_owner = ?)", t.ID, time.Now(), owner).Updates(map[string]any{"lease_owner": owner, "lease_until": until})
	return result.Error == nil && result.RowsAffected == 1
}

func (s *Service) Reconcile(ctx context.Context, id uint, owner string) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if owner == "" {
		owner = "fleet-worker"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var c Campaign
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return err
	}
	if c.State == StateAborted || c.State == StateSucceeded || c.State == StateFailed {
		return nil
	}
	if c.DryRun || !c.MutationAuthorized {
		c.State = StateReady
		c.Revision++
		return s.db.WithContext(ctx).Save(&c).Error
	}
	var ts []Target
	if err := s.db.WithContext(ctx).Where("campaign_id = ?", id).Order("id asc").Find(&ts).Error; err != nil {
		return err
	}
	for i := range ts {
		t := &ts[i]
		if terminal(t.State) {
			continue
		}
		if t.State == StateReady && !s.dispatchAllowed(&c, ts, i) {
			continue
		}
		leaseWasExpired := leaseExpired(t.LeaseUntil)
		if !s.acquireTarget(t, owner) {
			continue
		}
		if leaseWasExpired && activeTarget(t.State) {
			t.State = StateUnknown
		}
		var n model.Node
		err := s.db.WithContext(ctx).First(&n, t.NodeID).Error
		if err != nil {
			t.State = StateBlocked
			t.BlockedReason = ReasonDeleted
			_ = s.saveTarget(t)
			continue
		}
		ns := nodeSnapshot(n)
		ns.ExpectedRunID = t.RunID
		rel := ReleaseSnapshot{Channel: c.Channel, Tag: c.ReleaseTag, APIURL: c.ReleaseAPIURL, HTMLURL: c.ReleaseHTMLURL, ChecksumVerified: true}
		if t.State == StateReady || t.State == StatePreflight || t.State == StatePending {
			if reason := eligibility(ns, rel); reason != "" {
				t.State, t.BlockedReason = StateBlocked, reason
				_ = s.saveTarget(t)
				continue
			}
			pf, e := s.executor.Preflight(ctx, ns, rel)
			if e != nil {
				t.State = StateBlocked
				t.Error = e.Error()
			} else if !pf.Supported {
				t.State = StateBlocked
				t.BlockedReason = pf.Reason
			} else {
				t.State = StateDispatching
				t.Attempts++
				dr, e := s.executor.Dispatch(ctx, ns, rel)
				if e != nil {
					if errors.Is(e, ErrMutationDisabled) {
						t.State = StateBlocked
						t.BlockedReason = ReasonExecutionDisabled
						t.Error = e.Error()
					} else if errors.Is(e, ErrAmbiguousDispatch) {
						t.State = StateUnknown
						t.DispatchStatus = "ambiguous"
						t.DispatchEvidence = dr.Evidence
						t.Error = e.Error()
					} else {
						t.State = StateFailed
						t.Error = e.Error()
					}
				} else if dr.Ambiguous {
					t.State = StateUnknown
					t.DispatchStatus = "ambiguous"
					t.DispatchEvidence = dr.Evidence
					t.Error = ErrAmbiguousDispatch.Error()
				} else if strings.TrimSpace(dr.RunID) == "" {
					// A successful POST without a durable correlation key is not
					// evidence that this campaign owns the node's update. Preserve
					// the ambiguous state and require reconciliation before retry.
					t.State = StateUnknown
					t.DispatchStatus = "ambiguous"
					t.DispatchEvidence = dr.Evidence
					t.Error = ErrMissingRunID.Error()
				} else {
					now := time.Now()
					t.DispatchAt = &now
					t.DispatchEvidence = dr.Evidence
					t.RunID = dr.RunID
					t.DispatchStatus = "accepted"
					t.UpdateState = "pending"
					t.State = StateWaitingRestart
					auditTarget(ctx, "fleet_update.dispatch", &c, t, "accepted")
				}
			}
			_ = s.saveTarget(t)
			continue
		}
		if t.State == StateUnknown || t.State == StateDispatching || t.State == StateWaitingRestart || t.State == StateHealthCheck || t.State == StateSoaking {
			conv, e := s.executor.Reconcile(ctx, ns, rel)
			if e != nil {
				t.Error = e.Error()
				t.State = StateUnknown
			} else {
				t.ObservedVersion = conv.PanelVersion
				t.ObservedStatus = conv.NodeStatus
				t.ObservedXray = conv.XrayState
				t.UpdateState = conv.UpdateState
				if conv.ExitCode != 0 {
					code := conv.ExitCode
					t.UpdateExitCode = &code
				}
				t.UpdateFinishedAt = conv.FinishedAt
				t.RolledBack = conv.RolledBack
				t.RollbackHealthy = conv.RollbackHealthy
				if conv.UpdateState == "failed" {
					t.State = StateFailed
					t.DispatchStatus = "completed"
					t.Error = "node update failed"
					auditTarget(ctx, "fleet_update.completion", &c, t, "failed")
					if conv.RolledBack {
						auditTarget(ctx, "fleet_update.rollback_evidence", &c, t, map[bool]string{true: "healthy", false: "attempted"}[conv.RollbackHealthy])
					}
				} else if healthyConvergence(conv, rel) {
					if t.State != StateSoaking {
						t.State = StateSoaking
						started := time.Now()
						t.SoakStartedAt = &started
					}
					if c.SoakSeconds == 0 || t.SoakStartedAt != nil && time.Since(*t.SoakStartedAt) >= time.Duration(c.SoakSeconds)*time.Second {
						t.State = StateSucceeded
						auditTarget(ctx, "fleet_update.completion", &c, t, "success")
					}
				} else {
					t.State = StateHealthCheck
					if t.DispatchAt != nil && time.Since(*t.DispatchAt) >= time.Duration(c.HealthTimeoutSecs)*time.Second {
						t.State = StateFailed
						t.BlockedReason = ReasonHealthTimeout
					}
				}
			}
			_ = s.saveTarget(t)
		}
	}
	var remaining, failed, blocked int64
	s.db.Model(&Target{}).Where("campaign_id = ? AND state NOT IN ?", id, []string{StateSucceeded, StateFailed, StateAborted, StateBlocked}).Count(&remaining)
	s.db.Model(&Target{}).Where("campaign_id = ? AND state = ?", id, StateFailed).Count(&failed)
	s.db.Model(&Target{}).Where("campaign_id = ? AND state = ?", id, StateBlocked).Count(&blocked)
	if failed > 0 && c.StopOnFailure {
		c.State = StateFailed
		c.Error = "stopped on target failure"
	} else if remaining == 0 && blocked > 0 {
		c.State = StateBlocked
	} else if remaining == 0 {
		c.State = StateSucceeded
	} else {
		c.State = StateReady
	}
	c.LeaseOwner = ""
	c.LeaseUntil = time.Time{}
	c.Revision++
	s.db.WithContext(ctx).Save(&c)
	return nil
}

func activeTarget(state string) bool {
	return state == StateDispatching || state == StateWaitingRestart || state == StateHealthCheck || state == StateSoaking
}

// dispatchAllowed applies a deterministic canary, batch, and parallelism
// window. Target rows are read in ascending ID order, which is the immutable
// planning order and therefore stable across process restarts. Translation
// resources are validated alongside this orchestration contract in CI.
func (s *Service) dispatchAllowed(c *Campaign, ts []Target, index int) bool {
	if c.StopOnFailure {
		for i := range ts {
			if ts[i].State == StateFailed {
				return false
			}
		}
	}
	active, succeeded, readyOrdinal := 0, 0, 0
	for i := range ts {
		if activeTarget(ts[i].State) {
			active++
		}
		if ts[i].State == StateSucceeded {
			succeeded++
		}
		if ts[i].State == StateReady && i <= index {
			readyOrdinal++
		}
	}
	if active >= c.MaxParallel {
		return false
	}
	window := c.BatchSize
	if succeeded == 0 && c.CanaryCount > 0 {
		window = c.CanaryCount
		// Hold the campaign at the canary window until every canary target
		// converges. Active dispatches count toward that window even though
		// their rows are no longer in StateReady.
		if active >= c.CanaryCount {
			return false
		}
	}
	if window <= 0 {
		window = 1
	}
	return readyOrdinal <= window
}

func (s *Service) saveTarget(t *Target) error {
	t.LeaseOwner = ""
	t.LeaseUntil = time.Time{}
	return s.db.Save(t).Error
}

func (s *Service) Abort(ctx context.Context, id uint) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	var c Campaign
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return err
	}
	if terminal(c.State) && c.State != StateBlocked {
		return nil
	}
	c.State = StateAborted
	c.Revision++
	if err := s.db.WithContext(ctx).Save(&c).Error; err != nil {
		return err
	}
	err := s.db.WithContext(ctx).Model(&Target{}).Where("campaign_id = ? AND state NOT IN ?", id, []string{StateSucceeded, StateFailed}).Updates(map[string]any{"state": StateAborted}).Error
	if err == nil {
		auditCampaign(ctx, "fleet_update.abort", &c, "success")
	}
	return err
}

func (s *Service) Retry(ctx context.Context, id uint) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	var c Campaign
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return err
	}
	if c.State != StateFailed && c.State != StateAborted && c.State != StateBlocked {
		return fmt.Errorf("campaign is not retryable")
	}
	c.State = StateReady
	c.Error = ""
	if err := s.db.WithContext(ctx).Save(&c).Error; err != nil {
		return err
	}
	err := s.db.WithContext(ctx).Model(&Target{}).Where("campaign_id = ? AND state IN ?", id, []string{StateFailed, StateAborted, StateBlocked}).Updates(map[string]any{"state": StateReady, "blocked_reason": "", "error": "", "dispatch_status": "retry_requested"}).Error
	if err == nil {
		auditCampaign(ctx, "fleet_update.retry", &c, "success")
	}
	return err
}

func auditCampaign(ctx context.Context, event string, c *Campaign, outcome string) {
	_ = audit.Record(ctx, &audit.AuditEvent{EventType: event, Outcome: outcome, ActorType: "user", TargetType: "update_campaign", TargetRef: fmt.Sprint(c.ID), Metadata: "{}"})
}

func auditTarget(ctx context.Context, event string, c *Campaign, t *Target, outcome string) {
	_ = audit.Record(ctx, &audit.AuditEvent{EventType: event, Outcome: outcome, ActorType: "system", TargetType: "update_target", TargetRef: fmt.Sprintf("%d:%d", c.ID, t.NodeID), Metadata: fmt.Sprintf("{\"campaignId\":%d,\"runId\":%q}", c.ID, t.RunID)})
}
