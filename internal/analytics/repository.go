package analytics

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RetentionPolicy struct {
	RawEvents  time.Duration
	Sessions   time.Duration
	Aggregates time.Duration
}

func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{RawEvents: 7 * 24 * time.Hour, Sessions: 30 * 24 * time.Hour, Aggregates: 365 * 24 * time.Hour}
}

type Repository interface {
	RecordDestination(ctx context.Context, event MetadataEvent) error
	RecordDNS(ctx context.Context, observation DNSObservation) error
	UpsertSession(ctx context.Context, session NetworkSession) error
	UpsertAggregate(ctx context.Context, aggregate ServiceCategoryAggregate) error
	ListDestinations(ctx context.Context, clientEmail string, from, to int64, limit int) ([]DestinationObservation, error)
	ListAggregates(ctx context.Context, clientEmail, bucketWidth string, from, to int64) ([]ServiceCategoryAggregate, error)
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

type PruneResult struct{ RawEvents, DNSObservations, Sessions, Aggregates int64 }

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
		Domain: event.Domain, DestinationIP: event.DestinationIP, Port: event.Port, Protocol: event.Protocol, SNI: event.SNI, Category: event.Category,
		SessionKey: event.SessionKey, EventKey: event.EventKey, Source: event.Source, Provenance: event.Provenance, Confidence: event.Confidence,
	}
	if event.EventKey == "" {
		return r.db.WithContext(ctx).Create(&observation).Error
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_key"}}, DoNothing: true}).Create(&observation).Error
}

func (r *GormRepository) CommitAccessLogBatch(ctx context.Context, events []MetadataEvent, sessions []NetworkSession, cursor AccessLogCursor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if err := event.Validate(); err != nil {
				return err
			}
			observation := event.Observation()
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_key"}}, DoNothing: true}).Create(&observation).Error; err != nil {
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
	return r.db.WithContext(ctx).Create(&observation).Error
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
			{&DNSObservation{}, "observed_at", now.Add(-policy.RawEvents).UnixMilli(), &out.DNSObservations},
			{&NetworkSession{}, "last_seen", now.Add(-policy.Sessions).UnixMilli(), &out.Sessions},
			{&ServiceCategoryAggregate{}, "bucket_start", now.Add(-policy.Aggregates).UnixMilli(), &out.Aggregates},
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

func (NoopRepository) UpsertSession(context.Context, NetworkSession) error {
	return nil
}

func (NoopRepository) UpsertAggregate(context.Context, ServiceCategoryAggregate) error {
	return nil
}

func (NoopRepository) ListDestinations(context.Context, string, int64, int64, int) ([]DestinationObservation, error) {
	return nil, nil
}

func (NoopRepository) ListAggregates(context.Context, string, string, int64, int64) ([]ServiceCategoryAggregate, error) {
	return nil, nil
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
