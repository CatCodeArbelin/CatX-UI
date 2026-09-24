package analytics

// Enrichment is deliberately metadata-only. It consumes normalized domains,
// visible SNI/DNS evidence, and optional IP registry metadata; it never opens
// a connection to a destination and never inspects payloads.

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	SourceDomainCatalog          = "domain_catalog"
	SourceIPMetadata             = "ip_metadata"
	SourceClassificationConflict = "classification_conflict"
	ProvenanceFirstParty         = "first_party"
	ProvenanceDerived            = "derived"
	ProvenanceUnknown            = "unknown"
)

type IPMetadata struct {
	ASN        uint32
	Country    string
	Service    string
	Category   string
	Source     string
	Confidence float64
	TTL        time.Duration
}

// IPMetadataProvider is an optional, bounded metadata provider. Implementations
// must return registry metadata only; the service does not perform DNS or TLS.
type IPMetadataProvider interface {
	Lookup(context.Context, net.IP) (IPMetadata, error)
}

type EnrichmentInput struct {
	ObservedAt                 time.Time
	Domain, DestinationIP, SNI string
	Evidence                   []EvidenceObservation
}

type ClassificationCandidate struct {
	Service    string  `json:"service,omitempty"`
	Category   string  `json:"category"`
	Source     string  `json:"source"`
	Provenance string  `json:"provenance"`
	Confidence float64 `json:"confidence"`
	Level      string  `json:"level"`
	FirstParty bool    `json:"firstParty"`
	Reason     string  `json:"reason"`
}

type EnrichmentResult struct {
	Service, Category, Source, Provenance, Level string
	ASN                                          uint32
	Country                                      string
	Confidence                                   float64
	FirstParty, Conflict                         bool
	CandidatesJSON, Reason                       string
	Candidates                                   []ClassificationCandidate
}

type domainRule struct{ Domain, Service, Category string }

// Rules are intentionally small and explicit. A match is valid only for the
// exact domain or a dot-delimited child, never for an arbitrary string suffix.
var domainRules = []domainRule{
	{"facebook.com", "Facebook", CategorySocial}, {"instagram.com", "Instagram", CategorySocial}, {"twitter.com", "X", CategorySocial}, {"x.com", "X", CategorySocial},
	{"youtube.com", "YouTube", CategoryVideo}, {"netflix.com", "Netflix", CategoryVideo}, {"twitch.tv", "Twitch", CategoryVideo},
	{"telegram.org", "Telegram", CategoryMessaging}, {"whatsapp.com", "WhatsApp", CategoryMessaging}, {"signal.org", "Signal", CategoryMessaging},
	{"steampowered.com", "Steam", CategoryGaming}, {"epicgames.com", "Epic Games", CategoryGaming},
	{"google.com", "Google", CategorySearch}, {"bing.com", "Bing", CategorySearch},
	{"microsoft.com", "Microsoft", CategorySoftware}, {"windowsupdate.com", "Windows Update", CategorySoftware}, {"apple.com", "Apple", CategorySoftware},
	{"cloudflare.com", "Cloudflare", CategoryCloudCDN}, {"cloudfront.net", "Amazon CloudFront", CategoryCloudCDN}, {"fastly.net", "Fastly", CategoryCloudCDN},
	{"doubleclick.net", "Google Ads", CategoryAdvertising}, {"googlesyndication.com", "Google Ads", CategoryAdvertising},
	{"pornhub.com", "Pornhub", CategoryAdult}, {"xvideos.com", "XVideos", CategoryAdult},
	{"bet365.com", "Bet365", CategoryGambling}, {"paddypower.com", "Paddy Power", CategoryGambling},
}

type cacheEntry struct {
	metadata IPMetadata
	expires  time.Time
}

type Enricher struct {
	provider           IPMetadataProvider
	mu                 sync.Mutex
	cache              map[string]cacheEntry
	maxEntries         int
	defaultTTL, maxTTL time.Duration
	refresh            chan struct{}
}

var globalEnrichment struct {
	sync.RWMutex
	enricher *Enricher
	enabled  bool
}

func ConfigureEnrichment(enricher *Enricher, enabled bool) {
	globalEnrichment.Lock()
	defer globalEnrichment.Unlock()
	if enricher == nil {
		enricher = NewEnricher(nil)
	}
	globalEnrichment.enricher, globalEnrichment.enabled = enricher, enabled
}

func CurrentEnricher() *Enricher {
	globalEnrichment.RLock()
	defer globalEnrichment.RUnlock()
	if !globalEnrichment.enabled || globalEnrichment.enricher == nil {
		return nil
	}
	return globalEnrichment.enricher
}

func NewEnricher(provider IPMetadataProvider) *Enricher {
	return &Enricher{provider: provider, cache: make(map[string]cacheEntry), maxEntries: 1024, defaultTTL: time.Hour, maxTTL: 24 * time.Hour, refresh: make(chan struct{}, 4)}
}

func (e *Enricher) Enrich(ctx context.Context, input EnrichmentInput) EnrichmentResult {
	if e == nil {
		return unknownEnrichment()
	}
	domain := normalizeDomain(input.Domain)
	if domain == "" {
		domain = normalizeDomain(input.SNI)
	}
	result := unknownEnrichment()
	if ip := normalizeIP(input.DestinationIP); ip != "" {
		result = e.applyIP(ctx, result, ip, domain != "")
	}
	candidates := classifyDomain(domain, input.Evidence)
	if len(result.Candidates) > 0 {
		candidates = append(candidates, result.Candidates...)
		result.Candidates = nil
	}
	if len(candidates) > 0 {
		result = chooseClassification(candidates, result)
	}
	if result.Category == "" {
		result.Category = CategoryUnknown
	}
	if result.Source == "" {
		result.Source = SourceClassificationConflict
		result.Provenance = ProvenanceUnknown
		result.Level = ConfidenceLow
		result.Reason = "no safe domain or provider classification"
	}
	return result
}

func classifyDomain(domain string, evidence []EvidenceObservation) []ClassificationCandidate {
	if domain == "" {
		return nil
	}
	direct := false
	for _, item := range evidence {
		if item.Domain == domain && (item.Kind == EvidenceDirectDomain || item.Kind == EvidenceSNIDomain) && item.Selected {
			direct = true
		}
	}
	var out []ClassificationCandidate
	for _, rule := range domainRules {
		if domain != rule.Domain && !strings.HasSuffix(domain, "."+rule.Domain) {
			continue
		}
		exact := domain == rule.Domain
		score := 0.78
		level := ConfidenceMedium
		if exact {
			score, level = .95, ConfidenceHigh
		}
		prov := ProvenanceDerived
		first := false
		if direct {
			prov, first = ProvenanceFirstParty, true
		} else if !exact {
			score = .68
		}
		for _, item := range evidence {
			if item.Domain == domain && item.Confidence > 0 && item.Confidence < score {
				score = item.Confidence
			}
		}
		if score >= .9 {
			level = ConfidenceHigh
		} else if score >= .6 {
			level = ConfidenceMedium
		} else {
			level = ConfidenceLow
		}
		out = append(out, ClassificationCandidate{Service: rule.Service, Category: rule.Category, Source: SourceDomainCatalog, Provenance: prov, Confidence: score, Level: level, FirstParty: first, Reason: reasonForDomain(exact, direct)})
	}
	return out
}

func reasonForDomain(exact, direct bool) string {
	if direct && exact {
		return "exact domain with direct/SNI evidence"
	}
	if exact {
		return "exact domain match"
	}
	if direct {
		return "safe dot-delimited subdomain with direct/SNI evidence"
	}
	return "safe dot-delimited subdomain inheritance"
}

func chooseClassification(candidates []ClassificationCandidate, base EnrichmentResult) EnrichmentResult {
	if len(candidates) == 0 {
		return base
	}
	unique := map[string]bool{}
	for _, c := range candidates {
		unique[c.Category] = true
	}
	if len(unique) > 1 {
		base.Category, base.Service, base.Source, base.Provenance, base.Level = CategoryUnknown, "", SourceClassificationConflict, ProvenanceUnknown, ConfidenceLow
		base.Confidence, base.Conflict, base.Reason = .35, true, "domain evidence maps to conflicting categories"
		base.Candidates = candidates
		base.CandidatesJSON, _ = jsonString(candidates)
		return base
	}
	chosen := candidates[0]
	base.Service, base.Category, base.Source, base.Provenance, base.Confidence, base.Level, base.FirstParty, base.Reason = chosen.Service, chosen.Category, chosen.Source, chosen.Provenance, chosen.Confidence, chosen.Level, chosen.FirstParty, chosen.Reason
	base.Candidates = candidates
	base.CandidatesJSON, _ = jsonString(candidates)
	return base
}

func (e *Enricher) applyIP(ctx context.Context, result EnrichmentResult, raw string, hasDomain bool) EnrichmentResult {
	if e.provider == nil {
		return result
	}
	now := time.Now()
	e.mu.Lock()
	entry, ok := e.cache[raw]
	e.mu.Unlock()
	var metadata IPMetadata
	if ok && now.Before(entry.expires) {
		metadata = entry.metadata
	} else {
		select {
		case e.refresh <- struct{}{}:
			defer func() { <-e.refresh }()
		default:
			return result
		}
		var err error
		metadata, err = e.provider.Lookup(ctx, net.ParseIP(raw))
		if err != nil {
			return result
		}
		ttl := metadata.TTL
		if ttl <= 0 {
			ttl = e.defaultTTL
		}
		if ttl > e.maxTTL {
			ttl = e.maxTTL
		}
		e.mu.Lock()
		if len(e.cache) >= e.maxEntries {
			for key := range e.cache {
				delete(e.cache, key)
				break
			}
		}
		e.cache[raw] = cacheEntry{metadata: metadata, expires: now.Add(ttl)}
		e.mu.Unlock()
	}
	if metadata.Source == "" {
		metadata.Source = SourceIPMetadata
	}
	result.ASN, result.Country = metadata.ASN, strings.ToUpper(strings.TrimSpace(metadata.Country))
	if metadata.Service == "" || metadata.Category == "" {
		result.Source, result.Provenance, result.Reason = metadata.Source, ProvenanceUnknown, "IP metadata has no safe service/category classification"
		return result
	}
	// IP-only service claims are never definitive for shared/CDN addresses.
	if !hasDomain {
		result.Category, result.Source, result.Provenance, result.Confidence, result.Level, result.Reason = CategoryUnknown, metadata.Source, ProvenanceUnknown, min(.35, metadata.Confidence), ConfidenceLow, "IP metadata lacks supporting domain/SNI/DNS evidence"
		return result
	}
	// A provider claim is retained as a separate candidate so disagreement is
	// explainable. Infrastructure categories are not allowed to override a
	// supported domain identity.
	if metadata.Category != CategoryCloudCDN {
		result.Candidates = []ClassificationCandidate{{Service: metadata.Service, Category: metadata.Category, Source: metadata.Source, Provenance: ProvenanceDerived, Confidence: min(.65, metadata.Confidence), Level: ConfidenceMedium, Reason: "IP metadata supported by domain/SNI evidence"}}
	}
	return result
}

func unknownEnrichment() EnrichmentResult {
	return EnrichmentResult{Category: CategoryUnknown, Source: SourceClassificationConflict, Provenance: ProvenanceUnknown, Confidence: 0, Level: ConfidenceLow, Reason: "unknown"}
}
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func jsonString(v any) (string, error) { b, err := json.Marshal(v); return string(b), err }
