package policy

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreatePolicy(ctx context.Context, p *Policy) error {
	if err := validatePolicy(p); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *Repository) GetPolicy(ctx context.Context, id uint) (Policy, error) {
	var p Policy
	err := r.db.WithContext(ctx).First(&p, id).Error
	return p, err
}

func (r *Repository) ListPolicies(ctx context.Context) ([]Policy, error) {
	var rows []Policy
	err := r.db.WithContext(ctx).Order("priority DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdatePolicy(ctx context.Context, p *Policy) error {
	if err := validatePolicy(p); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&Policy{}).Where("id = ?", p.ID).Updates(map[string]any{"name": p.Name, "description": p.Description, "spec": p.Spec, "priority": p.Priority, "enabled": p.Enabled, "updated_at": time.Now().UnixMilli()}).Error
}

func (r *Repository) DeletePolicy(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, m := range []any{&PolicyAssignment{}, &PolicyOverride{}, &TemporaryOverride{}} {
			if err := tx.Where("policy_id = ?", id).Delete(m).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&Policy{}, id).Error
	})
}

func (r *Repository) CreateAssignment(ctx context.Context, a *PolicyAssignment) error {
	if err := validateAssignment(a); err != nil {
		return err
	}
	if err := r.policyExists(ctx, a.PolicyID); err != nil {
		return err
	}
	if err := clientOrGroupExists(r.db, a.TargetType, a.TargetRef); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) ListAssignments(ctx context.Context) ([]PolicyAssignment, error) {
	var rows []PolicyAssignment
	err := r.db.WithContext(ctx).Order("priority DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) DeleteAssignment(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PolicyAssignment{}, id).Error
}

func (r *Repository) UpdateAssignment(ctx context.Context, a *PolicyAssignment) error {
	if err := validateAssignment(a); err != nil {
		return err
	}
	if err := r.policyExists(ctx, a.PolicyID); err != nil {
		return err
	}
	if err := clientOrGroupExists(r.db, a.TargetType, a.TargetRef); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&PolicyAssignment{}).Where("id = ?", a.ID).Updates(map[string]any{"policy_id": a.PolicyID, "target_type": a.TargetType, "target_ref": a.TargetRef, "priority": a.Priority, "enabled": a.Enabled, "updated_at": time.Now().UnixMilli()}).Error
}

func (r *Repository) CreateOverride(ctx context.Context, o *PolicyOverride) error {
	if err := validateOverride(o); err != nil {
		return err
	}
	if err := r.policyExists(ctx, o.PolicyID); err != nil {
		return err
	}
	if err := clientOrGroupExists(r.db, o.TargetType, o.TargetRef); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *Repository) ListOverrides(ctx context.Context) ([]PolicyOverride, error) {
	var rows []PolicyOverride
	err := r.db.WithContext(ctx).Order("priority DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) DeleteOverride(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PolicyOverride{}, id).Error
}

func (r *Repository) UpdateOverride(ctx context.Context, o *PolicyOverride) error {
	if err := validateOverride(o); err != nil {
		return err
	}
	if err := r.policyExists(ctx, o.PolicyID); err != nil {
		return err
	}
	if err := clientOrGroupExists(r.db, o.TargetType, o.TargetRef); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&PolicyOverride{}).Where("id = ?", o.ID).Updates(map[string]any{"policy_id": o.PolicyID, "target_type": o.TargetType, "target_ref": o.TargetRef, "scope": o.Scope, "value": o.Value, "priority": o.Priority, "enabled": o.Enabled, "updated_at": time.Now().UnixMilli()}).Error
}

func (r *Repository) CreateTemporaryOverride(ctx context.Context, o *TemporaryOverride) error {
	if err := validateTemporary(o); err != nil {
		return err
	}
	if err := r.policyExists(ctx, o.PolicyID); err != nil {
		return err
	}
	if err := clientOrGroupExists(r.db, o.TargetType, o.TargetRef); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *Repository) ListTemporaryOverrides(ctx context.Context, now int64) ([]TemporaryOverride, error) {
	var rows []TemporaryOverride
	err := r.db.WithContext(ctx).Where("expires_at > ?", now).Order("priority DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) DeleteTemporaryOverride(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&TemporaryOverride{}, id).Error
}

func (r *Repository) UpdateTemporaryOverride(ctx context.Context, o *TemporaryOverride) error {
	if err := validateTemporary(o); err != nil {
		return err
	}
	if err := r.policyExists(ctx, o.PolicyID); err != nil {
		return err
	}
	if err := clientOrGroupExists(r.db, o.TargetType, o.TargetRef); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&TemporaryOverride{}).Where("id = ?", o.ID).Updates(map[string]any{"policy_id": o.PolicyID, "target_type": o.TargetType, "target_ref": o.TargetRef, "scope": o.Scope, "value": o.Value, "priority": o.Priority, "starts_at": o.StartsAt, "expires_at": o.ExpiresAt, "enabled": o.Enabled, "updated_at": time.Now().UnixMilli()}).Error
}

func (r *Repository) policyExists(ctx context.Context, id uint) error {
	var p Policy
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("policy does not exist")
		}
		return err
	}
	return nil
}

func (r *Repository) Resolve(ctx context.Context, clientEmail, groupName string, at int64) (ResolvedPolicy, error) {
	if at <= 0 {
		at = time.Now().UnixMilli()
	}
	result := ResolvedPolicy{TargetClient: clientEmail, TargetGroup: groupName, At: at, Items: []resolvedCandidate{}}
	var assignments []PolicyAssignment
	q := r.db.WithContext(ctx).Where("enabled = ? AND ((target_type = ? AND target_ref = ?) OR (target_type = ? AND target_ref = ?))", true, TargetClient, clientEmail, TargetGroup, groupName)
	if err := q.Find(&assignments).Error; err != nil {
		return result, err
	}
	for _, a := range assignments {
		var p Policy
		if err := r.db.WithContext(ctx).First(&p, a.PolicyID).Error; err != nil {
			return result, err
		}
		if !p.Enabled {
			continue
		}
		result.Items = append(result.Items, resolvedCandidate{Source: "assignment", PolicyID: p.ID, TargetType: a.TargetType, TargetRef: a.TargetRef, Priority: a.Priority + p.Priority, CreatedAt: a.CreatedAt, ID: a.ID, Active: true})
	}
	var overrides []PolicyOverride
	if err := r.db.WithContext(ctx).Where("enabled = ? AND ((target_type = ? AND target_ref = ?) OR (target_type = ? AND target_ref = ?))", true, TargetClient, clientEmail, TargetGroup, groupName).Find(&overrides).Error; err != nil {
		return result, err
	}
	for _, o := range overrides {
		if err := r.policyExists(ctx, o.PolicyID); err != nil {
			return result, err
		}
		result.Items = append(result.Items, resolvedCandidate{Source: "override", PolicyID: o.PolicyID, TargetType: o.TargetType, TargetRef: o.TargetRef, Priority: o.Priority, CreatedAt: o.CreatedAt, ID: o.ID, Active: true})
	}
	var temporary []TemporaryOverride
	if err := r.db.WithContext(ctx).Where("enabled = ? AND starts_at <= ? AND expires_at > ? AND ((target_type = ? AND target_ref = ?) OR (target_type = ? AND target_ref = ?))", true, at, at, TargetClient, clientEmail, TargetGroup, groupName).Find(&temporary).Error; err != nil {
		return result, err
	}
	for _, o := range temporary {
		if !activeAt(at, o.StartsAt, o.ExpiresAt) {
			continue
		}
		if err := r.policyExists(ctx, o.PolicyID); err != nil {
			return result, err
		}
		result.Items = append(result.Items, resolvedCandidate{Source: "temporary", PolicyID: o.PolicyID, TargetType: o.TargetType, TargetRef: o.TargetRef, Priority: o.Priority, CreatedAt: o.CreatedAt, ID: o.ID, Active: true})
	}
	sortCandidates(result.Items)
	return result, nil
}
