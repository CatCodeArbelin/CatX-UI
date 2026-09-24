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
)

var allFlags = [...]Flag{
	FlagAnalytics,
	FlagDNSIntelligence,
	FlagPolicies,
	FlagTrafficControl,
	FlagSecurityAnomaly,
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
	key := settingKey(flag)
	value := strconv.FormatBool(enabled)
	var row model.Setting
	err := s.db.Where("key = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.db.Create(&model.Setting{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	row.Value = value
	return s.db.Save(&row).Error
}
