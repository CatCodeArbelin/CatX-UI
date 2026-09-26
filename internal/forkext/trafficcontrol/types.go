package trafficcontrol

import (
	"context"
	"encoding/json"
	"errors"
)

const (
	featureName      = "traffic_control"
	stateUnsupported = "unsupported"
	stateReady       = "ready"
	stateDegraded    = "degraded"
	stateDisabled    = "disabled"
)

var ErrUnsupported = errors.New("traffic shaping is unsupported")

type State string

type Capabilities struct {
	Platform       string   `json:"platform"`
	State          State    `json:"state"`
	Reason         string   `json:"reason,omitempty"`
	Interfaces     []string `json:"interfaces,omitempty"`
	Tc             bool     `json:"tc"`
	Nftables       bool     `json:"nftables"`
	ConntrackMarks bool     `json:"conntrackMarks"`
	IFB            bool     `json:"ifb"`
	NetAdmin       bool     `json:"netAdmin"`
	// UserAttribution is deliberately false until an Xray-native, semantics-
	// preserving per-user socket-mark path is proven and enabled.
	UserAttribution bool `json:"userAttribution"`
}

type Status struct {
	Capabilities
	AppliedInterfaces []string `json:"appliedInterfaces,omitempty"`
	Generation        uint64   `json:"generation"`
	LastError         string   `json:"lastError,omitempty"`
}

// DesiredRule is deliberately a substrate contract. WP-6B supplies policy
// values through this boundary; WP-6A does not persist or invent them.
type DesiredRule struct {
	NodeKey         string   `json:"nodeKey"`
	ClientKey       string   `json:"clientKey"`
	Interface       string   `json:"interface"`
	Mark            uint32   `json:"mark"`
	UploadRateBps   uint64   `json:"uploadRateBps"`
	DownloadRateBps uint64   `json:"downloadRateBps"`
	Selectors       []string `json:"selectors,omitempty"`
}

type ReconcileRequest struct {
	Rules []DesiredRule `json:"rules"`
}

type ReconcileResponse struct {
	Status Status `json:"status"`
}

// Shaper is the stable substrate boundary. Policy producers own desired
// values; this package owns capability checks and kernel reconciliation.
type Shaper interface {
	Capabilities(context.Context) Capabilities
	ApplyClientLimit(context.Context, DesiredRule) error
	RemoveClientLimit(context.Context, string, string) error
	Reconcile(context.Context, []DesiredRule) (Status, error)
}

type RemoteTransport interface {
	TrafficControlCapabilities(context.Context) (json.RawMessage, error)
	TrafficControlReconcile(context.Context, json.RawMessage) (json.RawMessage, error)
}
