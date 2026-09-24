package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"
)

const (
	EvidenceDirectDomain = "direct_domain"
	EvidenceDNSDomain    = "dns_derived_domain"
	EvidenceSNIDomain    = "sni_derived_domain"
	EvidenceIPOnly       = "ip_only"
	EvidenceInferred     = "inferred_correlation"

	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// DNSInput is the safe observer contract. Observers provide DNS metadata only;
// the service never performs DNS lookups and never receives payload content.
type DNSInput struct {
	ObservedAt  time.Time
	ClientEmail string
	NodeID      int
	InboundID   int
	Domain      string
	RecordType  string
	ResolvedIP  string
	TTL         time.Duration
	Source      string
	EventKey    string
}

// DestinationInput is metadata visible at a connection boundary. DirectDomain
// and SNI are optional and are never synthesized by this package.
type DestinationInput struct {
	ObservedAt    time.Time
	ClientEmail   string
	NodeID        int
	InboundID     int
	DestinationIP string
	DirectDomain  string
	SNI           string
	SessionKey    string
	Source        string
	EventKey      string
}

type EvidenceResult struct {
	Items      []EvidenceObservation
	Candidates []string
	Ambiguous  bool
}

// EvidenceService correlates observations through the existing analytics
// repository. It has no in-memory evidence cache, making restart behavior
// identical to uninterrupted operation and keeping client boundaries explicit.
type EvidenceService struct {
	repo    Repository
	enabled bool
}

func NewEvidenceService(repo Repository, enabled bool) *EvidenceService {
	if repo == nil {
		repo = NoopRepository{}
	}
	return &EvidenceService{repo: repo, enabled: enabled}
}

func (s *EvidenceService) ObserveDNS(ctx context.Context, input DNSInput) error {
	if s == nil || !s.enabled {
		return nil
	}
	if input.ObservedAt.IsZero() || input.Domain == "" || input.ResolvedIP == "" || input.Source == "" || input.TTL < 0 {
		return errors.New("invalid DNS input")
	}
	domain := normalizeDomain(input.Domain)
	ip := normalizeIP(input.ResolvedIP)
	if domain == "" || ip == "" {
		return errors.New("invalid DNS domain or IP")
	}
	eventKey := input.EventKey
	if eventKey == "" {
		eventKey = digestEvidence(strings.Join([]string{input.ClientEmail, domain, ip, strings.ToUpper(strings.TrimSpace(input.RecordType)), input.ObservedAt.UTC().Format(time.RFC3339Nano), input.Source}, "|"))
	}
	observed := input.ObservedAt.UnixMilli()
	// A zero DNS TTL expires at the observation instant; it must never become
	// a durable, indefinitely reusable correlation hint.
	expires := input.ObservedAt.Add(input.TTL).UnixMilli()
	return s.repo.RecordDNS(ctx, DNSObservation{
		ObservedAt: observed, ClientEmail: input.ClientEmail, NodeID: input.NodeID, InboundID: input.InboundID,
		Domain: domain, RecordType: strings.ToUpper(strings.TrimSpace(input.RecordType)), ResolvedIP: ip,
		Source: input.Source, Provenance: ProvenanceObserved, Confidence: 1, ExpiresAt: expires, EventKey: eventKey,
	})
}

func (s *EvidenceService) Correlate(ctx context.Context, input DestinationInput) (EvidenceResult, error) {
	result := EvidenceResult{Items: []EvidenceObservation{}}
	if s == nil || !s.enabled {
		return result, nil
	}
	if input.ObservedAt.IsZero() || input.Source == "" {
		return result, errors.New("invalid destination input")
	}
	at := input.ObservedAt.UnixMilli()
	ip := normalizeIP(input.DestinationIP)
	direct := normalizeDomain(input.DirectDomain)
	sni := normalizeDomain(input.SNI)
	if ip == "" && direct == "" && sni == "" {
		return result, errors.New("destination input has no safe metadata")
	}

	candidates := map[string]struct{}{}
	dnsSources := map[string]map[string]struct{}{}
	if direct != "" {
		candidates[direct] = struct{}{}
	}
	if sni != "" {
		candidates[sni] = struct{}{}
	}
	dnsRows := []DNSObservation{}
	var err error
	if ip != "" {
		dnsRows, err = s.repo.ActiveDNSForDestination(ctx, input.ClientEmail, input.NodeID, input.InboundID, ip, at)
		if err != nil {
			return result, err
		}
	}
	for _, row := range dnsRows {
		if domain := normalizeDomain(row.Domain); domain != "" {
			candidates[domain] = struct{}{}
			if _, exists := dnsSources[domain]; !exists {
				dnsSources[domain] = map[string]struct{}{}
			}
			if row.Source != "" {
				dnsSources[domain][row.Source] = struct{}{}
			}
		}
	}
	result.Candidates = sortedStrings(candidates)
	result.Ambiguous = len(result.Candidates) > 1 && direct == ""

	items := make([]EvidenceObservation, 0, len(result.Candidates)+1)
	if direct != "" {
		items = append(items, evidenceItem(input, direct, ip, input.Source, EvidenceDirectDomain, ProvenanceObserved, 1, ConfidenceHigh, true, false, sni != "" && sni != direct))
	}
	if sni != "" {
		selected := direct == ""
		conflict := direct != "" && direct != sni
		items = append(items, evidenceItem(input, sni, ip, SourceSNI, EvidenceSNIDomain, ProvenanceObserved, .95, ConfidenceHigh, selected, result.Ambiguous, conflict))
	}
	for _, domain := range result.Candidates {
		selected := direct == "" && sni == "" && len(result.Candidates) == 1
		sources := sortedStrings(dnsSources[domain])
		if len(sources) == 0 {
			sources = []string{SourceDNSObserver}
		}
		for _, source := range sources {
			items = append(items, evidenceItem(input, domain, ip, source, EvidenceDNSDomain, ProvenanceCorrelated, .75, confidenceLevel(.75), selected, result.Ambiguous, (direct != "" && direct != domain) || (sni != "" && sni != domain)))
			if selected || result.Ambiguous {
				items = append(items, evidenceItem(input, domain, ip, source, EvidenceInferred, ProvenanceInferred, .55, ConfidenceMedium, selected, result.Ambiguous, false))
			}
		}
	}
	if len(items) == 0 && ip != "" {
		items = append(items, evidenceItem(input, "", ip, input.Source, EvidenceIPOnly, ProvenanceObserved, .35, ConfidenceLow, true, false, false))
	}
	for _, item := range items {
		if err := s.repo.RecordEvidence(ctx, item); err != nil {
			return EvidenceResult{}, err
		}
	}
	result.Items = items
	return result, nil
}

func evidenceItem(input DestinationInput, domain, ip, source, kind, provenance string, confidence float64, level string, selected, ambiguous, conflicting bool) EvidenceObservation {
	key := strings.Join([]string{input.EventKey, input.ClientEmail, input.SessionKey, domain, ip, kind}, "|")
	return EvidenceObservation{ObservedAt: input.ObservedAt.UnixMilli(), ClientEmail: input.ClientEmail, NodeID: input.NodeID, InboundID: input.InboundID, SessionKey: input.SessionKey, Domain: domain, DestinationIP: ip, SNI: normalizeDomain(input.SNI), Kind: kind, Source: source, Provenance: provenance, Confidence: confidence, Level: level, Selected: selected, Ambiguous: ambiguous, Conflicting: conflicting, EventKey: digestEvidence(key)}
}

func confidenceLevel(score float64) string {
	switch {
	case score >= .9:
		return ConfidenceHigh
	case score >= .6:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}

func normalizeDomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(value, ".")))
	if value == "" || strings.ContainsAny(value, "/\\ @") {
		return ""
	}
	return value
}

func normalizeIP(value string) string {
	value = strings.TrimSpace(value)
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	if net.ParseIP(value) == nil {
		return ""
	}
	return value
}

func sortedStrings(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func digestEvidence(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
