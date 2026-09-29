package forkext

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerFeatureSettingsRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/fork/settings/features", func(c *gin.Context) {
		db := currentSettingsDB()
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		items, err := NewSettings(db).FeatureFlags()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": gin.H{"items": items, "restartRequired": true}})
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
		items, err := NewSettings(db).FeatureFlags()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "feature settings unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "feature settings saved", "obj": gin.H{"items": items, "restartRequired": true}})
	})
}
