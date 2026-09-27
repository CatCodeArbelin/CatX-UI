package audit

import "time"

const (
	EventSuccess = "success"
	EventFailure = "failure"

	DeliveryPending = "pending"
	DeliveryLeased  = "leased"
	DeliverySuccess = "success"
	DeliveryDead    = "dead"

	DefaultAuditRetentionDays      = 180
	DefaultDeliveryRetentionDays   = 30
	DefaultDeadLetterRetentionDays = 90
)

// AuditEvent is the authoritative, metadata-only record of a privileged
// operation. Payloads are never accepted by this model.
type AuditEvent struct {
	ID         string    `json:"id" gorm:"primaryKey;size:36"`
	CreatedAt  time.Time `json:"createdAt" gorm:"index:idx_audit_created_at"`
	EventType  string    `json:"eventType" gorm:"size:120;index:idx_audit_type_created,priority:1"`
	Outcome    string    `json:"outcome" gorm:"size:16;index:idx_audit_outcome_created,priority:1"`
	ActorType  string    `json:"actorType" gorm:"size:32;index:idx_audit_actor_created,priority:1"`
	ActorID    string    `json:"actorId,omitempty" gorm:"size:120;index:idx_audit_actor_created,priority:2"`
	ActorName  string    `json:"actorName,omitempty" gorm:"size:120"`
	AuthMethod string    `json:"authMethod,omitempty" gorm:"size:32"`
	TokenID    string    `json:"tokenId,omitempty" gorm:"size:64"`
	TargetType string    `json:"targetType,omitempty" gorm:"size:64;index:idx_audit_target_created,priority:1"`
	TargetRef  string    `json:"targetRef,omitempty" gorm:"size:128;index:idx_audit_target_created,priority:2"`
	NodeScope  string    `json:"nodeScope,omitempty" gorm:"size:16;index:idx_audit_node_created,priority:1"`
	NodeID     string    `json:"nodeId,omitempty" gorm:"size:120;index:idx_audit_node_created,priority:2"`
	RequestID  string    `json:"requestId" gorm:"size:36;index:idx_audit_request"`
	SourceIP   string    `json:"sourceIp,omitempty" gorm:"size:64"`
	StatusCode int       `json:"statusCode,omitempty"`
	Metadata   string    `json:"metadata,omitempty" gorm:"type:text"`
	ExpiresAt  time.Time `json:"-" gorm:"index:idx_audit_expires"`
}

func (AuditEvent) TableName() string { return "fork_audit_events" }

type WebhookEndpoint struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	Name         string    `json:"name" gorm:"size:120;not null"`
	URL          string    `json:"url" gorm:"size:2048;not null"`
	SecretCipher string    `json:"-" gorm:"type:text;not null"`
	EventTypes   string    `json:"eventTypes" gorm:"type:text;not null"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (WebhookEndpoint) TableName() string { return "fork_webhook_endpoints" }

type WebhookDelivery struct {
	ID            string     `json:"id" gorm:"primaryKey;size:36"`
	EndpointID    string     `json:"endpointId" gorm:"size:36;not null;uniqueIndex:idx_webhook_endpoint_event,priority:1"`
	EventID       string     `json:"eventId" gorm:"size:36;not null;uniqueIndex:idx_webhook_endpoint_event,priority:2;index:idx_webhook_status_retry,priority:1"`
	Status        string     `json:"status" gorm:"size:16;not null;index:idx_webhook_status_retry,priority:2"`
	Attempt       int        `json:"attempt"`
	NextAttemptAt time.Time  `json:"nextAttemptAt" gorm:"index:idx_webhook_status_retry,priority:3"`
	LeaseUntil    *time.Time `json:"-"`
	LastError     string     `json:"lastError,omitempty" gorm:"size:500"`
	DeliveredAt   *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt" gorm:"index:idx_webhook_created"`
}

func (WebhookDelivery) TableName() string { return "fork_webhook_deliveries" }

type RetentionSettingsRow struct {
	ID             uint `gorm:"primaryKey"`
	AuditDays      int  `gorm:"not null"`
	DeliveryDays   int  `gorm:"not null"`
	DeadLetterDays int  `gorm:"not null"`
}

func (RetentionSettingsRow) TableName() string { return "fork_audit_retention" }

type AuditRetentionSettings struct {
	AuditDays      int `json:"auditDays"`
	DeliveryDays   int `json:"deliveryDays"`
	DeadLetterDays int `json:"deadLetterDays"`
}

type WebhookEndpointRequest struct {
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Secret     string   `json:"secret"`
	EventTypes []string `json:"eventTypes"`
}

type RetentionRequest struct {
	AuditDays      int `json:"auditDays"`
	DeliveryDays   int `json:"deliveryDays"`
	DeadLetterDays int `json:"deadLetterDays"`
}

type EventPage struct {
	Items []AuditEvent `json:"items"`
	Next  string       `json:"next,omitempty"`
}

type DeliveryPage struct {
	Items []WebhookDelivery `json:"items"`
	Next  string            `json:"next,omitempty"`
}
