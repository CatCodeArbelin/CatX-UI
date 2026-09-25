package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

// ResolvedCandidate is the inspectable precedence item used by the decision
// engine. Its fields intentionally retain the source and target that won.
type ResolvedCandidate struct {
	Source     string `json:"source"`
	PolicyID   uint   `json:"policyId"`
	TargetType string `json:"targetType"`
	TargetRef  string `json:"targetRef"`
	Priority   int    `json:"priority"`
	CreatedAt  int64  `json:"createdAt"`
	ID         uint   `json:"id"`
	Active     bool   `json:"active"`
	Scope      string `json:"scope,omitempty"`
	StartsAt   int64  `json:"startsAt,omitempty"`
	ExpiresAt  int64  `json:"expiresAt,omitempty"`
}

type CandidateExplanation struct {
	Candidate ResolvedCandidate `json:"candidate"`
	Won       bool              `json:"won"`
	Reason    string            `json:"reason"`
}

type Decision struct {
	ClientEmail         string                 `json:"clientEmail"`
	GroupName           string                 `json:"groupName,omitempty"`
	PolicyID            uint                   `json:"policyId"`
	PolicyName          string                 `json:"policyName"`
	Action              string                 `json:"action"`
	Services            []string               `json:"services,omitempty"`
	Categories          []string               `json:"categories,omitempty"`
	Destinations        []string               `json:"destinations,omitempty"`
	Winner              ResolvedCandidate      `json:"winner"`
	Candidates          []ResolvedCandidate    `json:"candidates"`
	Explanations        []CandidateExplanation `json:"explanations"`
	Explanation         string                 `json:"explanation"`
	OverrideScope       string                 `json:"overrideScope,omitempty"`
	Quarantined         bool                   `json:"quarantined,omitempty"`
	QuarantineReleased  bool                   `json:"quarantineReleased,omitempty"`
	QuarantineAllowlist []string               `json:"quarantineAllowlist,omitempty"`
	ManagedDNS          bool                   `json:"managedDns,omitempty"`
	DNSOutboundTag      string                 `json:"dnsOutboundTag,omitempty"`
	SafeSearch          bool                   `json:"safeSearch,omitempty"`
	DNSLimitations      []string               `json:"dnsLimitations,omitempty"`
}

// ResolveDecisions resolves one deterministic decision per known client. It
// is the repository-facing layer; Xray rule emission consumes only Decisions.
func (r *Repository) ResolveDecisions(ctx context.Context, emails []string, at int64) ([]Decision, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("policy database is unavailable")
	}
	clients := make(map[string]model.ClientRecord, len(emails))
	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		var c model.ClientRecord
		if err := r.db.WithContext(ctx).Where("email = ?", email).First(&c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		clients[email] = c
	}
	keys := make([]string, 0, len(clients))
	for email := range clients {
		keys = append(keys, email)
	}
	sort.Strings(keys)
	out := make([]Decision, 0, len(keys))
	for _, email := range keys {
		c := clients[email]
		decision, err := r.ResolveDecision(ctx, c.Email, c.Group, at)
		if err != nil {
			return nil, err
		}
		if decision == nil {
			continue
		}
		out = append(out, *decision)
	}
	return out, nil
}

// ResolveDecision is shared by live config decoration and hypothetical
// simulation. It performs no writes and accepts a client/group pair directly,
// allowing simulation of clients that are not persisted yet.
func (r *Repository) ResolveDecision(ctx context.Context, clientEmail, groupName string, at int64) (*Decision, error) {
	resolved, err := r.Resolve(ctx, clientEmail, groupName, at)
	if err != nil {
		return nil, err
	}
	if len(resolved.Items) == 0 {
		return nil, nil
	}
	winner := resolved.Items[0]
	var p Policy
	if err := r.db.WithContext(ctx).First(&p, winner.PolicyID).Error; err != nil {
		return nil, err
	}
	var spec Definition
	if err := json.Unmarshal([]byte(p.Spec), &spec); err != nil {
		return nil, fmt.Errorf("policy %d: malformed spec: %w", p.ID, err)
	}
	if winner.Source == "override" || winner.Source == "temporary" {
		var value string
		if winner.Source == "override" {
			var row PolicyOverride
			if err := r.db.WithContext(ctx).First(&row, winner.ID).Error; err != nil {
				return nil, err
			}
			value = row.Value
		} else {
			var row TemporaryOverride
			if err := r.db.WithContext(ctx).First(&row, winner.ID).Error; err != nil {
				return nil, err
			}
			value = row.Value
		}
		var patch struct {
			Action              string         `json:"action"`
			Services            []string       `json:"services"`
			Categories          []string       `json:"categories"`
			Destinations        []string       `json:"destinations"`
			Domains             []string       `json:"domains"`
			Quarantine          *bool          `json:"quarantine"`
			QuarantineAllowlist []string       `json:"quarantineAllowlist"`
			DNS                 map[string]any `json:"dns"`
		}
		if err := json.Unmarshal([]byte(value), &patch); err != nil {
			return nil, fmt.Errorf("policy %d: malformed override: %w", p.ID, err)
		}
		switch winner.Scope {
		case ScopePolicy:
			if patch.Action != "" {
				spec.Action = patch.Action
			}
			if patch.Services != nil {
				spec.Services = patch.Services
			}
			if patch.Categories != nil {
				spec.Categories = patch.Categories
			}
			if patch.Destinations != nil {
				spec.Destinations = patch.Destinations
			}
		case ScopeDomain:
			if patch.Destinations != nil {
				spec.Destinations = patch.Destinations
			} else {
				spec.Destinations = patch.Domains
			}
		case ScopeService:
			if patch.Services != nil {
				spec.Services = patch.Services
			}
		case ScopeDNS:
			if patch.DNS != nil {
				spec.DNS = patch.DNS
			}
		}
		if winner.Scope == ScopePolicy {
			if patch.Quarantine != nil {
				spec.Quarantine = *patch.Quarantine
			}
			if patch.QuarantineAllowlist != nil {
				spec.QuarantineAllowlist = patch.QuarantineAllowlist
			}
		}
	}
	quarantined, released, allowlist, managedDNS, dnsOutboundTag, safeSearch, limitations, err := r.resolveCapabilities(ctx, resolved.Items)
	if err != nil {
		return nil, err
	}
	if spec.Action == "" && !quarantined && !managedDNS && !safeSearch {
		return nil, nil
	}
	if spec.Action == "" {
		spec.Action = "deny"
	}
	if spec.Action != "allow" && spec.Action != "deny" {
		return nil, fmt.Errorf("policy %d: unsupported action %q", p.ID, spec.Action)
	}
	explanations := make([]CandidateExplanation, 0, len(resolved.Items))
	for i, candidate := range resolved.Items {
		reason := "lost to a higher-precedence candidate"
		if i == 0 {
			reason = "selected by deterministic precedence ordering"
		}
		explanations = append(explanations, CandidateExplanation{Candidate: candidate, Won: i == 0, Reason: reason})
	}
	explanation := fmt.Sprintf("%s %s target %s by %s", winner.Source, winner.TargetType, winner.TargetRef, p.Name)
	if quarantined {
		explanation += "; quarantine is effective and generic policy overrides cannot bypass it"
	}
	if released {
		explanation += "; bounded quarantine-release is effective"
	}
	return &Decision{ClientEmail: clientEmail, GroupName: groupName, PolicyID: p.ID, PolicyName: p.Name, Action: spec.Action, Services: spec.Services, Categories: spec.Categories, Destinations: spec.Destinations, Winner: winner, Candidates: resolved.Items, Explanations: explanations, OverrideScope: winner.Scope, Explanation: explanation, Quarantined: quarantined, QuarantineReleased: released, QuarantineAllowlist: allowlist, ManagedDNS: managedDNS, DNSOutboundTag: dnsOutboundTag, SafeSearch: safeSearch, DNSLimitations: limitations}, nil
}

// resolveCapabilities intentionally runs alongside ordinary winner selection.
// Quarantine is a safety state: generic allow/domain/service/category rules do
// not clear it. Only an active temporary override with the dedicated
// quarantine-release scope can release it.
func (r *Repository) resolveCapabilities(ctx context.Context, candidates []resolvedCandidate) (bool, bool, []string, bool, string, bool, []string, error) {
	quarantined, released, managedDNS, safeSearch := false, false, false, false
	dnsOutboundTag := ""
	dnsSelected := false
	allow := map[string]struct{}{}
	limitations := []string{}
	for _, candidate := range candidates {
		var p Policy
		if err := r.db.WithContext(ctx).First(&p, candidate.PolicyID).Error; err != nil {
			return false, false, nil, false, "", false, nil, err
		}
		var d Definition
		if err := json.Unmarshal([]byte(p.Spec), &d); err != nil {
			return false, false, nil, false, "", false, nil, err
		}
		if candidate.Scope == ScopeDNS && (candidate.Source == "override" || candidate.Source == "temporary") {
			var value string
			if candidate.Source == "override" {
				var row PolicyOverride
				if err := r.db.WithContext(ctx).First(&row, candidate.ID).Error; err != nil {
					return false, false, nil, false, "", false, nil, err
				}
				value = row.Value
			} else {
				var row TemporaryOverride
				if err := r.db.WithContext(ctx).First(&row, candidate.ID).Error; err != nil {
					return false, false, nil, false, "", false, nil, err
				}
				value = row.Value
			}
			var patch struct {
				DNS map[string]any `json:"dns"`
			}
			if err := json.Unmarshal([]byte(value), &patch); err != nil {
				return false, false, nil, false, "", false, nil, err
			}
			if patch.DNS != nil {
				d.DNS = patch.DNS
			}
		}
		if d.Quarantine {
			quarantined = true
			for _, item := range d.QuarantineAllowlist {
				allow[strings.ToLower(strings.TrimSpace(item))] = struct{}{}
			}
		}
		if !dnsSelected && len(d.DNS) > 0 {
			dnsSelected = true
			if managed, _ := d.DNS["managed"].(bool); managed {
				managedDNS = true
			}
			if tag, _ := d.DNS["dnsOutboundTag"].(string); strings.TrimSpace(tag) != "" {
				dnsOutboundTag = strings.TrimSpace(tag)
			}
			if ss, _ := d.DNS["safeSearch"].(bool); ss {
				safeSearch = true
				managedDNS = true
				if dnsOutboundTag == "" {
					limitations = append(limitations, "SafeSearch requires an explicit dnsOutboundTag whose resolver enforces SafeSearch")
				}
			}
		}
		if candidate.Source == "temporary" && candidate.Scope == ScopeQuarantineRelease && d.Quarantine {
			var row TemporaryOverride
			if err := r.db.WithContext(ctx).First(&row, candidate.ID).Error; err != nil {
				return false, false, nil, false, "", false, nil, err
			}
			var release struct {
				Released bool `json:"released"`
			}
			if err := json.Unmarshal([]byte(row.Value), &release); err != nil {
				return false, false, nil, false, "", false, nil, err
			}
			if release.Released {
				released = true
			}
		}
	}
	if released {
		quarantined = false
	}
	if managedDNS || safeSearch {
		limitations = append(limitations, "client-controlled DoH/DoH3, DoT, and cached DNS are outside Xray-observed DNS paths")
	}
	allowlist := make([]string, 0, len(allow))
	for item := range allow {
		allowlist = append(allowlist, item)
	}
	sort.Strings(allowlist)
	return quarantined, released, allowlist, managedDNS, dnsOutboundTag, safeSearch, uniqueStrings(limitations), nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
