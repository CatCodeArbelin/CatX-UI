package risk

import (
	"context"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func riskDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:risk-"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&analytics.NetworkSession{}, &analytics.DNSObservation{}); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestReplaySafeIPHistoryAndExplainableScore(t *testing.T) {
	db := riskDB(t)
	Configure(db, true)
	analytics.Configure(analytics.NewRepository(db, true), true)
	t.Cleanup(func() { Configure(nil, false); analytics.Configure(analytics.NoopRepository{}, false) })
	now := time.Now().UnixMilli()
	if err := db.Create(&analytics.NetworkSession{SessionKey: "s1", ClientEmail: "alice", FirstSeen: now - 1000, LastSeen: now, Source: analytics.SourceXray, Provenance: analytics.ProvenanceObserved, Confidence: 1}).Error; err != nil {
		t.Fatal(err)
	}
	entries := map[string][]model.ClientIpEntry{"alice": {{IP: "198.51.100.10", Timestamp: now / 1000}, {IP: "198.51.100.11", Timestamp: now / 1000}}}
	if err := RecordIPHistory(context.Background(), "node-a", entries, "test"); err != nil {
		t.Fatal(err)
	}
	if err := RecordIPHistory(context.Background(), "node-a", entries, "test"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&IPHistory{}).Where("client_email = ?", "alice").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("replay created %d rows, want 2", count)
	}
	summary, err := SummaryFor(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Score == 0 || summary.Score >= 70 {
		t.Fatalf("score = %d, want bounded informational score", summary.Score)
	}
	if summary.Confidence >= 1 {
		t.Fatalf("confidence = %v, provider absence must remain degraded", summary.Confidence)
	}
}

func TestSuppressionAcknowledgementAndIndependentDeletion(t *testing.T) {
	db := riskDB(t)
	Configure(db, true)
	analytics.Configure(analytics.NewRepository(db, true), true)
	t.Cleanup(func() { Configure(nil, false); analytics.Configure(analytics.NoopRepository{}, false) })
	now := time.Now().UnixMilli()
	if err := db.Create(&Event{ClientEmail: "bob", Kind: SignalIPChurn, ObservedAt: now, DedupeKey: "event-1"}).Error; err != nil {
		t.Fatal(err)
	}
	var event Event
	if err := db.Where("dedupe_key = ?", "event-1").First(&event).Error; err != nil {
		t.Fatal(err)
	}
	if err := Acknowledge(context.Background(), "bob", event.ID); err != nil {
		t.Fatal(err)
	}
	if err := Suppress(context.Background(), "bob", SignalIPChurn, now+time.Hour.Milliseconds(), "known travel"); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&IPHistory{ClientEmail: "bob", IP: "198.51.100.20", ObservedAt: now, IngestedAt: now, DedupeKey: "ip-1"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteHistory(context.Background(), "bob"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&IPHistory{}).Where("client_email = ?", "bob").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("history count = %d after independent deletion", count)
	}
}

func TestSingleDNSSignalCannotBecomeHighRisk(t *testing.T) {
	db := riskDB(t)
	Configure(db, true)
	analytics.Configure(analytics.NewRepository(db, true), true)
	t.Cleanup(func() { Configure(nil, false); analytics.Configure(analytics.NoopRepository{}, false) })
	now := time.Now().UnixMilli()
	if err := db.Create(&analytics.NetworkSession{SessionKey: "dns", ClientEmail: "carol", FirstSeen: now, LastSeen: now, Source: analytics.SourceXray, Provenance: analytics.ProvenanceObserved, Confidence: 1}).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if err := db.Create(&analytics.DNSObservation{ClientEmail: "carol", ObservedAt: now, Domain: "unique-" + string(rune('a'+i%26)) + ".example", Source: analytics.SourceDNSObserver, Provenance: analytics.ProvenanceObserved, Confidence: 1}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := RecordIPHistory(context.Background(), "node-a", map[string][]model.ClientIpEntry{"carol": {{IP: "198.51.100.30", Timestamp: now / 1000}}}, "test"); err != nil {
		t.Fatal(err)
	}
	summary, err := SummaryFor(context.Background(), "carol")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Score > 8 {
		t.Fatalf("DNS-only score = %d, want <= 8", summary.Score)
	}
}
