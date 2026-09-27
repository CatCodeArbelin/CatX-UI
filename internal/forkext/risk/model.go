package risk

import "time"

const (
	StateInsufficient = "insufficient_evidence"
	StateHealthy      = "healthy"
	StateDegraded     = "degraded"
	StateStale        = "stale"

	SignalConcurrentIPs     = "concurrent_ips"
	SignalConcurrentNodes   = "concurrent_nodes"
	SignalCountryDivergence = "country_divergence"
	SignalASNDivergence     = "asn_divergence"
	SignalIPChurn           = "ip_churn"
	SignalDNSAnomaly        = "dns_anomaly"
)

const (
	DefaultIPRetention    = 30 * 24 * time.Hour
	DefaultEventRetention = 90 * 24 * time.Hour
)

// IPHistory is the durable fork-owned extension of the existing IP
// observation path. It contains metadata only and never payload data.
type IPHistory struct {
	ID                 uint    `json:"id" gorm:"primaryKey"`
	ClientEmail        string  `json:"clientEmail" gorm:"not null;index:idx_risk_ip_client_time,priority:1"`
	IP                 string  `json:"ip" gorm:"not null;index:idx_risk_ip_value_time,priority:1"`
	NodeGuid           string  `json:"nodeGuid,omitempty" gorm:"index:idx_risk_ip_node_time,priority:1"`
	ObservedAt         int64   `json:"observedAt" gorm:"not null;index:idx_risk_ip_client_time,priority:2;index:idx_risk_ip_node_time,priority:2"`
	IngestedAt         int64   `json:"ingestedAt" gorm:"not null"`
	ASN                uint32  `json:"asn,omitempty"`
	Country            string  `json:"country,omitempty"`
	MetadataSource     string  `json:"metadataSource,omitempty"`
	MetadataConfidence float64 `json:"metadataConfidence"`
	MetadataState      string  `json:"metadataState"`
	Source             string  `json:"source" gorm:"not null"`
	DedupeKey          string  `json:"-" gorm:"uniqueIndex:idx_risk_ip_dedupe;not null"`
}

func (IPHistory) TableName() string { return "fork_risk_ip_history" }

type Event struct {
	ID                uint    `json:"id" gorm:"primaryKey"`
	ClientEmail       string  `json:"clientEmail" gorm:"not null;index:idx_risk_event_client_time,priority:1"`
	Kind              string  `json:"kind" gorm:"not null;index:idx_risk_event_kind_time,priority:1"`
	ObservedAt        int64   `json:"observedAt" gorm:"not null;index:idx_risk_event_client_time,priority:2;index:idx_risk_event_kind_time,priority:2"`
	ScoreContribution int     `json:"scoreContribution"`
	Confidence        float64 `json:"confidence"`
	State             string  `json:"state"`
	Evidence          string  `json:"evidence" gorm:"type:text"`
	SourceNode        string  `json:"sourceNode,omitempty"`
	AcknowledgedAt    int64   `json:"acknowledgedAt,omitempty"`
	DedupeKey         string  `json:"-" gorm:"uniqueIndex:idx_risk_event_dedupe;not null"`
}

func (Event) TableName() string { return "fork_risk_events" }

type Score struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	ClientEmail  string  `json:"clientEmail" gorm:"uniqueIndex;not null"`
	Score        int     `json:"score"`
	Confidence   float64 `json:"confidence"`
	State        string  `json:"state"`
	Evidence     int     `json:"evidence"`
	CalculatedAt int64   `json:"calculatedAt"`
	Version      int     `json:"version"`
}

func (Score) TableName() string { return "fork_risk_scores" }

type Suppression struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	ClientEmail string `json:"clientEmail" gorm:"not null;index:idx_risk_suppression_client"`
	Kind        string `json:"kind" gorm:"not null"`
	Reason      string `json:"reason,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
	ExpiresAt   int64  `json:"expiresAt"`
}

func (Suppression) TableName() string { return "fork_risk_suppressions" }

type Summary struct {
	Enabled      bool        `json:"enabled"`
	ClientEmail  string      `json:"clientEmail"`
	Score        int         `json:"score"`
	Confidence   float64     `json:"confidence"`
	State        string      `json:"state"`
	Evidence     int         `json:"evidence"`
	CalculatedAt int64       `json:"calculatedAt"`
	Events       []Event     `json:"events"`
	IPs          []IPHistory `json:"ipHistory"`
}

type RetentionSettings struct {
	IPDays    int `json:"ipDays"`
	EventDays int `json:"eventDays"`
}
