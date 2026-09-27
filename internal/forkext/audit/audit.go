package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	database                 atomic.Pointer[gorm.DB]
	enabled                  atomic.Bool
	errInvalidWebhookURL     = errors.New("webhook URL must be public HTTPS without query, fragment, userinfo, or redirects")
	errPrivateWebhookAddress = errors.New("webhook destination is not public")
)

func Configure(db *gorm.DB, on bool) {
	database.Store(db)
	enabled.Store(on && db != nil)
}

func Enabled() bool { return enabled.Load() && database.Load() != nil }

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&AuditEvent{}, &WebhookEndpoint{}, &WebhookDelivery{}, &RetentionSettingsRow{}); err != nil {
		return err
	}
	var row RetentionSettingsRow
	if err := db.First(&row, 1).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&RetentionSettingsRow{ID: 1, AuditDays: DefaultAuditRetentionDays, DeliveryDays: DefaultDeliveryRetentionDays, DeadLetterDays: DefaultDeadLetterRetentionDays}).Error
	} else {
		return err
	}
}

func newID() string { return uuid.NewString() }

func NewRequestID() string { return uuid.NewString() }

func requestID(c *gin.Context) string {
	if value, ok := c.Get("catx_request_id"); ok {
		if id, ok := value.(string); ok && id != "" {
			return id
		}
	}
	id := NewRequestID()
	c.Set("catx_request_id", id)
	c.Header("X-CatX-Request-ID", id)
	return id
}

// Middleware creates a server-owned correlation ID and records successful
// privileged mutations after their authoritative handler returns successfully.
// It deliberately never reads the request body or arbitrary headers.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !Enabled() {
			c.Next()
			return
		}
		id := requestID(c)
		c.Next()
		if c.IsAborted() || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			return
		}
		status := c.Writer.Status()
		if status < http.StatusOK || status >= http.StatusBadRequest || strings.HasPrefix(c.FullPath(), "/panel/api/fork/audit") {
			return
		}
		actorType, actorID, actorName, authMethod := actor(c)
		tokenID, _ := c.Get("catx_token_id")
		token, _ := tokenID.(string)
		params := make(map[string]any)
		for _, p := range c.Params {
			if p.Key == "id" || p.Key == "guid" || p.Key == "nodeId" {
				params[p.Key] = targetRef(p.Value)
			}
		}
		_ = Record(context.Background(), &AuditEvent{
			ID: newID(), CreatedAt: time.Now().UTC(), EventType: "http." + strings.ToLower(c.Request.Method),
			Outcome: EventSuccess, ActorType: actorType, ActorID: actorID, ActorName: actorName,
			AuthMethod: authMethod, TokenID: token, TargetType: "route", TargetRef: targetRef(c.FullPath()),
			RequestID: id, SourceIP: clientIP(c), StatusCode: status,
			Metadata: sanitizeMetadata(map[string]any{"route": c.FullPath(), "params": params}),
		})
	}
}

func actor(c *gin.Context) (string, string, string, string) {
	if portal, ok := c.Get("catx_portal_actor"); ok {
		if isPortal, ok := portal.(bool); ok && isPortal {
			id, _ := c.Get("catx_actor_id")
			name, _ := c.Get("catx_actor_name")
			method, _ := c.Get("catx_auth_method")
			actorID, _ := id.(string)
			actorName, _ := name.(string)
			auth, _ := method.(string)
			return "client_portal", actorID, actorName, auth
		}
	}
	if id, ok := c.Get("catx_actor_id"); ok {
		actorID, _ := id.(string)
		name, _ := c.Get("catx_actor_name")
		method, _ := c.Get("catx_auth_method")
		actorName, _ := name.(string)
		auth, _ := method.(string)
		return "user", actorID, actorName, auth
	}
	return "system", "", "", "unknown"
}

func clientIP(c *gin.Context) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(c.Request.RemoteAddr)
}

func Record(ctx context.Context, event *AuditEvent) error {
	db := database.Load()
	if !Enabled() || db == nil || event == nil {
		return nil
	}
	if event.ID == "" {
		event.ID = newID()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.ExpiresAt.IsZero() {
		event.ExpiresAt = event.CreatedAt.Add(DefaultAuditRetentionDays * 24 * time.Hour)
	}
	if event.Metadata == "" {
		event.Metadata = "{}"
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		metrics.auditWrites.WithLabelValues(event.Outcome).Inc()
		return enqueueForEvent(tx, event)
	})
}

func RecordInTx(ctx context.Context, tx *gorm.DB, event *AuditEvent) error {
	if !Enabled() || tx == nil || event == nil {
		return nil
	}
	if event.ID == "" {
		event.ID = newID()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.ExpiresAt.IsZero() {
		event.ExpiresAt = event.CreatedAt.Add(DefaultAuditRetentionDays * 24 * time.Hour)
	}
	if event.Metadata == "" {
		event.Metadata = "{}"
	}
	if err := tx.WithContext(ctx).Create(event).Error; err != nil {
		return err
	}
	metrics.auditWrites.WithLabelValues(event.Outcome).Inc()
	return enqueueForEvent(tx, event)
}

func RecordLogin(ctx context.Context, username, sourceIP, outcome, reason string, userID int) error {
	actorType := "anonymous"
	actorID := ""
	if userID > 0 {
		actorType = "user"
		actorID = strconv.Itoa(userID)
	}
	return Record(ctx, &AuditEvent{
		ID: newID(), CreatedAt: time.Now().UTC(), EventType: "auth.login", Outcome: outcome,
		ActorType: actorType, ActorID: actorID, TargetType: "panel_user", TargetRef: targetRef(username),
		AuthMethod: "password", SourceIP: sourceIP,
		Metadata: sanitizeMetadata(map[string]any{"reason": reason}),
	})
}

func enqueueForEvent(db *gorm.DB, event *AuditEvent) error {
	var endpoints []WebhookEndpoint
	if err := db.Where("enabled = ?", true).Find(&endpoints).Error; err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		if !subscribed(endpoint.EventTypes, event.EventType) {
			continue
		}
		delivery := &WebhookDelivery{ID: newID(), EndpointID: endpoint.ID, EventID: event.ID, Status: DeliveryPending, NextAttemptAt: time.Now().UTC(), CreatedAt: time.Now().UTC()}
		if err := db.Where("endpoint_id = ? AND event_id = ?", endpoint.ID, event.ID).FirstOrCreate(delivery).Error; err != nil {
			return err
		}
	}
	return nil
}

func subscribed(raw, eventType string) bool {
	var values []string
	if json.Unmarshal([]byte(raw), &values) != nil {
		return false
	}
	for _, value := range values {
		if value == "*" || value == eventType {
			return true
		}
	}
	return false
}

func CreateEndpoint(ctx context.Context, name, rawURL, secret string, events []string) (WebhookEndpoint, error) {
	if !Enabled() {
		return WebhookEndpoint{}, errors.New("audit feature is disabled")
	}
	name = strings.TrimSpace(name)
	rawURL = strings.TrimSpace(rawURL)
	if name == "" || len(name) > 120 || len(rawURL) > 2048 || len(secret) > 512 || len(events) == 0 || len(events) > 32 {
		return WebhookEndpoint{}, errors.New("invalid webhook fields")
	}
	for _, eventType := range events {
		if strings.TrimSpace(eventType) == "" || len(eventType) > 120 {
			return WebhookEndpoint{}, errors.New("invalid webhook event type")
		}
	}
	if err := validateURL(ctx, rawURL); err != nil {
		return WebhookEndpoint{}, err
	}
	if strings.TrimSpace(secret) == "" {
		return WebhookEndpoint{}, errors.New("webhook secret is required")
	}
	ciphertext, err := encryptSecret(secret)
	if err != nil {
		return WebhookEndpoint{}, err
	}
	encoded, err := json.Marshal(events)
	if err != nil {
		return WebhookEndpoint{}, err
	}
	now := time.Now().UTC()
	endpoint := WebhookEndpoint{ID: newID(), Name: strings.TrimSpace(name), URL: rawURL, SecretCipher: ciphertext, EventTypes: string(encoded), Enabled: true, CreatedAt: now, UpdatedAt: now}
	if err := database.Load().WithContext(ctx).Create(&endpoint).Error; err != nil {
		return WebhookEndpoint{}, err
	}
	endpoint.SecretCipher = ""
	return endpoint, nil
}

func validateURL(ctx context.Context, raw string) error {
	if err := publicHTTPS(raw); err != nil {
		return err
	}
	u, _ := url.Parse(raw)
	resolver := net.DefaultResolver
	if deadline, ok := ctx.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}
	ips, err := resolver.LookupIP(ctx, "ip", u.Hostname())
	if err != nil || len(ips) == 0 {
		return errPrivateWebhookAddress
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return errPrivateWebhookAddress
		}
	}
	return nil
}

func GenerateSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func ParsePageSize(raw string) int {
	value, _ := strconv.Atoi(raw)
	if value < 1 {
		value = 50
	}
	if value > 200 {
		value = 200
	}
	return value
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%T", err)
}
