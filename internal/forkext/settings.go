package forkext

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"

	"gorm.io/gorm"
)

// Flag is a namespaced fork capability switch. Fork flags default to OFF and
// are kept out of the upstream reflected settings DTOs.
type Flag string

const (
	FlagAnalytics       Flag = "analytics.enabled"
	FlagDNSIntelligence Flag = "dns_intelligence.enabled"
	FlagPolicies        Flag = "policies.enabled"
	FlagTrafficControl  Flag = "traffic_control.enabled"
	FlagSecurityAnomaly Flag = "security_anomaly.enabled"
	FlagAudit           Flag = "audit.enabled"
	FlagWebhooks        Flag = "webhooks.enabled"
	FlagMetrics         Flag = "metrics.enabled"
	FlagSelfService     Flag = "self_service.enabled"
	FlagFleetUpdates    Flag = "fleet_updates.enabled"
	FlagFleetMutation   Flag = "fleet_updates.mutation.enabled"
)

var allFlags = [...]Flag{
	FlagAnalytics,
	FlagDNSIntelligence,
	FlagPolicies,
	FlagTrafficControl,
	FlagSecurityAnomaly,
	FlagAudit,
	FlagWebhooks,
	FlagMetrics,
	FlagSelfService,
	FlagFleetUpdates,
	FlagFleetMutation,
}

func settingKey(flag Flag) string {
	return "fork." + string(flag)
}

func validFlag(flag Flag) bool {
	for _, known := range allFlags {
		if flag == known {
			return true
		}
	}
	return false
}

// Settings is the fork-owned settings facade. It stores values in the
// existing key/value settings table, so the facade is compatible with both
// SQLite and PostgreSQL without changing the upstream model or schema.
type Settings struct {
	db *gorm.DB
}

// FeatureFlagInfo is the small, fork-owned DTO used by the CatX feature
// settings page. It deliberately excludes reserved flags that have no active
// runtime behavior.
type FeatureFlagInfo struct {
	Key             Flag   `json:"key"`
	Enabled         bool   `json:"enabled"`
	Requires        []Flag `json:"requires,omitempty"`
	RestartRequired bool   `json:"restartRequired"`
}

var managedFeatureFlags = []struct {
	flag     Flag
	requires []Flag
}{
	{FlagAnalytics, nil},
	{FlagDNSIntelligence, []Flag{FlagAnalytics}},
	{FlagPolicies, nil},
	{FlagTrafficControl, nil},
	{FlagSecurityAnomaly, []Flag{FlagAnalytics}},
	{FlagAudit, nil},
	{FlagSelfService, nil},
	{FlagFleetUpdates, nil},
	{FlagFleetMutation, []Flag{FlagFleetUpdates}},
}

func NewSettings(db *gorm.DB) Settings {
	return Settings{db: db}
}

// Flags returns every known flag, including defaults for keys not yet stored.
func (s Settings) Flags() (map[Flag]bool, error) {
	values := make(map[Flag]bool, len(allFlags))
	for _, flag := range allFlags {
		value, err := s.Enabled(flag)
		if err != nil {
			return nil, err
		}
		values[flag] = value
	}
	return values, nil
}

// Enabled reports a flag value. A nil database is treated as the safe,
// disabled configuration, which keeps startup and feature-off paths benign.
func (s Settings) Enabled(flag Flag) (bool, error) {
	if !validFlag(flag) {
		return false, fmt.Errorf("unknown fork feature flag %q", flag)
	}
	if s.db == nil {
		return false, nil
	}
	var row model.Setting
	err := s.db.Where("key = ?", settingKey(flag)).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	value, err := strconv.ParseBool(row.Value)
	if err != nil {
		return false, fmt.Errorf("invalid value for fork flag %q: %w", flag, err)
	}
	return value, nil
}

// Set updates a namespaced fork flag without touching upstream settings.
func (s Settings) Set(flag Flag, enabled bool) error {
	if !validFlag(flag) {
		return fmt.Errorf("unknown fork feature flag %q", flag)
	}
	if s.db == nil {
		return errors.New("fork settings database is nil")
	}
	return setFlag(s.db, flag, enabled)
}

func setFlag(db *gorm.DB, flag Flag, enabled bool) error {
	key := settingKey(flag)
	value := strconv.FormatBool(enabled)
	var row model.Setting
	err := db.Where("key = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&model.Setting{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	row.Value = value
	return db.Save(&row).Error
}

func managedFeature(flag Flag) (requires []Flag, ok bool) {
	for _, feature := range managedFeatureFlags {
		if feature.flag == flag {
			return feature.requires, true
		}
	}
	return nil, false
}

// FeatureFlags returns only flags with an active CatX runtime consumer.
func (s Settings) FeatureFlags() ([]FeatureFlagInfo, error) {
	result := make([]FeatureFlagInfo, 0, len(managedFeatureFlags))
	for _, feature := range managedFeatureFlags {
		enabled, err := s.Enabled(feature.flag)
		if err != nil {
			return nil, err
		}
		result = append(result, FeatureFlagInfo{
			Key:             feature.flag,
			Enabled:         enabled,
			Requires:        feature.requires,
			RestartRequired: true,
		})
	}
	return result, nil
}

func validateFeatureState(values map[Flag]bool) error {
	for _, feature := range managedFeatureFlags {
		if !values[feature.flag] {
			continue
		}
		for _, dependency := range feature.requires {
			if !values[dependency] {
				return fmt.Errorf("%s requires %s", feature.flag, dependency)
			}
		}
	}
	return nil
}

// UpdateFeatures validates the complete resulting state and persists the
// requested changes atomically. Dependencies are never enabled implicitly.
func (s Settings) UpdateFeatures(updates map[Flag]bool) error {
	if s.db == nil {
		return errors.New("fork settings database is nil")
	}
	current := make(map[Flag]bool, len(managedFeatureFlags))
	for _, feature := range managedFeatureFlags {
		enabled, err := s.Enabled(feature.flag)
		if err != nil {
			return err
		}
		current[feature.flag] = enabled
	}
	for flag, enabled := range updates {
		if _, ok := managedFeature(flag); !ok {
			return fmt.Errorf("unknown or unmanaged fork feature flag %q", flag)
		}
		current[flag] = enabled
	}
	if err := validateFeatureState(current); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for flag, enabled := range updates {
			if err := setFlag(tx, flag, enabled); err != nil {
				return err
			}
		}
		return nil
	})
}
