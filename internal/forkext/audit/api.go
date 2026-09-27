package audit

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil || !Enabled() {
		return
	}
	fork := api.Group("/fork")
	fork.GET("/metrics", func(c *gin.Context) { MetricsHandler().ServeHTTP(c.Writer, c.Request) })
	auditGroup := fork.Group("/audit")
	auditGroup.GET("/events", listEvents)
	auditGroup.GET("/events/:id", getEvent)
	auditGroup.DELETE("/events", deleteEvents)
	auditGroup.GET("/webhooks", listEndpoints)
	auditGroup.POST("/webhooks", createEndpoint)
	auditGroup.DELETE("/webhooks/:id", deleteEndpoint)
	auditGroup.POST("/webhooks/:id/replay", replayEndpoint)
	auditGroup.GET("/webhooks/deliveries", listDeliveries)
	auditGroup.POST("/webhooks/deliveries/:id/replay", replayDelivery)
	auditGroup.GET("/retention", getRetention)
	auditGroup.POST("/retention", setRetention)
}

func listEvents(c *gin.Context) {
	db := database.Load()
	q := db.WithContext(c.Request.Context()).Model(&AuditEvent{}).Order("created_at desc")
	if eventType := strings.TrimSpace(c.Query("eventType")); eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if outcome := strings.TrimSpace(c.Query("outcome")); outcome != "" {
		q = q.Where("outcome = ?", outcome)
	}
	if actor := strings.TrimSpace(c.Query("actorId")); actor != "" {
		q = q.Where("actor_id = ?", actor)
	}
	if target := strings.TrimSpace(c.Query("targetType")); target != "" {
		q = q.Where("target_type = ?", target)
	}
	if node := strings.TrimSpace(c.Query("nodeId")); node != "" {
		q = q.Where("node_id = ?", node)
	}
	if requestID := strings.TrimSpace(c.Query("requestId")); requestID != "" {
		q = q.Where("request_id = ?", requestID)
	}
	limit := ParsePageSize(c.Query("limit"))
	var rows []AuditEvent
	if cursor := strings.TrimSpace(c.Query("cursor")); cursor != "" {
		q = q.Where("created_at < ?", cursor)
	}
	if err := q.Limit(limit + 1).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "audit query failed"})
		return
	}
	page := EventPage{Items: rows}
	if len(rows) > limit {
		page.Items = rows[:limit]
		page.Next = page.Items[len(page.Items)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": page})
}

func getEvent(c *gin.Context) {
	var row AuditEvent
	if database.Load().WithContext(c.Request.Context()).First(&row, "id = ?", c.Param("id")).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "audit event not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}

func deleteEvents(c *gin.Context) {
	before, err := time.Parse(time.RFC3339, c.Query("before"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "before must be RFC3339"})
		return
	}
	result := database.Load().WithContext(c.Request.Context()).Where("created_at < ?", before).Delete(&AuditEvent{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "audit deletion failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"deleted": result.RowsAffected}})
}

func listEndpoints(c *gin.Context) {
	var rows []WebhookEndpoint
	if err := database.Load().WithContext(c.Request.Context()).Order("created_at desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "webhook query failed"})
		return
	}
	for i := range rows {
		rows[i].SecretCipher = ""
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": rows})
}

func createEndpoint(c *gin.Context) {
	var req WebhookEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid webhook request"})
		return
	}
	row, err := CreateEndpoint(c.Request.Context(), req.Name, req.URL, req.Secret, req.EventTypes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": ErrorMessage(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}

func deleteEndpoint(c *gin.Context) {
	result := database.Load().WithContext(c.Request.Context()).Delete(&WebhookEndpoint{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "webhook deletion failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func listDeliveries(c *gin.Context) {
	q := database.Load().WithContext(c.Request.Context()).Model(&WebhookDelivery{}).Order("created_at desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		q = q.Where("status = ?", status)
	}
	limit := ParsePageSize(c.Query("limit"))
	var rows []WebhookDelivery
	if err := q.Limit(limit + 1).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "delivery query failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": DeliveryPage{Items: rows}})
}

func replayEndpoint(c *gin.Context) {
	var endpoint WebhookEndpoint
	if database.Load().First(&endpoint, "id = ?", c.Param("id")).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "webhook not found"})
		return
	}
	var events []AuditEvent
	if err := database.Load().Order("created_at desc").Limit(1).Find(&events).Error; err != nil || len(events) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "no audit event available"})
		return
	}
	delivery := WebhookDelivery{ID: uuid.NewString(), EndpointID: endpoint.ID, EventID: events[0].ID, Status: DeliveryPending, NextAttemptAt: time.Now().UTC(), CreatedAt: time.Now().UTC()}
	if err := database.Load().Create(&delivery).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "msg": "delivery already queued"})
		return
	}
	Wake()
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": delivery})
}

func replayDelivery(c *gin.Context) {
	var source WebhookDelivery
	if database.Load().First(&source, "id = ?", c.Param("id")).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "delivery not found"})
		return
	}
	delivery := WebhookDelivery{ID: uuid.NewString(), EndpointID: source.EndpointID, EventID: source.EventID, Status: DeliveryPending, NextAttemptAt: time.Now().UTC(), CreatedAt: time.Now().UTC()}
	if err := database.Load().Create(&delivery).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "msg": "delivery already queued"})
		return
	}
	Wake()
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": delivery})
}

func getRetention(c *gin.Context) {
	var row RetentionSettingsRow
	if err := database.Load().First(&row, 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "retention query failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": AuditRetentionSettings{AuditDays: row.AuditDays, DeliveryDays: row.DeliveryDays, DeadLetterDays: row.DeadLetterDays}})
}

func setRetention(c *gin.Context) {
	var req RetentionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AuditDays < 1 || req.DeliveryDays < 1 || req.DeadLetterDays < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid retention"})
		return
	}
	row := RetentionSettingsRow{ID: 1, AuditDays: req.AuditDays, DeliveryDays: req.DeliveryDays, DeadLetterDays: req.DeadLetterDays}
	if err := database.Load().Save(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "retention update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": req})
}
