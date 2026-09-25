// Package policy owns the durable policy data contract. It intentionally does
// not make routing or enforcement decisions; WP-3B consumes this package.
package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"gorm.io/gorm"
)

const (
	TargetClient = "client"
	TargetGroup  = "group"
	ScopePolicy  = "policy"
	ScopeService = "service"
	ScopeDomain  = "domain"
	ScopeDNS     = "dns"
	ScopeQuota   = "quota"
	ScopeQoS     = "qos"
)

// Definition is an intentionally declarative, future-compatible policy
// document. WP-3A stores and validates it; WP-3B is responsible for meaning.
type Definition struct {
	Action       string         `json:"action"`
	Services     []string       `json:"services,omitempty"`
	Categories   []string       `json:"categories,omitempty"`
	Destinations []string       `json:"destinations,omitempty"`
	DNS          map[string]any `json:"dns,omitempty"`
	Quota        map[string]any `json:"quota,omitempty"`
	QoS          map[string]any `json:"qos,omitempty"`
	ScheduleRef  string         `json:"scheduleRef,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type Policy struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Description string `json:"description,omitempty"`
	Spec        string `json:"spec" gorm:"type:text;not null"`
	Priority    int    `json:"priority" gorm:"not null;default:0;index"`
	Enabled     bool   `json:"enabled" gorm:"not null;index"`
	CreatedAt   int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt   int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (Policy) TableName() string { return "fork_policies" }

type PolicyAssignment struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PolicyID   uint   `json:"policyId" gorm:"not null;index:idx_policy_assignments_policy;uniqueIndex:uidx_policy_assignment,priority:1"`
	TargetType string `json:"targetType" gorm:"not null;index:idx_policy_assignments_target,priority:1;uniqueIndex:uidx_policy_assignment,priority:2"`
	TargetRef  string `json:"targetRef" gorm:"not null;index:idx_policy_assignments_target,priority:2;uniqueIndex:uidx_policy_assignment,priority:3"`
	Priority   int    `json:"priority" gorm:"not null;default:0"`
	Enabled    bool   `json:"enabled" gorm:"not null"`
	CreatedAt  int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt  int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (PolicyAssignment) TableName() string { return "fork_policy_assignments" }

type PolicyOverride struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PolicyID   uint   `json:"policyId" gorm:"not null;index;uniqueIndex:uidx_policy_override,priority:1"`
	TargetType string `json:"targetType" gorm:"not null;index:idx_policy_overrides_target,priority:1;uniqueIndex:uidx_policy_override,priority:2"`
	TargetRef  string `json:"targetRef" gorm:"not null;index:idx_policy_overrides_target,priority:2;uniqueIndex:uidx_policy_override,priority:3"`
	Scope      string `json:"scope" gorm:"not null;uniqueIndex:uidx_policy_override,priority:4"`
	Value      string `json:"value" gorm:"type:text;not null"`
	Priority   int    `json:"priority" gorm:"not null;default:0"`
	Enabled    bool   `json:"enabled" gorm:"not null"`
	CreatedAt  int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt  int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (PolicyOverride) TableName() string { return "fork_policy_overrides" }

type TemporaryOverride struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PolicyID   uint   `json:"policyId" gorm:"not null;index"`
	TargetType string `json:"targetType" gorm:"not null;index:idx_policy_temp_target,priority:1"`
	TargetRef  string `json:"targetRef" gorm:"not null;index:idx_policy_temp_target,priority:2"`
	Scope      string `json:"scope" gorm:"not null"`
	Value      string `json:"value" gorm:"type:text;not null"`
	Priority   int    `json:"priority" gorm:"not null;default:0"`
	StartsAt   int64  `json:"startsAt" gorm:"not null;index"`
	ExpiresAt  int64  `json:"expiresAt" gorm:"not null;index"`
	Enabled    bool   `json:"enabled" gorm:"not null"`
	CreatedAt  int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt  int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (TemporaryOverride) TableName() string { return "fork_policy_temporary_overrides" }

func Migrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	return db.AutoMigrate(&Policy{}, &PolicyAssignment{}, &PolicyOverride{}, &TemporaryOverride{})
}

func validateTarget(targetType, targetRef string) error {
	if targetType != TargetClient && targetType != TargetGroup {
		return fmt.Errorf("invalid target type")
	}
	if strings.TrimSpace(targetRef) == "" {
		return fmt.Errorf("target reference is required")
	}
	return nil
}

func validateScope(scope string) error {
	switch scope {
	case ScopePolicy, ScopeService, ScopeDomain, ScopeDNS, ScopeQuota, ScopeQoS:
		return nil
	}
	return fmt.Errorf("invalid override scope")
}

func validateJSON(raw string, field string) error {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return fmt.Errorf("malformed %s", field)
	}
	if value == nil {
		return fmt.Errorf("%s must be an object", field)
	}
	if _, ok := value.(map[string]any); !ok {
		return fmt.Errorf("%s must be an object", field)
	}
	return nil
}

func validateDefinition(spec string) error {
	if err := validateJSON(spec, "policy spec"); err != nil {
		return err
	}
	var d Definition
	if err := json.Unmarshal([]byte(spec), &d); err != nil {
		return fmt.Errorf("malformed policy spec")
	}
	if d.Action != "" && d.Action != "allow" && d.Action != "deny" {
		return fmt.Errorf("policy action must be allow or deny")
	}
	return nil
}

func normalizeDefinition(spec string) (string, error) {
	if err := validateDefinition(spec); err != nil {
		return "", err
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(spec), &value); err != nil {
		return "", err
	}
	b, err := json.Marshal(value)
	return string(b), err
}

func validatePolicy(p *Policy) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" || len(p.Name) > 128 {
		return errors.New("policy name is required and must be at most 128 characters")
	}
	if p.Priority < -100000 || p.Priority > 100000 {
		return errors.New("policy priority is out of range")
	}
	spec, err := normalizeDefinition(p.Spec)
	if err != nil {
		return err
	}
	p.Spec = spec
	return nil
}

func validateAssignment(a *PolicyAssignment) error {
	return validateTarget(a.TargetType, a.TargetRef)
}

func validateOverride(o *PolicyOverride) error {
	if err := validateTarget(o.TargetType, o.TargetRef); err != nil {
		return err
	}
	if err := validateScope(o.Scope); err != nil {
		return err
	}
	if o.Priority < -100000 || o.Priority > 100000 {
		return errors.New("override priority is out of range")
	}
	return validateJSON(o.Value, "override value")
}

func validateTemporary(o *TemporaryOverride) error {
	if err := validateTarget(o.TargetType, o.TargetRef); err != nil {
		return err
	}
	if err := validateScope(o.Scope); err != nil {
		return err
	}
	if o.StartsAt <= 0 || o.ExpiresAt <= o.StartsAt {
		return errors.New("temporary override must have a positive start before expiry")
	}
	if o.ExpiresAt-o.StartsAt > int64((365*24*time.Hour)/time.Millisecond) {
		return errors.New("temporary override may not exceed 365 days")
	}
	if o.Priority < -100000 || o.Priority > 100000 {
		return errors.New("temporary override priority is out of range")
	}
	return validateJSON(o.Value, "temporary override value")
}

func clientOrGroupExists(db *gorm.DB, targetType, ref string) error {
	if db == nil {
		return errors.New("policy database is unavailable")
	}
	if targetType == TargetClient {
		var row model.ClientRecord
		if err := db.Where("email = ?", ref).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("client target does not exist")
			}
			return err
		}
		return nil
	}
	var group model.ClientGroup
	if err := db.Where("name = ?", ref).First(&group).Error; err == nil {
		return nil
	}
	var count int64
	if err := db.Model(&model.ClientRecord{}).Where("group_name = ?", ref).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("group target does not exist")
	}
	return nil
}

type resolvedCandidate struct {
	Source     string
	PolicyID   uint
	TargetType string
	TargetRef  string
	Priority   int
	CreatedAt  int64
	ID         uint
	Active     bool
}

type ResolvedPolicy struct {
	TargetClient string              `json:"targetClient,omitempty"`
	TargetGroup  string              `json:"targetGroup,omitempty"`
	At           int64               `json:"at"`
	Items        []resolvedCandidate `json:"items"`
}

func sortCandidates(items []resolvedCandidate) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		rank := func(x resolvedCandidate) int {
			source := map[string]int{"assignment": 1, "override": 2, "temporary": 3}[x.Source]
			target := map[string]int{TargetGroup: 1, TargetClient: 2}[x.TargetType]
			return source*1000000 + target*100000 + x.Priority
		}
		ra, rb := rank(a), rank(b)
		if ra != rb {
			return ra > rb
		}
		if a.CreatedAt != b.CreatedAt {
			return a.CreatedAt < b.CreatedAt
		}
		return a.ID < b.ID
	})
}

// activeAt is deliberately evaluated at read time, so restart or missed
// cleanup cannot make an expired temporary override visible.
func activeAt(now, starts, expires int64) bool { return starts <= now && now < expires }
