package analytics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RetentionPolicy struct {
	RawEvents       time.Duration
	DNSObservations time.Duration
	Sessions        time.Duration
	Aggregates      time.Duration
}

func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{RawEvents: 7 * 24 * time.Hour, DNSObservations: 7 * 24 * time.Hour, Sessions: 30 * 24 * time.Hour, Aggregates: 365 * 24 * time.Hour}
}

type Repository interface {
	RecordDestination(ctx context.Context, event MetadataEvent) error
	RecordDNS(ctx context.Context, observation DNSObservation) error
	RecordEvidence(ctx context.Context, evidence EvidenceObservation) error
	ActiveDNSForDestination(ctx context.Context, clientEmail string, nodeID, inboundID int, destinationIP string, observedAt int64) ([]DNSObservation, error)
	UpsertSession(ctx context.Context, session NetworkSession) error
	UpsertAggregate(ctx context.Context, aggregate ServiceCategoryAggregate) error
	ListDestinations(ctx context.Context, clientEmail string, from, to int64, limit int) ([]DestinationObservation, error)
	ListDestinationPage(ctx context.Context, clientEmail string, from, to int64, limit, offset int) (DestinationPage, error)
	ListDNSPage(ctx context.Context, clientEmail string, from, to int64, limit, offset int) (DNSPage, error)
	ListSessionPage(ctx context.Context, clientEmail string, from, to int64, limit, offset int) (SessionPage, error)
	ListAggregates(ctx context.Context, clientEmail, bucketWidth string, from, to int64) ([]ServiceCategoryAggregate, error)
	RecordTrafficSnapshots(ctx context.Context, snapshots []TrafficSnapshot) error
	QueryTrafficHistory(ctx context.Context, clientEmail string, from, to int64) (TrafficHistory, error)
	SummarizeSessions(ctx context.Context, clientEmail string, from, to int64) (SessionSummary, error)
	Prune(ctx context.Context, now time.Time, policy RetentionPolicy) (PruneResult, error)
	CommitAccessLogBatch(ctx context.Context, events []MetadataEvent, sessions []NetworkSession, cursor AccessLogCursor) error
	LoadAccessLogCursor(ctx context.Context, cursorKey string) (AccessLogCursor, error)
}

type SessionSummary struct {
	Count     int64
	FirstSeen int64
	LastSeen  int64
}

type DestinationPage struct {
	Items []DestinationObservation
	Total int64
}

type DNSPage struct {
	Items []DNSObservation
	Total int64
}

type SessionPage struct {
	Items []NetworkSession
	Total int64
}

type PruneResult struct{ RawEvents, DNSObservations, Evidence, Sessions, Aggregates, TrafficSnapshots int64 }

type TrafficHistory struct {
	From              int64              `json:"from"`
	To                int64              `json:"to"`
	Up                int64              `json:"up"`
	Down              int64              `json:"down"`
	Clients           int64              `json:"clients"`
	Inbounds          int64              `json:"inbounds"`
	Nodes             int64              `json:"nodes"`
	ServiceBreakdown  []TrafficBreakdown `json:"serviceBreakdown"`
	CategoryBreakdown []TrafficBreakdown `json:"categoryBreakdown"`
}

type TrafficBreakdown struct {
	Name         string `json:"name"`
	Observations int64  `json:"observations"`
	Sessions     int64  `json:"sessions"`
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB, enabled bool) Repository {
	if !enabled || db == nil {
		return NoopRepository{}
	}
	return &GormRepository{db: db}
}

func (r *GormRepository) RecordDestination(ctx context.Context, event MetadataEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	observation := DestinationObservation{ /* populated below for readable validation boundary */
		ObservedAt: event.ObservedAt, ClientEmail: event.ClientEmail, ClientGroup: event.ClientGroup, NodeID: event.NodeID, InboundID: event.InboundID,
		Domain: event.Domain, DestinationIP: event.DestinationIP, Port: event.Port, Protocol: event.Protocol, SNI: event.SNI, Category: event.Category, Service: event.Service, ASN: event.ASN, Country: event.Country,
		ClassificationSource: event.ClassificationSource, ClassificationProvenance: event.ClassificationProvenance, ClassificationConfidence: event.ClassificationConfidence, ClassificationLevel: event.ClassificationLevel, ClassificationFirstParty: event.ClassificationFirstParty, ClassificationConflict: event.ClassificationConflict, ClassificationCandidates: event.ClassificationCandidates, ClassificationReason: event.ClassificationReason,
		SessionKey: event.SessionKey, EventKey: event.EventKey, Source: event.Source, Provenance: event.Provenance, Confidence: event.Confidence,
	}
	if event.EventKey == "" {
		return r.db.WithContext(ctx).Create(&observation).Error
	}
	return r.db.WithContext(ctx).Clauses(eventKeyDoNothing()).Create(&observation).Error
}

func (r *GormRepository) CommitAccessLogBatch(ctx context.Context, events []MetadataEvent, sessions []NetworkSession, cursor AccessLogCursor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if err := event.Validate(); err != nil {
				return err
			}
			observation := event.Observation()
			if err := tx.Clauses(eventKeyDoNothing()).Create(&observation).Error; err != nil {
				return err
			}
		}
		for _, session := range sessions {
			if session.SessionKey == "" || session.FirstSeen <= 0 || session.LastSeen < session.FirstSeen {
				return errors.New("invalid analytics network session")
			}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "session_key"}}, DoUpdates: clause.AssignmentColumns([]string{"client_email", "node_id", "inbound_id", "last_seen", "protocol", "source", "provenance", "confidence"})}).Create(&session).Error; err != nil {
				return err
			}
		}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "cursor_key"}}, DoUpdates: clause.AssignmentColumns([]string{"file_identity", "offset", "generation", "updated_at"})}).Create(&cursor).Error
	})
}

func (r *GormRepository) LoadAccessLogCursor(ctx context.Context, cursorKey string) (AccessLogCursor, error) {
	var cursor AccessLogCursor
	err := r.db.WithContext(ctx).Where("cursor_key = ?", cursorKey).First(&cursor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AccessLogCursor{CursorKey: cursorKey}, nil
	}
	return cursor, err
}

func (r *GormRepository) RecordDNS(ctx context.Context, observation DNSObservation) error {
	if observation.ObservedAt <= 0 || observation.Domain == "" || observation.Source == "" || observation.Provenance == "" || observation.Confidence < 0 || observation.Confidence > 1 {
		return errors.New("invalid analytics DNS observation")
	}
	if observation.EventKey == "" {
		observation.EventKey = digest(strings.Join([]string{observation.ClientEmail, observation.Domain, observation.ResolvedIP, observation.RecordType, strconv.FormatInt(observation.ObservedAt, 10), strconv.FormatInt(observation.ExpiresAt, 10)}, "|"))
	}
	return r.db.WithContext(ctx).Clauses(eventKeyDoNothing()).Create(&observation).Error
}

func eventKeyDoNothing() clause.OnConflict {
	return clause.OnConflict{
		Columns:     []clause.Column{{Name: "event_key"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "event_key <> ''"}}},
		DoNothing:   true,
	}
}

func (r *GormRepository) RecordEvidence(ctx context.Context, evidence EvidenceObservation) error {
	if evidence.ObservedAt <= 0 || evidence.Domain == "" && evidence.DestinationIP == "" || evidence.Kind == "" || evidence.Source == "" || evidence.Provenance == "" || evidence.Level == "" || evidence.Confidence < 0 || evidence.Confidence > 1 {
		return errors.New("invalid analytics evidence observation")
	}
	if evidence.EventKey == "" {
		evidence.EventKey = digest(strings.Join([]string{evidence.ClientEmail, evidence.DestinationIP, evidence.Domain, evidence.Kind, strconv.FormatInt(evidence.ObservedAt, 10), evidence.SessionKey}, "|"))
	}
	return r.db.WithContext(ctx).Clauses(eventKeyDoNothing()).Create(&evidence).Error
}

func (r *GormRepository) ActiveDNSForDestination(ctx context.Context, clientEmail string, nodeID, inboundID int, destinationIP string, observedAt int64) ([]DNSObservation, error) {
	q := r.db.WithContext(ctx).Where("client_email = ? AND node_id = ? AND inbound_id = ? AND resolved_ip = ? AND observed_at <= ? AND (expires_at = 0 OR expires_at > ?)", clientEmail, nodeID, inboundID, destinationIP, observedAt, observedAt)
	var rows []DNSObservation
	err := q.Order("observed_at DESC, id DESC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) UpsertSession(ctx context.Context, session NetworkSession) error {
	if session.SessionKey == "" || session.FirstSeen <= 0 || session.LastSeen < session.FirstSeen || session.Source == "" || session.Provenance == "" || session.Confidence < 0 || session.Confidence > 1 {
		return errors.New("invalid analytics network session")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"client_email", "node_id", "inbound_id", "last_seen", "protocol", "source", "provenance", "confidence"}),
	}).Create(&session).Error
}

func (r *GormRepository) UpsertAggregate(ctx context.Context, aggregate ServiceCategoryAggregate) error {
	if aggregate.BucketStart <= 0 || aggregate.BucketWidth == "" || aggregate.FirstSeen <= 0 || aggregate.LastSeen < aggregate.FirstSeen || aggregate.Source == "" || aggregate.Provenance == "" || aggregate.Confidence < 0 || aggregate.Confidence > 1 {
		return errors.New("invalid analytics aggregate")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "bucket_start"}, {Name: "bucket_width"}, {Name: "client_email"}, {Name: "category"}},
		DoUpdates: clause.AssignmentColumns([]string{"client_group", "observation_count", "session_count", "first_seen", "last_seen", "source", "provenance", "confidence"}),
	}).Create(&aggregate).Error
}

func (r *GormRepository) ListDestinations(ctx context.Context, clientEmail string, from, to int64, limit int) ([]DestinationObservation, error) {
	if limit <= 0 {
		limit = 1000
	}
	q := r.db.WithContext(ctx).Where("observed_at >= ? AND observed_at < ?", from, to).Order("observed_at ASC").Limit(limit)
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	var rows []DestinationObservation
	return rows, q.Find(&rows).Error
}

func (r *GormRepository) ListDestinationPage(ctx context.Context, clientEmail string, from, to int64, limit, offset int) (DestinationPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	q := r.db.WithContext(ctx).Model(&DestinationObservation{}).Where("observed_at >= ? AND observed_at < ?", from, to)
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return DestinationPage{}, err
	}
	var rows []DestinationObservation
	err := q.Order("observed_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return DestinationPage{Items: rows, Total: total}, err
}

func (r *GormRepository) ListSessionPage(ctx context.Context, clientEmail string, from, to int64, limit, offset int) (SessionPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	q := r.db.WithContext(ctx).Model(&NetworkSession{}).Where("last_seen >= ? AND first_seen < ?", from, to)
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return SessionPage{}, err
	}
	var rows []NetworkSession
	err := q.Order("last_seen DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return SessionPage{Items: rows, Total: total}, err
}

func (r *GormRepository) ListDNSPage(ctx context.Context, clientEmail string, from, to int64, limit, offset int) (DNSPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	q := r.db.WithContext(ctx).Model(&DNSObservation{}).Where("observed_at >= ? AND observed_at < ?", from, to)
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return DNSPage{}, err
	}
	var rows []DNSObservation
	err := q.Order("observed_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return DNSPage{Items: rows, Total: total}, err
}

func (r *GormRepository) ListAggregates(ctx context.Context, clientEmail, bucketWidth string, from, to int64) ([]ServiceCategoryAggregate, error) {
	q := r.db.WithContext(ctx).Where("bucket_start >= ? AND bucket_start < ?", from, to).Order("bucket_start ASC")
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	if bucketWidth != "" {
		q = q.Where("bucket_width = ?", bucketWidth)
	}
	var rows []ServiceCategoryAggregate
	return rows, q.Find(&rows).Error
}

func (r *GormRepository) RecordTrafficSnapshots(ctx context.Context, snapshots []TrafficSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	for _, snapshot := range snapshots {
		if snapshot.ObservedAt <= 0 || snapshot.CounterKey == "" || snapshot.Scope == "" || snapshot.Up < 0 || snapshot.Down < 0 {
			return errors.New("invalid traffic snapshot")
		}
	}
	return r.db.WithContext(ctx).Create(&snapshots).Error
}

func (r *GormRepository) QueryTrafficHistory(ctx context.Context, clientEmail string, from, to int64) (TrafficHistory, error) {
	if from < 0 || to <= from {
		return TrafficHistory{}, errors.New("invalid traffic range")
	}
	q := r.db.WithContext(ctx).Where("observed_at <= ?", to).Order("counter_key ASC, observed_at ASC, id ASC")
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	var rows []TrafficSnapshot
	if err := q.Find(&rows).Error; err != nil {
		return TrafficHistory{}, err
	}
	result := TrafficHistory{From: from, To: to}
	last := map[string]TrafficSnapshot{}
	seenClient, seenInbound, seenNode := map[string]struct{}{}, map[string]struct{}{}, map[int]struct{}{}
	for _, row := range rows {
		previous, hasPrevious := last[row.Scope+"\x00"+row.CounterKey]
		last[row.Scope+"\x00"+row.CounterKey] = row
		if row.ObservedAt < from || row.ObservedAt > to || !hasPrevious {
			continue
		}
		if row.Up < previous.Up || row.Down < previous.Down {
			continue
		}
		result.Up += row.Up - previous.Up
		result.Down += row.Down - previous.Down
		if row.ClientEmail != "" {
			seenClient[row.ClientEmail] = struct{}{}
		}
		if row.InboundID != 0 {
			seenInbound[fmt.Sprint(row.InboundID)] = struct{}{}
		}
		if row.NodeID != 0 {
			seenNode[row.NodeID] = struct{}{}
		}
	}
	result.Clients, result.Inbounds, result.Nodes = int64(len(seenClient)), int64(len(seenInbound)), int64(len(seenNode))
	obs, err := r.destinationBreakdown(ctx, clientEmail, from, to)
	if err != nil {
		return TrafficHistory{}, err
	}
	result.ServiceBreakdown, result.CategoryBreakdown = obs.services, obs.categories
	return result, nil
}

type breakdownResult struct{ services, categories []TrafficBreakdown }

func (r *GormRepository) destinationBreakdown(ctx context.Context, clientEmail string, from, to int64) (breakdownResult, error) {
	base := r.db.WithContext(ctx).Model(&DestinationObservation{}).Where("observed_at >= ? AND observed_at < ?", from, to)
	if clientEmail != "" {
		base = base.Where("client_email = ?", clientEmail)
	}
	collect := func(column string) ([]TrafficBreakdown, error) {
		var rows []TrafficBreakdown
		if err := base.Select(column + " AS name, COUNT(*) AS observations, COUNT(DISTINCT session_key) AS sessions").Where(column + " <> ''").Group(column).Order("observations DESC, name ASC").Scan(&rows).Error; err != nil {
			return nil, err
		}
		return rows, nil
	}
	services, err := collect("service")
	if err != nil {
		return breakdownResult{}, err
	}
	categories, err := collect("category")
	if err != nil {
		return breakdownResult{}, err
	}
	return breakdownResult{services: services, categories: categories}, nil
}

func (r *GormRepository) SummarizeSessions(ctx context.Context, clientEmail string, from, to int64) (SessionSummary, error) {
	q := r.db.WithContext(ctx).Model(&NetworkSession{}).Where("last_seen >= ? AND first_seen < ?", from, to)
	if clientEmail != "" {
		q = q.Where("client_email = ?", clientEmail)
	}
	var summary SessionSummary
	err := q.Select("COUNT(*) AS count, COALESCE(MIN(first_seen), 0) AS first_seen, COALESCE(MAX(last_seen), 0) AS last_seen").Scan(&summary).Error
	return summary, err
}

func (r *GormRepository) Prune(ctx context.Context, now time.Time, policy RetentionPolicy) (PruneResult, error) {
	// Preserve the pre-WP-2C policy contract for callers that do not yet
	// provide a DNS-specific duration: DNS then follows raw-event retention.
	if policy.DNSObservations <= 0 {
		policy.DNSObservations = policy.RawEvents
	}
	if policy.RawEvents <= 0 || policy.Sessions <= 0 || policy.Aggregates <= 0 {
		return PruneResult{}, errors.New("analytics retention durations must be positive")
	}
	var out PruneResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range []struct {
			model  any
			column string
			before int64
			count  *int64
		}{
			{&DestinationObservation{}, "observed_at", now.Add(-policy.RawEvents).UnixMilli(), &out.RawEvents},
			{&DNSObservation{}, "observed_at", now.Add(-policy.DNSObservations).UnixMilli(), &out.DNSObservations},
			{&EvidenceObservation{}, "observed_at", now.Add(-policy.RawEvents).UnixMilli(), &out.Evidence},
			{&NetworkSession{}, "last_seen", now.Add(-policy.Sessions).UnixMilli(), &out.Sessions},
			{&ServiceCategoryAggregate{}, "bucket_start", now.Add(-policy.Aggregates).UnixMilli(), &out.Aggregates},
			{&TrafficSnapshot{}, "observed_at", now.Add(-policy.Aggregates).UnixMilli(), &out.TrafficSnapshots},
		} {
			result := tx.Where(item.column+" < ?", item.before).Delete(item.model)
			if result.Error != nil {
				return result.Error
			}
			*item.count = result.RowsAffected
		}
		return nil
	})
	return out, err
}

type NoopRepository struct{}

func (NoopRepository) RecordDestination(context.Context, MetadataEvent) error {
	return nil
}

func (NoopRepository) RecordDNS(context.Context, DNSObservation) error {
	return nil
}

func (NoopRepository) RecordEvidence(context.Context, EvidenceObservation) error { return nil }

func (NoopRepository) ActiveDNSForDestination(context.Context, string, int, int, string, int64) ([]DNSObservation, error) {
	return nil, nil
}

func (NoopRepository) UpsertSession(context.Context, NetworkSession) error {
	return nil
}

func (NoopRepository) UpsertAggregate(context.Context, ServiceCategoryAggregate) error {
	return nil
}

func (NoopRepository) ListDestinations(context.Context, string, int64, int64, int) ([]DestinationObservation, error) {
	return nil, nil
}

func (NoopRepository) ListDestinationPage(context.Context, string, int64, int64, int, int) (DestinationPage, error) {
	return DestinationPage{Items: []DestinationObservation{}}, nil
}

func (NoopRepository) ListDNSPage(context.Context, string, int64, int64, int, int) (DNSPage, error) {
	return DNSPage{Items: []DNSObservation{}}, nil
}

func (NoopRepository) ListSessionPage(context.Context, string, int64, int64, int, int) (SessionPage, error) {
	return SessionPage{Items: []NetworkSession{}}, nil
}

func (NoopRepository) ListAggregates(context.Context, string, string, int64, int64) ([]ServiceCategoryAggregate, error) {
	return nil, nil
}

func (NoopRepository) RecordTrafficSnapshots(context.Context, []TrafficSnapshot) error { return nil }
func (NoopRepository) QueryTrafficHistory(_ context.Context, _ string, from, to int64) (TrafficHistory, error) {
	return TrafficHistory{From: from, To: to, ServiceBreakdown: []TrafficBreakdown{}, CategoryBreakdown: []TrafficBreakdown{}}, nil
}

func (NoopRepository) SummarizeSessions(context.Context, string, int64, int64) (SessionSummary, error) {
	return SessionSummary{}, nil
}

func (NoopRepository) Prune(context.Context, time.Time, RetentionPolicy) (PruneResult, error) {
	return PruneResult{}, nil
}

func (NoopRepository) CommitAccessLogBatch(context.Context, []MetadataEvent, []NetworkSession, AccessLogCursor) error {
	return nil
}

func (NoopRepository) LoadAccessLogCursor(_ context.Context, key string) (AccessLogCursor, error) {
	return AccessLogCursor{CursorKey: key}, nil
}

type activityQuery struct {
	From     int64
	To       int64
	Page     int
	PageSize int
}

func parseActivityQuery(c *gin.Context) (activityQuery, error) {
	now := time.Now().UnixMilli()
	q := activityQuery{From: now - 7*24*time.Hour.Milliseconds(), To: now, Page: 1, PageSize: 50}
	var err error
	if value := c.Query("from"); value != "" {
		q.From, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return q, fmt.Errorf("invalid from")
		}
	}
	if value := c.Query("to"); value != "" {
		q.To, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return q, fmt.Errorf("invalid to")
		}
	}
	if value := c.Query("page"); value != "" {
		q.Page, err = strconv.Atoi(value)
		if err != nil || q.Page < 1 {
			return q, fmt.Errorf("invalid page")
		}
	}
	if value := c.Query("pageSize"); value != "" {
		q.PageSize, err = strconv.Atoi(value)
		if err != nil || q.PageSize < 1 {
			return q, fmt.Errorf("invalid pageSize")
		}
	}
	if q.PageSize > 200 {
		q.PageSize = 200
	}
	if q.From < 0 || q.To <= q.From || q.To > now+5*time.Minute.Milliseconds() || q.To-q.From > (400*24*time.Hour).Milliseconds() {
		return q, fmt.Errorf("invalid time range")
	}
	return q, nil
}

func activityEnvelope(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": obj})
}

// RegisterActivityRoutes exposes read-only analytics views through the fixed
// fork route hook. It never returns raw log lines or any content data.
func RegisterActivityRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/analytics/status", func(c *gin.Context) {
		status := CurrentStatus()
		activityEnvelope(c, gin.H{"enabled": status.Enabled, "dnsIntelligence": status.DNSIntelligence})
	})
	api.GET("/analytics/traffic", func(c *gin.Context) {
		configured.RLock()
		repo, enabled := configured.repo, configured.enabled
		configured.RUnlock()
		query, err := parseActivityQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		client := c.Query("clientEmail")
		result := TrafficHistory{From: query.From, To: query.To, ServiceBreakdown: []TrafficBreakdown{}, CategoryBreakdown: []TrafficBreakdown{}}
		if enabled && repo != nil {
			result, err = repo.QueryTrafficHistory(c.Request.Context(), client, query.From, query.To)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "traffic history unavailable"})
			return
		}
		activityEnvelope(c, gin.H{"enabled": enabled, "history": result})
	})
	api.GET("/analytics/clients/:email/traffic", func(c *gin.Context) {
		configured.RLock()
		repo, enabled := configured.repo, configured.enabled
		configured.RUnlock()
		query, err := parseActivityQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		result := TrafficHistory{From: query.From, To: query.To, ServiceBreakdown: []TrafficBreakdown{}, CategoryBreakdown: []TrafficBreakdown{}}
		if enabled && repo != nil {
			result, err = repo.QueryTrafficHistory(c.Request.Context(), c.Param("email"), query.From, query.To)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "traffic history unavailable"})
			return
		}
		activityEnvelope(c, gin.H{"enabled": enabled, "history": result})
	})
	api.GET("/analytics/clients/:email/activity", func(c *gin.Context) {
		configured.RLock()
		repo, enabled := configured.repo, configured.enabled
		configured.RUnlock()
		query, err := parseActivityQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		page := DestinationPage{Items: []DestinationObservation{}}
		if enabled && repo != nil {
			page, err = repo.ListDestinationPage(c.Request.Context(), c.Param("email"), query.From, query.To, query.PageSize, (query.Page-1)*query.PageSize)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "analytics unavailable"})
			return
		}
		activityEnvelope(c, gin.H{"enabled": enabled, "items": page.Items, "page": query.Page, "pageSize": query.PageSize, "total": page.Total, "from": query.From, "to": query.To})
	})
	api.GET("/analytics/clients/:email/sessions", func(c *gin.Context) {
		configured.RLock()
		repo, enabled := configured.repo, configured.enabled
		configured.RUnlock()
		query, err := parseActivityQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		page := SessionPage{Items: []NetworkSession{}}
		if enabled && repo != nil {
			page, err = repo.ListSessionPage(c.Request.Context(), c.Param("email"), query.From, query.To, query.PageSize, (query.Page-1)*query.PageSize)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "analytics unavailable"})
			return
		}
		activityEnvelope(c, gin.H{"enabled": enabled, "items": page.Items, "page": query.Page, "pageSize": query.PageSize, "total": page.Total, "from": query.From, "to": query.To})
	})
	api.GET("/analytics/clients/:email/dns", func(c *gin.Context) {
		configured.RLock()
		repo, enabled, dnsEnabled := configured.repo, configured.enabled, configured.evidenceEnabled
		configured.RUnlock()
		query, err := parseActivityQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		page := DNSPage{Items: []DNSObservation{}}
		if enabled && dnsEnabled && repo != nil {
			page, err = repo.ListDNSPage(c.Request.Context(), c.Param("email"), query.From, query.To, query.PageSize, (query.Page-1)*query.PageSize)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "DNS intelligence unavailable"})
			return
		}
		activityEnvelope(c, gin.H{"enabled": enabled && dnsEnabled, "items": page.Items, "page": query.Page, "pageSize": query.PageSize, "total": page.Total, "from": query.From, "to": query.To})
	})
}
