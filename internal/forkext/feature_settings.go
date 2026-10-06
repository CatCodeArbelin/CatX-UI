package forkext

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/policy"
)

var errFeatureSettingsUnavailable = errors.New("feature settings unavailable")

func readFeatureSettings() (gin.H, error) {
	db := currentSettingsDB()
	if db == nil {
		return nil, errFeatureSettingsUnavailable
	}
	items, err := NewSettings(db).FeatureFlags()
	if err != nil {
		return nil, errFeatureSettingsUnavailable
	}
	restartRequired := false
	for _, item := range items {
		if item.RestartRequired {
			restartRequired = true
			break
		}
	}
	return gin.H{"items": items, "restartRequired": restartRequired}, nil
}

func registerFeatureSettingsRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/fork/settings/features", func(c *gin.Context) {
		payload, err := readFeatureSettings()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": payload})
	})
	api.PUT("/fork/settings/features", func(c *gin.Context) {
		db := currentSettingsDB()
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		var input struct {
			Flags map[string]bool `json:"flags" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil || input.Flags == nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid feature settings"})
			return
		}
		updates := make(map[Flag]bool, len(input.Flags))
		for key, enabled := range input.Flags {
			updates[Flag(key)] = enabled
		}
		if err := NewSettings(db).UpdateFeatures(updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		if _, changed := updates[FlagPolicies]; changed {
			// Policy persistence and the feature toggle share the existing Xray
			// apply boundary. Keep the API truthful while that boundary is pending.
			policy.MarkRuntimeApplyRequired()
		}
		// Schema preparation is deliberately explicit and synchronous. A
		// generic panel restart must never become an implicit migration entry
		// point, and a successful settings write must not hide an activation
		// failure behind a later asynchronous restart.
		if err := PrepareRuntimeFromSettings(db); err != nil {
			payload, readErr := readFeatureSettings()
			if readErr != nil {
				c.JSON(http.StatusConflict, gin.H{"success": false, "msg": "feature settings saved but activation preparation failed"})
				return
			}
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"msg":     "feature settings saved but activation preparation failed: " + err.Error(),
				"obj":     payload,
			})
			return
		}
		payload, err := readFeatureSettings()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "feature settings saved", "obj": payload})
	})
}
