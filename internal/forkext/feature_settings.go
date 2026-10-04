package forkext

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
		payload, err := readFeatureSettings()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "feature settings saved", "obj": payload})
	})
}
