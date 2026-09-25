package forkext

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"

	"gorm.io/gorm"
)

const (
	retentionRawDays       = "fork.analytics.retention.raw_days"
	retentionDNSDays       = "fork.analytics.retention.dns_days"
	retentionSessionDays   = "fork.analytics.retention.session_days"
	retentionAggregateDays = "fork.analytics.retention.aggregate_days"
)

type retentionDays struct {
	RawEvents  int `json:"rawEvents"`
	DNS        int `json:"dnsObservations"`
	Sessions   int `json:"sessions"`
	Aggregates int `json:"aggregates"`
}

type analyticsSettingsResponse struct {
	Enabled         bool          `json:"enabled"`
	DNSIntelligence bool          `json:"dnsIntelligence"`
	Retention       retentionDays `json:"retention"`
	Privacy         struct {
		MetadataOnly bool `json:"metadataOnly"`
		Uncertain    bool `json:"classificationsMayBeUncertain"`
	} `json:"privacy"`
}

var settingsState struct {
	sync.RWMutex
	db *gorm.DB
}

func setSettingsDB(db *gorm.DB) {
	settingsState.Lock()
	settingsState.db = db
	settingsState.Unlock()
}

func currentSettingsDB() *gorm.DB {
	settingsState.RLock()
	defer settingsState.RUnlock()
	return settingsState.db
}

func retentionFromPolicy(policy analytics.RetentionPolicy) retentionDays {
	return retentionDays{RawEvents: int(policy.RawEvents / (24 * time.Hour)), DNS: int(policy.DNSObservations / (24 * time.Hour)), Sessions: int(policy.Sessions / (24 * time.Hour)), Aggregates: int(policy.Aggregates / (24 * time.Hour))}
}

func retentionPolicyFromDB(db *gorm.DB) analytics.RetentionPolicy {
	policy := analytics.DefaultRetentionPolicy()
	if db == nil {
		return policy
	}
	for key, target := range map[string]*time.Duration{
		retentionRawDays:       &policy.RawEvents,
		retentionDNSDays:       &policy.DNSObservations,
		retentionSessionDays:   &policy.Sessions,
		retentionAggregateDays: &policy.Aggregates,
	} {
		var row model.Setting
		if err := db.Where("key = ?", key).First(&row).Error; err != nil {
			continue
		}
		if days, err := strconv.Atoi(row.Value); err == nil && days >= 1 && days <= 3650 {
			*target = time.Duration(days) * 24 * time.Hour
		}
	}
	return policy
}

func saveRetention(db *gorm.DB, key string, days int) error {
	if days < 1 || days > 3650 {
		return errors.New("retention must be between 1 and 3650 days")
	}
	var row model.Setting
	err := db.Where("key = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&model.Setting{Key: key, Value: strconv.Itoa(days)}).Error
	}
	if err != nil {
		return err
	}
	row.Value = strconv.Itoa(days)
	return db.Save(&row).Error
}

func analyticsSettings(db *gorm.DB) analyticsSettingsResponse {
	status := analytics.CurrentStatus()
	result := analyticsSettingsResponse{Enabled: status.Enabled, DNSIntelligence: status.DNSIntelligence, Retention: retentionFromPolicy(status.Retention)}
	result.Privacy.MetadataOnly = true
	result.Privacy.Uncertain = true
	if db != nil {
		if dns, err := NewSettings(db).Enabled(FlagDNSIntelligence); err == nil {
			result.DNSIntelligence = status.Enabled && dns
		}
	}
	return result
}

func registerAnalyticsSettingsRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/analytics/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": analyticsSettings(currentSettingsDB())})
	})
	api.POST("/analytics/settings", func(c *gin.Context) {
		db := currentSettingsDB()
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "analytics settings unavailable"})
			return
		}
		var input struct {
			DNSIntelligence *bool          `json:"dnsIntelligence"`
			Retention       *retentionDays `json:"retention"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid analytics settings"})
			return
		}
		if input.DNSIntelligence != nil {
			if err := NewSettings(db).Set(FlagDNSIntelligence, *input.DNSIntelligence); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "could not save DNS privacy setting"})
				return
			}
			analytics.SetEvidenceEnabled(*input.DNSIntelligence)
		}
		if input.Retention != nil {
			values := map[string]int{retentionRawDays: input.Retention.RawEvents, retentionDNSDays: input.Retention.DNS, retentionSessionDays: input.Retention.Sessions, retentionAggregateDays: input.Retention.Aggregates}
			for key, days := range values {
				if days == 0 {
					continue
				}
				if err := saveRetention(db, key, days); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
					return
				}
			}
			analytics.SetRetentionPolicy(retentionPolicyFromDB(db))
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "analytics settings saved", "obj": analyticsSettings(db)})
	})
}
