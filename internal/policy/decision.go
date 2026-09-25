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
}

type Decision struct {
	ClientEmail   string              `json:"clientEmail"`
	GroupName     string              `json:"groupName,omitempty"`
	PolicyID      uint                `json:"policyId"`
	PolicyName    string              `json:"policyName"`
	Action        string              `json:"action"`
	Services      []string            `json:"services,omitempty"`
	Categories    []string            `json:"categories,omitempty"`
	Destinations  []string            `json:"destinations,omitempty"`
	Winner        ResolvedCandidate   `json:"winner"`
	Candidates    []ResolvedCandidate `json:"candidates"`
	Explanation   string              `json:"explanation"`
	OverrideScope string              `json:"overrideScope,omitempty"`
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
		resolved, err := r.Resolve(ctx, c.Email, c.Group, at)
		if err != nil {
			return nil, err
		}
		if len(resolved.Items) == 0 {
			continue
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
				Action       string   `json:"action"`
				Services     []string `json:"services"`
				Categories   []string `json:"categories"`
				Destinations []string `json:"destinations"`
				Domains      []string `json:"domains"`
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
			}
		}
		if spec.Action == "" {
			continue
		}
		if spec.Action != "allow" && spec.Action != "deny" {
			return nil, fmt.Errorf("policy %d: unsupported action %q", p.ID, spec.Action)
		}
		out = append(out, Decision{ClientEmail: c.Email, GroupName: c.Group, PolicyID: p.ID, PolicyName: p.Name, Action: spec.Action, Services: spec.Services, Categories: spec.Categories, Destinations: spec.Destinations, Winner: winner, Candidates: resolved.Items, OverrideScope: winner.Scope, Explanation: fmt.Sprintf("%s %s target %s by %s", winner.Source, winner.TargetType, winner.TargetRef, p.Name)})
	}
	return out, nil
}
