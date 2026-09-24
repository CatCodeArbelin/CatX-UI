package analytics

import (
	"context"
	"sync"
	"testing"
	"time"
)

type evidenceMemoryRepo struct {
	NoopRepository
	mu       sync.Mutex
	dns      []DNSObservation
	evidence []EvidenceObservation
}

func (r *evidenceMemoryRepo) RecordDNS(_ context.Context, row DNSObservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.dns {
		if existing.EventKey == row.EventKey {
			return nil
		}
	}
	r.dns = append(r.dns, row)
	return nil
}

func (r *evidenceMemoryRepo) ActiveDNSForDestination(_ context.Context, client string, nodeID, inboundID int, ip string, at int64) ([]DNSObservation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var rows []DNSObservation
	for _, row := range r.dns {
		if row.ClientEmail == client && row.NodeID == nodeID && row.InboundID == inboundID && row.ResolvedIP == ip && row.ObservedAt <= at && (row.ExpiresAt == 0 || row.ExpiresAt > at) {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func (r *evidenceMemoryRepo) RecordEvidence(_ context.Context, row EvidenceObservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.evidence {
		if existing.EventKey == row.EventKey {
			return nil
		}
	}
	r.evidence = append(r.evidence, row)
	return nil
}

func TestEvidenceDNSCorrelationTTLAndRestart(t *testing.T) {
	repo := &evidenceMemoryRepo{}
	service := NewEvidenceService(repo, true)
	ctx := context.Background()
	at := time.UnixMilli(10_000)
	if err := service.ObserveDNS(ctx, DNSInput{ObservedAt: at, ClientEmail: "alice", Domain: "Example.COM.", ResolvedIP: "192.0.2.10", TTL: time.Second, Source: SourceDNSObserver}); err != nil {
		t.Fatalf("observe DNS: %v", err)
	}
	first, err := service.Correlate(ctx, DestinationInput{ObservedAt: at.Add(500 * time.Millisecond), ClientEmail: "alice", DestinationIP: "192.0.2.10", SessionKey: "s1", Source: SourceAccessLog, EventKey: "connection-1"})
	if err != nil || len(first.Candidates) != 1 || first.Candidates[0] != "example.com" || !first.Items[0].Selected {
		t.Fatalf("active correlation = %+v, %v", first, err)
	}
	// A new service instance sees the same durable repository state; no cache is
	// required for restart/resume behavior.
	restarted, err := NewEvidenceService(repo, true).Correlate(ctx, DestinationInput{ObservedAt: at.Add(2 * time.Second), ClientEmail: "alice", DestinationIP: "192.0.2.10", SessionKey: "s2", Source: SourceAccessLog, EventKey: "connection-2"})
	if err != nil || len(restarted.Candidates) != 0 || len(restarted.Items) != 1 || restarted.Items[0].Kind != EvidenceIPOnly {
		t.Fatalf("expired correlation after restart = %+v, %v", restarted, err)
	}
}

func TestEvidenceAmbiguityConflictAndSNIPrecedence(t *testing.T) {
	repo := &evidenceMemoryRepo{}
	service := NewEvidenceService(repo, true)
	at := time.UnixMilli(20_000)
	for _, domain := range []string{"one.example", "two.example"} {
		if err := service.ObserveDNS(context.Background(), DNSInput{ObservedAt: at, ClientEmail: "alice", Domain: domain, ResolvedIP: "192.0.2.20", TTL: time.Hour, Source: SourceDNSObserver}); err != nil {
			t.Fatalf("observe DNS: %v", err)
		}
	}
	result, err := service.Correlate(context.Background(), DestinationInput{ObservedAt: at.Add(time.Minute), ClientEmail: "alice", DestinationIP: "192.0.2.20", SNI: "visible.example", Source: SourceSNI, EventKey: "connection-ambiguous"})
	if err != nil || !result.Ambiguous || len(result.Candidates) != 3 {
		t.Fatalf("ambiguous result = %+v, %v", result, err)
	}
	var selectedSNI, conflictingDNS bool
	for _, item := range result.Items {
		if item.Kind == EvidenceSNIDomain && item.Selected && item.Provenance == ProvenanceObserved {
			selectedSNI = item.Source == SourceSNI
		}
		if item.Kind == EvidenceDNSDomain && item.Conflicting {
			conflictingDNS = true
		}
	}
	if !selectedSNI || !conflictingDNS {
		t.Fatalf("SNI precedence/provenance missing: %+v", result.Items)
	}
}

func TestEvidenceClientIsolationDeduplicationAndDisabledNoOp(t *testing.T) {
	repo := &evidenceMemoryRepo{}
	service := NewEvidenceService(repo, true)
	at := time.UnixMilli(30_000)
	if err := service.ObserveDNS(context.Background(), DNSInput{ObservedAt: at, ClientEmail: "alice", Domain: "private.example", ResolvedIP: "192.0.2.30", TTL: time.Hour, Source: SourceDNSObserver, EventKey: "dns-1"}); err != nil {
		t.Fatalf("observe DNS: %v", err)
	}
	isolated, err := service.Correlate(context.Background(), DestinationInput{ObservedAt: at.Add(time.Second), ClientEmail: "bob", DestinationIP: "192.0.2.30", Source: SourceAccessLog, EventKey: "bob-1"})
	if err != nil || len(isolated.Candidates) != 0 || len(isolated.Items) != 1 || isolated.Items[0].Kind != EvidenceIPOnly {
		t.Fatalf("client isolation = %+v, %v", isolated, err)
	}
	first, err := service.Correlate(context.Background(), DestinationInput{ObservedAt: at.Add(time.Second), ClientEmail: "alice", DestinationIP: "192.0.2.30", Source: SourceAccessLog, EventKey: "alice-1"})
	if err != nil {
		t.Fatalf("first correlation: %v", err)
	}
	second, err := service.Correlate(context.Background(), DestinationInput{ObservedAt: at.Add(time.Second), ClientEmail: "alice", DestinationIP: "192.0.2.30", Source: SourceAccessLog, EventKey: "alice-1"})
	if err != nil || len(first.Items) != len(second.Items) || len(repo.evidence) != len(first.Items)+1 {
		t.Fatalf("deduplication result = first=%+v second=%+v stored=%d err=%v", first, second, len(repo.evidence), err)
	}
	disabled := NewEvidenceService(repo, false)
	if err := disabled.ObserveDNS(context.Background(), DNSInput{ObservedAt: at, ClientEmail: "leak", Domain: "should-not-store.example", ResolvedIP: "192.0.2.31", TTL: time.Hour, Source: SourceDNSObserver}); err != nil {
		t.Fatalf("disabled observe: %v", err)
	}
	if len(repo.dns) != 1 {
		t.Fatalf("disabled service changed durable DNS state: %d", len(repo.dns))
	}
}

func TestEvidenceCorrelationConcurrent(t *testing.T) {
	repo := &evidenceMemoryRepo{}
	service := NewEvidenceService(repo, true)
	at := time.UnixMilli(40_000)
	if err := service.ObserveDNS(context.Background(), DNSInput{ObservedAt: at, ClientEmail: "alice", Domain: "parallel.example", ResolvedIP: "192.0.2.40", TTL: time.Hour, Source: SourceDNSObserver}); err != nil {
		t.Fatalf("observe DNS: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = service.Correlate(context.Background(), DestinationInput{ObservedAt: at.Add(time.Duration(i) * time.Millisecond), ClientEmail: "alice", DestinationIP: "192.0.2.40", Source: SourceAccessLog, EventKey: "parallel" + string(rune(i))})
		}(i)
	}
	wg.Wait()
}
