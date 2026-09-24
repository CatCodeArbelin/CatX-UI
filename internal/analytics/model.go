// Package analytics contains the fork-owned, metadata-only analytics foundation.
// It deliberately references upstream identities by their stable identifiers
// instead of copying clients, groups, nodes, or traffic counters.
package analytics

import "fmt"

const (
	BucketHourly = "hour"
	BucketDaily  = "day"

	SourceAccessLog      = "access_log"
	SourceXray           = "xray"
	SourceDNSObserver    = "dns_observer"
	SourceSNI            = "sni"
	SourceDestination    = "destination"
	ProvenanceObserved   = "observed"
	ProvenanceInferred   = "inferred"
	ProvenanceCorrelated = "correlated"
)

// DestinationObservation is a normalized metadata event. It contains no
// payload, credential, cookie, authorization, or decrypted message data.
type DestinationObservation struct {
	ID uint `json:"id" gorm:"primaryKey"`

	ObservedAt  int64  `json:"observedAt" gorm:"not null;index:idx_analytics_dest_time;index:idx_analytics_dest_client_time,priority:2;index:idx_analytics_dest_group_time,priority:2;index:idx_analytics_dest_node_time,priority:2;index:idx_analytics_dest_inbound_time,priority:2;index:idx_analytics_dest_domain_time,priority:2;index:idx_analytics_dest_ip_time,priority:2;index:idx_analytics_dest_category_time,priority:2"`
	ClientEmail string `json:"clientEmail,omitempty" gorm:"index:idx_analytics_dest_client_time,priority:1"`
	ClientGroup string `json:"clientGroup,omitempty" gorm:"index:idx_analytics_dest_group_time,priority:1"`
	NodeID      int    `json:"nodeId,omitempty" gorm:"index:idx_analytics_dest_node_time,priority:1"`
	InboundID   int    `json:"inboundId,omitempty" gorm:"index:idx_analytics_dest_inbound_time,priority:1"`

	Domain        string  `json:"domain,omitempty" gorm:"index:idx_analytics_dest_domain_time,priority:1"`
	DestinationIP string  `json:"destinationIp,omitempty" gorm:"index:idx_analytics_dest_ip_time,priority:1"`
	Port          int     `json:"port,omitempty"`
	Protocol      string  `json:"protocol,omitempty"`
	SNI           string  `json:"sni,omitempty"`
	Category      string  `json:"category,omitempty" gorm:"index:idx_analytics_dest_category_time,priority:1"`
	SessionKey    string  `json:"sessionKey,omitempty" gorm:"index:idx_analytics_dest_session"`
	Source        string  `json:"source" gorm:"not null"`
	Provenance    string  `json:"provenance" gorm:"not null"`
	Confidence    float64 `json:"confidence" gorm:"not null"`
}

func (DestinationObservation) TableName() string { return "analytics_destination_observations" }

// DNSObservation stores only DNS metadata needed for future observers.
type DNSObservation struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	ObservedAt  int64   `json:"observedAt" gorm:"not null;index:idx_analytics_dns_time;index:idx_analytics_dns_client_time,priority:2;index:idx_analytics_dns_node_time,priority:2;index:idx_analytics_dns_domain_time,priority:2"`
	ClientEmail string  `json:"clientEmail,omitempty" gorm:"index:idx_analytics_dns_client_time,priority:1"`
	NodeID      int     `json:"nodeId,omitempty" gorm:"index:idx_analytics_dns_node_time,priority:1"`
	InboundID   int     `json:"inboundId,omitempty"`
	Domain      string  `json:"domain" gorm:"not null;index:idx_analytics_dns_domain_time,priority:1"`
	RecordType  string  `json:"recordType,omitempty"`
	ResolvedIP  string  `json:"resolvedIp,omitempty"`
	Source      string  `json:"source" gorm:"not null"`
	Provenance  string  `json:"provenance" gorm:"not null"`
	Confidence  float64 `json:"confidence" gorm:"not null"`
}

func (DNSObservation) TableName() string { return "analytics_dns_observations" }

// NetworkSession correlates metadata observations without duplicating byte
// counters. Traffic remains authoritative in the upstream Xray/panel models.
type NetworkSession struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	SessionKey  string  `json:"sessionKey" gorm:"uniqueIndex;not null"`
	ClientEmail string  `json:"clientEmail,omitempty" gorm:"index:idx_analytics_session_client_time,priority:1"`
	NodeID      int     `json:"nodeId,omitempty" gorm:"index:idx_analytics_session_node_time,priority:1"`
	InboundID   int     `json:"inboundId,omitempty"`
	FirstSeen   int64   `json:"firstSeen" gorm:"not null;index:idx_analytics_session_first_seen"`
	LastSeen    int64   `json:"lastSeen" gorm:"not null;index:idx_analytics_session_last_seen"`
	Protocol    string  `json:"protocol,omitempty"`
	Source      string  `json:"source" gorm:"not null"`
	Provenance  string  `json:"provenance" gorm:"not null"`
	Confidence  float64 `json:"confidence" gorm:"not null"`
}

func (NetworkSession) TableName() string { return "analytics_network_sessions" }

// ServiceCategoryAggregate is the durable skeleton for hourly/daily history.
// Counts are derived from metadata observations and sessions, never from a
// second traffic-accounting system.
type ServiceCategoryAggregate struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	BucketStart      int64   `json:"bucketStart" gorm:"not null;uniqueIndex:idx_analytics_aggregate_bucket,priority:1;index:idx_analytics_aggregate_client_time,priority:2"`
	BucketWidth      string  `json:"bucketWidth" gorm:"not null;uniqueIndex:idx_analytics_aggregate_bucket,priority:2"`
	ClientEmail      string  `json:"clientEmail,omitempty" gorm:"uniqueIndex:idx_analytics_aggregate_bucket,priority:3;index:idx_analytics_aggregate_client_time,priority:1"`
	ClientGroup      string  `json:"clientGroup,omitempty"`
	Category         string  `json:"category,omitempty" gorm:"uniqueIndex:idx_analytics_aggregate_bucket,priority:4"`
	ObservationCount int64   `json:"observationCount" gorm:"not null;default:0"`
	SessionCount     int64   `json:"sessionCount" gorm:"not null;default:0"`
	FirstSeen        int64   `json:"firstSeen" gorm:"not null"`
	LastSeen         int64   `json:"lastSeen" gorm:"not null"`
	Source           string  `json:"source" gorm:"not null"`
	Provenance       string  `json:"provenance" gorm:"not null"`
	Confidence       float64 `json:"confidence" gorm:"not null"`
}

func (ServiceCategoryAggregate) TableName() string { return "analytics_service_category_aggregates" }

// MetadataEvent is the ingestion-facing abstraction shared by future sources.
type MetadataEvent struct {
	ObservedAt                                                 int64
	ClientEmail, ClientGroup                                   string
	NodeID, InboundID                                          int
	Domain, DestinationIP, Protocol, SNI, Category, SessionKey string
	Port                                                       int
	Source, Provenance                                         string
	Confidence                                                 float64
}

func (e MetadataEvent) Validate() error {
	if e.ObservedAt <= 0 {
		return fmt.Errorf("analytics event observed time must be positive")
	}
	if e.Domain == "" && e.DestinationIP == "" {
		return fmt.Errorf("analytics event needs a domain or destination IP")
	}
	if e.Source == "" {
		return fmt.Errorf("analytics event source is required")
	}
	if e.Provenance == "" {
		return fmt.Errorf("analytics event provenance is required")
	}
	if e.Confidence < 0 || e.Confidence > 1 {
		return fmt.Errorf("analytics event confidence must be between 0 and 1")
	}
	if e.Port < 0 || e.Port > 65535 {
		return fmt.Errorf("analytics event port is out of range")
	}
	return nil
}

func (e MetadataEvent) Observation() DestinationObservation {
	return DestinationObservation{ObservedAt: e.ObservedAt, ClientEmail: e.ClientEmail, ClientGroup: e.ClientGroup, NodeID: e.NodeID, InboundID: e.InboundID, Domain: e.Domain, DestinationIP: e.DestinationIP, Port: e.Port, Protocol: e.Protocol, SNI: e.SNI, Category: e.Category, SessionKey: e.SessionKey, Source: e.Source, Provenance: e.Provenance, Confidence: e.Confidence}
}
