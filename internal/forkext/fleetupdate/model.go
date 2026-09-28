package fleetupdate

import "time"

const (
	StatePending        = "pending"
	StatePreflight      = "preflight"
	StateReady          = "ready"
	StateBlocked        = "blocked"
	StateDispatching    = "dispatching"
	StateWaitingRestart = "waiting_restart"
	StateHealthCheck    = "health_check"
	StateSoaking        = "soaking"
	StateSucceeded      = "succeeded"
	StateFailed         = "failed"
	StateAborted        = "aborted"
	StateUnknown        = "unknown"
)

const (
	ReasonOffline            = "offline"
	ReasonStaleHeartbeat     = "stale_heartbeat"
	ReasonDisabled           = "disabled"
	ReasonUnsupported        = "unsupported"
	ReasonAlreadyCurrent     = "already_current"
	ReasonDegraded           = "degraded"
	ReasonDeleted            = "deleted"
	ReasonTransitive         = "transitive_read_only"
	ReasonConcurrentCampaign = "concurrent_campaign"
	ReasonExecutionDisabled  = "execution_disabled"
	ReasonHealthTimeout      = "health_timeout"
)

type Campaign struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" gorm:"size:160;not null"`
	Channel           string    `json:"channel" gorm:"size:16;not null;index"`
	ReleaseTag        string    `json:"releaseTag" gorm:"size:64;not null"`
	ReleaseAPIURL     string    `json:"releaseApiUrl" gorm:"size:512;not null"`
	ReleaseHTMLURL    string    `json:"releaseHtmlUrl" gorm:"size:512;not null"`
	DryRun            bool      `json:"dryRun" gorm:"not null"`
	CanaryCount       int       `json:"canaryCount" gorm:"not null;default:0"`
	BatchSize         int       `json:"batchSize" gorm:"not null;default:1"`
	MaxParallel       int       `json:"maxParallel" gorm:"not null;default:1"`
	HealthTimeoutSecs int       `json:"healthTimeoutSecs" gorm:"not null;default:300"`
	SoakSeconds       int       `json:"soakSeconds" gorm:"not null;default:0"`
	StopOnFailure     bool      `json:"stopOnFailure" gorm:"not null;default:true"`
	State             string    `json:"state" gorm:"size:32;not null;index"`
	Error             string    `json:"error,omitempty" gorm:"size:500"`
	LeaseOwner        string    `json:"-" gorm:"size:120;index"`
	LeaseUntil        time.Time `json:"-" gorm:"index"`
	Revision          int64     `json:"revision" gorm:"not null;default:1"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (Campaign) TableName() string { return "fork_update_campaigns" }

type Target struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	CampaignID       uint       `json:"campaignId" gorm:"not null;uniqueIndex:idx_update_campaign_node,priority:1;index"`
	NodeID           int        `json:"nodeId" gorm:"not null;uniqueIndex:idx_update_campaign_node,priority:2;index"`
	NodeGuid         string     `json:"nodeGuid" gorm:"size:160"`
	NodeName         string     `json:"nodeName" gorm:"size:160;not null"`
	NodeAddress      string     `json:"-" gorm:"size:255"`
	NodePort         int        `json:"-"`
	NodeScheme       string     `json:"-" gorm:"size:16"`
	InitialVersion   string     `json:"initialVersion" gorm:"size:64"`
	State            string     `json:"state" gorm:"size:32;not null;index"`
	BlockedReason    string     `json:"blockedReason,omitempty" gorm:"size:64"`
	Error            string     `json:"error,omitempty" gorm:"size:500"`
	ObservedVersion  string     `json:"observedVersion" gorm:"size:64"`
	ObservedStatus   string     `json:"observedStatus" gorm:"size:32"`
	ObservedXray     string     `json:"observedXray" gorm:"size:32"`
	DispatchKey      string     `json:"dispatchKey" gorm:"size:100;uniqueIndex"`
	DispatchEvidence string     `json:"dispatchEvidence,omitempty" gorm:"size:500"`
	LeaseOwner       string     `json:"-" gorm:"size:120;index"`
	LeaseUntil       time.Time  `json:"-" gorm:"index"`
	Attempts         int        `json:"attempts" gorm:"not null;default:0"`
	DispatchAt       *time.Time `json:"dispatchAt,omitempty"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (Target) TableName() string { return "fork_update_campaign_targets" }

type ReleaseSnapshot struct {
	Channel          string `json:"channel"`
	Tag              string `json:"tag"`
	APIURL           string `json:"apiUrl"`
	HTMLURL          string `json:"htmlUrl"`
	ChecksumVerified bool   `json:"checksumVerified"`
}

type NodeSnapshot struct {
	ID            int
	Guid          string
	Name          string
	Address       string
	Port          int
	Scheme        string
	Enabled       bool
	Transitive    bool
	Status        string
	LastHeartbeat int64
	PanelVersion  string
	XrayState     string
}

type PlanRequest struct {
	Name              string `json:"name"`
	Channel           string `json:"channel"`
	NodeIDs           []int  `json:"nodeIds"`
	DryRun            bool   `json:"dryRun"`
	CanaryCount       int    `json:"canaryCount"`
	BatchSize         int    `json:"batchSize"`
	MaxParallel       int    `json:"maxParallel"`
	HealthTimeoutSecs int    `json:"healthTimeoutSecs"`
	SoakSeconds       int    `json:"soakSeconds"`
	StopOnFailure     *bool  `json:"stopOnFailure"`
}

type Plan struct {
	Campaign Campaign `json:"campaign"`
	Targets  []Target `json:"targets"`
}
