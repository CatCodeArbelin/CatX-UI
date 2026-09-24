package analytics

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

type testIPProvider struct {
	mu       sync.Mutex
	calls    int
	metadata IPMetadata
}

func (p *testIPProvider) Lookup(_ context.Context, _ net.IP) (IPMetadata, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return p.metadata, nil
}

func (p *testIPProvider) Calls() int { p.mu.Lock(); defer p.mu.Unlock(); return p.calls }

func TestEnrichmentExactDomainAndSafeSubdomainInheritance(t *testing.T) {
	e := NewEnricher(nil)
	exact := e.Enrich(context.Background(), EnrichmentInput{Domain: "youtube.com", Evidence: []EvidenceObservation{{Domain: "youtube.com", Kind: EvidenceDirectDomain, Selected: true}}})
	if exact.Service != "YouTube" || exact.Category != CategoryVideo || exact.Level != ConfidenceHigh || !exact.FirstParty || exact.Provenance != ProvenanceFirstParty {
		t.Fatalf("exact = %+v", exact)
	}
	sub := e.Enrich(context.Background(), EnrichmentInput{Domain: "img.youtube.com"})
	if sub.Service != "YouTube" || sub.Category != CategoryVideo || sub.Level != ConfidenceMedium || sub.FirstParty {
		t.Fatalf("subdomain = %+v", sub)
	}
	unsafe := e.Enrich(context.Background(), EnrichmentInput{Domain: "notyoutube.com"})
	if unsafe.Category != CategoryUnknown {
		t.Fatalf("unsafe suffix classified = %+v", unsafe)
	}
}

func TestEnrichmentConflictRemainsUnknownAndExplainable(t *testing.T) {
	p := &testIPProvider{metadata: IPMetadata{ASN: 64500, Country: "us", Service: "Other Search", Category: CategorySearch, Source: "provider-b", Confidence: .9, TTL: time.Hour}}
	e := NewEnricher(p)
	result := e.Enrich(context.Background(), EnrichmentInput{Domain: "youtube.com", DestinationIP: "192.0.2.10", Evidence: []EvidenceObservation{{Domain: "youtube.com", Kind: EvidenceDirectDomain, Selected: true}}})
	if !result.Conflict || result.Category != CategoryUnknown || result.Confidence >= .6 || result.Provenance != ProvenanceUnknown {
		t.Fatalf("conflict = %+v", result)
	}
	if result.CandidatesJSON == "" || result.Reason == "" {
		t.Fatalf("conflict explanation missing = %+v", result)
	}
}

func TestEnrichmentSharedIPDoesNotIdentifyService(t *testing.T) {
	p := &testIPProvider{metadata: IPMetadata{ASN: 13335, Country: "US", Service: "Cloudflare customer", Category: CategorySocial, Source: SourceIPMetadata, Confidence: 1, TTL: time.Hour}}
	result := NewEnricher(p).Enrich(context.Background(), EnrichmentInput{DestinationIP: "192.0.2.20"})
	if result.Service != "" || result.Category != CategoryUnknown || result.Confidence > .35 || result.ASN != 13335 || result.Country != "US" {
		t.Fatalf("shared IP = %+v", result)
	}
}

func TestEnrichmentCacheExpiryAndBoundedRefresh(t *testing.T) {
	p := &testIPProvider{metadata: IPMetadata{ASN: 64501, Country: "DE", Service: "Provider", Category: CategoryCloudCDN, Source: SourceIPMetadata, Confidence: .8, TTL: 10 * time.Millisecond}}
	e := NewEnricher(p)
	input := EnrichmentInput{DestinationIP: "192.0.2.30"}
	_ = e.Enrich(context.Background(), input)
	_ = e.Enrich(context.Background(), input)
	if p.Calls() != 1 {
		t.Fatalf("cache calls = %d", p.Calls())
	}
	time.Sleep(20 * time.Millisecond)
	_ = e.Enrich(context.Background(), input)
	if p.Calls() != 2 {
		t.Fatalf("refresh calls = %d", p.Calls())
	}
}

func TestEnrichmentConcurrentCacheAccess(t *testing.T) {
	p := &testIPProvider{metadata: IPMetadata{ASN: 64502, Country: "NL", TTL: time.Hour}}
	e := NewEnricher(p)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = e.Enrich(context.Background(), EnrichmentInput{DestinationIP: "192.0.2.40"})
		}()
	}
	wg.Wait()
	if p.Calls() == 0 || p.Calls() > 4 {
		t.Fatalf("bounded concurrent refresh calls = %d", p.Calls())
	}
}

func TestEnrichmentUnknownFallbackAndDisabledGlobalNoOp(t *testing.T) {
	result := NewEnricher(nil).Enrich(context.Background(), EnrichmentInput{Domain: "unlisted.example", DestinationIP: "not-an-ip"})
	if result.Category != CategoryUnknown || result.Level != ConfidenceLow || result.Reason == "" {
		t.Fatalf("unknown = %+v", result)
	}
	ConfigureEnrichment(nil, false)
	if got := CurrentEnricher().Enrich(context.Background(), EnrichmentInput{Domain: "youtube.com"}); got.Category != CategoryUnknown {
		t.Fatalf("disabled enrichment = %+v", got)
	}
}
