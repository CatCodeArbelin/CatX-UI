package analytics

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:analytics-foundation-"+strings.ReplaceAll(t.Name(), "/", "-")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate setting: %v", err)
	}
	return db
}

func enableAnalytics(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Create(&model.Setting{Key: "fork.analytics.enabled", Value: "true"}).Error; err != nil {
		t.Fatalf("enable analytics: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("analytics migration: %v", err)
	}
}

func TestMigrationIsOptInAndSQLiteCRUDRetention(t *testing.T) {
	db := testDB(t)
	if db.Migrator().HasTable(&DestinationObservation{}) {
		t.Fatal("analytics table exists while feature is disabled")
	}
	enableAnalytics(t, db)

	for _, m := range []any{&DestinationObservation{}, &DNSObservation{}, &NetworkSession{}, &ServiceCategoryAggregate{}, &TrafficSnapshot{}, &EvidenceObservation{}} {
		if !db.Migrator().HasTable(m) {
			t.Fatalf("missing table for %T", m)
		}
	}
	repo := NewRepository(db, true)
	now := time.Now().Truncate(time.Millisecond)
	ctx := context.Background()
	if err := repo.RecordDestination(ctx, MetadataEvent{ObservedAt: now.UnixMilli(), ClientEmail: "alice", Domain: "example.com", Port: 443, Protocol: "tls", Source: SourceSNI, Provenance: ProvenanceObserved, Confidence: 1}); err != nil {
		t.Fatalf("record destination: %v", err)
	}
	if err := repo.RecordDNS(ctx, DNSObservation{ObservedAt: now.UnixMilli(), ClientEmail: "alice", Domain: "example.com", RecordType: "A", ResolvedIP: "192.0.2.1", Source: SourceDNSObserver, Provenance: ProvenanceObserved, Confidence: .9, ExpiresAt: now.Add(time.Hour).UnixMilli(), EventKey: "dns-test"}); err != nil {
		t.Fatalf("record dns: %v", err)
	}
	if err := repo.RecordDNS(ctx, DNSObservation{ObservedAt: now.UnixMilli(), ClientEmail: "alice", Domain: "example.com", RecordType: "A", ResolvedIP: "192.0.2.1", Source: SourceDNSObserver, Provenance: ProvenanceObserved, Confidence: .9, ExpiresAt: now.Add(time.Hour).UnixMilli(), EventKey: "dns-test"}); err != nil {
		t.Fatalf("deduplicate dns: %v", err)
	}
	activeDNS, err := repo.ActiveDNSForDestination(ctx, "alice", 0, 0, "192.0.2.1", now.UnixMilli()+1)
	if err != nil || len(activeDNS) != 1 {
		t.Fatalf("active dns = %d, %v", len(activeDNS), err)
	}
	if err := repo.RecordEvidence(ctx, EvidenceObservation{ObservedAt: now.UnixMilli(), ClientEmail: "alice", Domain: "example.com", DestinationIP: "192.0.2.1", Kind: EvidenceDNSDomain, Source: SourceDNSObserver, Provenance: ProvenanceCorrelated, Confidence: .75, Level: ConfidenceMedium, Selected: true, EventKey: "evidence-test"}); err != nil {
		t.Fatalf("record evidence: %v", err)
	}
	if err := repo.UpsertSession(ctx, NetworkSession{SessionKey: "s1", ClientEmail: "alice", FirstSeen: now.UnixMilli(), LastSeen: now.UnixMilli(), Protocol: "tls", Source: SourceXray, Provenance: ProvenanceCorrelated, Confidence: .8}); err != nil {
		t.Fatalf("record session: %v", err)
	}
	if err := repo.UpsertAggregate(ctx, ServiceCategoryAggregate{BucketStart: now.UnixMilli(), BucketWidth: "hour", ClientEmail: "alice", Category: "web", ObservationCount: 1, SessionCount: 1, FirstSeen: now.UnixMilli(), LastSeen: now.UnixMilli(), Source: SourceXray, Provenance: ProvenanceCorrelated, Confidence: .8}); err != nil {
		t.Fatalf("record aggregate: %v", err)
	}
	if err := repo.RecordTrafficSnapshots(ctx, []TrafficSnapshot{
		{ObservedAt: now.Add(-time.Minute).UnixMilli(), Scope: TrafficScopeClient, CounterKey: "alice", ClientEmail: "alice", Up: 100, Down: 200},
		{ObservedAt: now.UnixMilli(), Scope: TrafficScopeClient, CounterKey: "alice", ClientEmail: "alice", Up: 175, Down: 500},
	}); err != nil {
		t.Fatalf("record traffic snapshots: %v", err)
	}
	history, err := repo.QueryTrafficHistory(ctx, "alice", now.Add(-time.Minute).UnixMilli(), now.Add(time.Minute).UnixMilli())
	if err != nil || history.Up != 75 || history.Down != 300 {
		t.Fatalf("traffic history = %+v, %v", history, err)
	}
	rows, err := repo.ListDestinations(ctx, "alice", now.Add(-time.Minute).UnixMilli(), now.Add(time.Minute).UnixMilli(), 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("list destinations = %d, %v", len(rows), err)
	}
	summary, err := repo.SummarizeSessions(ctx, "alice", now.Add(-time.Minute).UnixMilli(), now.Add(time.Minute).UnixMilli())
	if err != nil || summary.Count != 1 || summary.FirstSeen != now.UnixMilli() || summary.LastSeen != now.UnixMilli() {
		t.Fatalf("session summary = %+v, %v", summary, err)
	}

	old := now.Add(-48 * time.Hour).UnixMilli()
	if err := db.Create(&DestinationObservation{ObservedAt: old, Domain: "old.example", Source: SourceAccessLog, Provenance: ProvenanceObserved, Confidence: 1}).Error; err != nil {
		t.Fatalf("seed old event: %v", err)
	}
	if err := db.Create(&DNSObservation{ObservedAt: old, Domain: "old-dns.example", ResolvedIP: "192.0.2.8", Source: SourceDNSObserver, Provenance: ProvenanceObserved, Confidence: 1, EventKey: "old-dns"}).Error; err != nil {
		t.Fatalf("seed old DNS: %v", err)
	}
	if err := db.Create(&NetworkSession{SessionKey: "old-session", FirstSeen: old, LastSeen: old, Source: SourceXray, Provenance: ProvenanceCorrelated, Confidence: 1}).Error; err != nil {
		t.Fatalf("seed old session: %v", err)
	}
	if err := db.Create(&ServiceCategoryAggregate{BucketStart: old, BucketWidth: BucketDaily, Category: CategoryUnknown, FirstSeen: old, LastSeen: old, Source: SourceXray, Provenance: ProvenanceCorrelated, Confidence: 1}).Error; err != nil {
		t.Fatalf("seed old aggregate: %v", err)
	}
	result, err := repo.Prune(ctx, now, RetentionPolicy{RawEvents: 24 * time.Hour, DNSObservations: time.Hour, Sessions: 24 * time.Hour, Aggregates: 24 * time.Hour})
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if result.RawEvents != 1 {
		t.Fatalf("pruned raw events = %d, want 1", result.RawEvents)
	}
	if result.DNSObservations != 1 || result.Sessions != 1 || result.Aggregates != 1 {
		t.Fatalf("independent prune counts = %+v", result)
	}
	var upstream int64
	if err := db.Model(&model.Setting{}).Count(&upstream).Error; err != nil || upstream != 1 {
		t.Fatalf("upstream settings changed: count=%d err=%v", upstream, err)
	}
}

func TestDisabledRepositoryIsNoOp(t *testing.T) {
	db := testDB(t)
	repo := NewRepository(db, false)
	if err := repo.RecordDestination(context.Background(), MetadataEvent{}); err != nil {
		t.Fatalf("disabled record: %v", err)
	}
	rows, err := repo.ListDestinations(context.Background(), "", 0, 1, 1)
	if err != nil || rows != nil {
		t.Fatalf("disabled list = %#v, %v", rows, err)
	}
}

type fakeRuntime struct{}

func (fakeRuntime) GetOnlineUsers() ([]xray.OnlineUser, error) {
	return []xray.OnlineUser{{Email: "alice"}}, nil
}

func (fakeRuntime) GetTraffic() ([]*xray.Traffic, []*xray.ClientTraffic, error) {
	return []*xray.Traffic{{Tag: "inbound-1"}}, []*xray.ClientTraffic{{Email: "alice", Up: 10}}, nil
}

func TestXrayRuntimeAdapterReadsExistingRuntimeData(t *testing.T) {
	snapshot, err := NewXrayRuntimeAdapter(fakeRuntime{}).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.OnlineUsers) != 1 || len(snapshot.ClientTraffic) != 1 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestPostgresSchemaAndCRUD(t *testing.T) {
	if os.Getenv("XUI_DB_TYPE") != "postgres" || strings.TrimSpace(os.Getenv("XUI_DB_DSN")) == "" {
		t.Skip("set XUI_DB_TYPE=postgres and XUI_DB_DSN to run analytics PostgreSQL test")
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("XUI_DB_DSN")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	for _, m := range []any{&DestinationObservation{}, &DNSObservation{}, &NetworkSession{}, &ServiceCategoryAggregate{}, &TrafficSnapshot{}, &EvidenceObservation{}} {
		_ = db.Migrator().DropTable(m)
	}
	t.Cleanup(func() {
		for _, m := range []any{&DestinationObservation{}, &DNSObservation{}, &NetworkSession{}, &ServiceCategoryAggregate{}, &TrafficSnapshot{}, &EvidenceObservation{}} {
			_ = db.Migrator().DropTable(m)
		}
	})
	if err := Migrate(db); err != nil {
		t.Fatalf("postgres migration: %v", err)
	}
	if err := NewRepository(db, true).RecordDestination(context.Background(), MetadataEvent{ObservedAt: time.Now().UnixMilli(), Domain: "example.com", Source: SourceDestination, Provenance: ProvenanceObserved, Confidence: 1}); err != nil {
		t.Fatalf("postgres CRUD: %v", err)
	}
	if err := NewRepository(db, true).RecordEvidence(context.Background(), EvidenceObservation{ObservedAt: time.Now().UnixMilli(), Domain: "example.com", DestinationIP: "192.0.2.1", Kind: EvidenceDirectDomain, Source: SourceDestination, Provenance: ProvenanceObserved, Confidence: 1, Level: ConfidenceHigh, EventKey: "postgres-evidence"}); err != nil {
		t.Fatalf("postgres evidence CRUD: %v", err)
	}
}
