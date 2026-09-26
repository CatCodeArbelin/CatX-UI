package trafficpolicy

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficcontrol"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	StateActive      = "active"
	StateThrottled   = "throttled"
	StateDisabled    = "disabled"
	StateUnsupported = "unsupported"
	StateDegraded    = "degraded"
	OwnerSoftQuota   = "soft_quota"
	OwnerNone        = "none"
)

var ErrDisabled = errors.New("traffic policy is disabled")

// Policy contains desired policy only. Traffic bytes remain authoritative in
// xray.ClientTraffic; this table never accumulates a second traffic counter.
type Policy struct {
	ID                  uint   `gorm:"primaryKey" json:"id"`
	ClientEmail         string `gorm:"uniqueIndex;not null" json:"clientEmail"`
	Enabled             bool   `gorm:"not null;default:false" json:"enabled"`
	WindowSeconds       int64  `gorm:"not null;default:0" json:"windowSeconds"`
	QuotaBytes          int64  `gorm:"not null;default:0" json:"quotaBytes"`
	ActiveUploadBps     uint64 `gorm:"not null;default:0" json:"activeUploadBps"`
	ActiveDownloadBps   uint64 `gorm:"not null;default:0" json:"activeDownloadBps"`
	ThrottleUploadBps   uint64 `gorm:"not null;default:0" json:"throttleUploadBps"`
	ThrottleDownloadBps uint64 `gorm:"not null;default:0" json:"throttleDownloadBps"`
	UpdatedAt           int64  `gorm:"not null;default:0" json:"updatedAt"`
}

func (Policy) TableName() string { return "fork_traffic_policies" }

// State stores a cumulative-counter checkpoint and lifecycle ownership. The
// current usage is always derived as client_traffics - baseline.
type State struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	ClientEmail  string `gorm:"uniqueIndex;not null" json:"clientEmail"`
	WindowStart  int64  `gorm:"not null;default:0" json:"windowStart"`
	WindowEnd    int64  `gorm:"not null;default:0" json:"windowEnd"`
	BaselineUp   int64  `gorm:"not null;default:0" json:"-"`
	BaselineDown int64  `gorm:"not null;default:0" json:"-"`
	Lifecycle    string `gorm:"not null;default:active" json:"lifecycle"`
	Owner        string `gorm:"not null;default:none" json:"owner"`
	Reason       string `gorm:"not null;default:''" json:"reason"`
	LastError    string `gorm:"not null;default:''" json:"lastError,omitempty"`
	UpdatedAt    int64  `gorm:"not null;default:0" json:"updatedAt"`
}

func (State) TableName() string { return "fork_traffic_policy_states" }

type View struct {
	Policy
	Lifecycle       string `json:"lifecycle"`
	Owner           string `json:"owner"`
	Reason          string `json:"reason"`
	WindowStart     int64  `json:"windowStart"`
	WindowEnd       int64  `json:"windowEnd"`
	UsedBytes       int64  `json:"usedBytes"`
	RemainingBytes  int64  `json:"remainingBytes"`
	Enforcement     string `json:"enforcement"`
	EnforcementNote string `json:"enforcementNote,omitempty"`
}

var service struct {
	sync.RWMutex
	db      *gorm.DB
	enabled bool
}

func Models() []any { return []any{&Policy{}, &State{}} }

func Migrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("traffic policy migration requires database")
	}
	return db.AutoMigrate(Models()...)
}

func Configure(db *gorm.DB, enabled bool) {
	service.Lock()
	service.db, service.enabled = db, enabled
	service.Unlock()
}

func getDB() (*gorm.DB, bool) {
	service.RLock()
	db, enabled := service.db, service.enabled
	service.RUnlock()
	return db, enabled && db != nil
}

func validatePolicy(p *Policy) error {
	if strings.TrimSpace(p.ClientEmail) == "" || p.WindowSeconds < 60 || p.WindowSeconds > 366*24*60*60 {
		return errors.New("client email and fixed window of 60..31622400 seconds are required")
	}
	if p.QuotaBytes < 0 {
		return errors.New("quotaBytes must not be negative")
	}
	return nil
}

func window(now, seconds int64) (int64, int64) {
	start := now - now%seconds
	return start, start + seconds
}

func clampDelta(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}

func usage(row xray.ClientTraffic, state State) int64 {
	up := clampDelta(row.Up - state.BaselineUp)
	down := clampDelta(row.Down - state.BaselineDown)
	if up > 9223372036854775807-down {
		return 9223372036854775807
	}
	return up + down
}

func lockState(tx *gorm.DB, email string) (State, error) {
	var s State
	q := tx.Where("client_email = ?", email)
	if tx.Name() == "postgres" {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return s, q.First(&s).Error
}

// ApplyDeltas is called inside the same transaction as the authoritative
// client_traffics increment. It only advances a checkpoint and lifecycle.
func ApplyDeltas(tx *gorm.DB, deltas []*xray.ClientTraffic) error {
	db, enabled := getDB()
	if !enabled || db == nil || len(deltas) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, delta := range deltas {
		if delta == nil || delta.Email == "" {
			continue
		}
		var p Policy
		if err := tx.Where("client_email = ? AND enabled = ?", delta.Email, true).First(&p).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		} else if err != nil {
			return err
		}
		var row xray.ClientTraffic
		if err := tx.Where("email = ?", delta.Email).First(&row).Error; err != nil {
			return err
		}
		state, err := lockState(tx, delta.Email)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			preUp, preDown := row.Up-delta.Up, row.Down-delta.Down
			start, end := window(now, p.WindowSeconds)
			state = State{ClientEmail: p.ClientEmail, WindowStart: start, WindowEnd: end, BaselineUp: clampDelta(preUp), BaselineDown: clampDelta(preDown), Lifecycle: StateActive, Owner: OwnerNone, UpdatedAt: now}
			if err := tx.Create(&state).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		preUp, preDown := clampDelta(row.Up-delta.Up), clampDelta(row.Down-delta.Down)
		if preUp < state.BaselineUp || preDown < state.BaselineDown {
			// Xray restart/rebaseline or a node reset moved the authoritative
			// cumulative counter backwards. Rebase before charging this poll.
			state.BaselineUp, state.BaselineDown = preUp, preDown
		}
		if now >= state.WindowEnd {
			state.WindowStart, state.WindowEnd = window(now, p.WindowSeconds)
			state.BaselineUp, state.BaselineDown = preUp, preDown
			if state.Owner == OwnerSoftQuota {
				state.Lifecycle, state.Owner, state.Reason = StateActive, OwnerNone, "window reset"
			}
		}
		if !row.Enable {
			state.Lifecycle, state.Reason = StateDisabled, "upstream client disabled"
		} else if p.QuotaBytes > 0 && usage(row, state) >= p.QuotaBytes {
			state.Lifecycle, state.Owner, state.Reason = StateThrottled, OwnerSoftQuota, "fixed-window quota reached"
		} else if state.Owner == OwnerSoftQuota {
			state.Lifecycle, state.Owner, state.Reason = StateActive, OwnerNone, "quota no longer reached"
		}
		state.UpdatedAt = now
		if err := tx.Save(&state).Error; err != nil {
			return err
		}
	}
	return nil
}

func effectiveView(tx *gorm.DB, email string) (View, error) {
	var p Policy
	if err := tx.Where("client_email = ?", email).First(&p).Error; err != nil {
		return View{}, err
	}
	var state State
	err := tx.Where("client_email = ?", email).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var row xray.ClientTraffic
		if err := tx.Where("email = ?", email).First(&row).Error; err != nil {
			return View{}, err
		}
		start, end := window(time.Now().Unix(), p.WindowSeconds)
		state = State{ClientEmail: email, WindowStart: start, WindowEnd: end, BaselineUp: row.Up, BaselineDown: row.Down, Lifecycle: StateActive, Owner: OwnerNone}
		if err := tx.Create(&state).Error; err != nil {
			return View{}, err
		}
	} else if err != nil {
		return View{}, err
	}
	var row xray.ClientTraffic
	if err := tx.Where("email = ?", email).First(&row).Error; err != nil {
		return View{}, err
	}
	used := usage(row, state)
	remaining := int64(0)
	if p.QuotaBytes > used {
		remaining = p.QuotaBytes - used
	}
	v := View{Policy: p, Lifecycle: state.Lifecycle, Owner: state.Owner, Reason: state.Reason, WindowStart: state.WindowStart, WindowEnd: state.WindowEnd, UsedBytes: used, RemainingBytes: remaining}
	if state.Lifecycle == StateDisabled {
		v.Enforcement, v.EnforcementNote = StateDisabled, state.Reason
	} else if !trafficcontrol.StatusView().UserAttribution {
		v.Enforcement, v.EnforcementNote = StateUnsupported, "kernel attribution is not proven for generic Xray users"
	} else {
		v.Enforcement = state.Lifecycle
	}
	return v, nil
}

func Get(email string) (View, error) {
	db, enabled := getDB()
	if !enabled {
		return View{}, ErrDisabled
	}
	return effectiveView(db, email)
}

func Upsert(email string, input Policy) (View, error) {
	db, enabled := getDB()
	if !enabled {
		return View{}, ErrDisabled
	}
	input.ClientEmail = strings.TrimSpace(email)
	if err := validatePolicy(&input); err != nil {
		return View{}, err
	}
	input.UpdatedAt = time.Now().UnixMilli()
	err := db.Transaction(func(tx *gorm.DB) error {
		var old Policy
		if err := tx.Where("client_email = ?", email).First(&old).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&input).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if err := tx.Model(&old).Updates(map[string]any{"enabled": input.Enabled, "window_seconds": input.WindowSeconds, "quota_bytes": input.QuotaBytes, "active_upload_bps": input.ActiveUploadBps, "active_download_bps": input.ActiveDownloadBps, "throttle_upload_bps": input.ThrottleUploadBps, "throttle_download_bps": input.ThrottleDownloadBps, "updated_at": input.UpdatedAt}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return View{}, err
	}
	return Get(email)
}

func Reset(email string) error {
	db, enabled := getDB()
	if !enabled {
		return ErrDisabled
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var row xray.ClientTraffic
		if err := tx.Where("email = ?", email).First(&row).Error; err != nil {
			return err
		}
		var s State
		if err := tx.Where("client_email = ?", email).First(&s).Error; err != nil {
			return err
		}
		var p Policy
		if err := tx.Where("client_email = ?", email).First(&p).Error; err != nil {
			return err
		}
		now := time.Now().Unix()
		s.WindowStart, s.WindowEnd = window(now, p.WindowSeconds)
		s.BaselineUp, s.BaselineDown = row.Up, row.Down
		if s.Owner == OwnerSoftQuota {
			s.Lifecycle, s.Owner, s.Reason = StateActive, OwnerNone, "operator reset"
		}
		return tx.Save(&s).Error
	})
}

func Reconcile() error {
	db, enabled := getDB()
	if !enabled {
		return nil
	}
	var policies []Policy
	if err := db.Where("enabled = ?", true).Find(&policies).Error; err != nil {
		return err
	}
	for _, p := range policies {
		if err := db.Transaction(func(tx *gorm.DB) error {
			return ApplyDeltas(tx, []*xray.ClientTraffic{{Email: p.ClientEmail}})
		}); err != nil {
			return err
		}
	}
	return nil
}

func RegisterJobs(scheduler *cron.Cron) {
	if scheduler == nil {
		return
	}
	_, _ = scheduler.AddFunc("@every 30s", func() { _ = Reconcile() })
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	g := api.Group("/traffic-control/clients")
	g.GET("/:email/policy", func(c *gin.Context) {
		v, err := Get(c.Param("email"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": v})
	})
	g.PUT("/:email/policy", func(c *gin.Context) {
		var p Policy
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		v, err := Upsert(c.Param("email"), p)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": v})
	})
	g.POST("/:email/policy/reset", func(c *gin.Context) {
		if err := Reset(c.Param("email")); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		v, err := Get(c.Param("email"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": v})
	})
}

// DesiredRule builds the WP-6A substrate state only after attribution has been
// proven. The caller supplies the node/interface/mark allocated by the
// substrate; this package never invents a routing or mark allocation path.
func DesiredRule(view View, nodeKey, iface string, mark uint32, selectors []string) (trafficcontrol.DesiredRule, error) {
	if !trafficcontrol.StatusView().UserAttribution {
		return trafficcontrol.DesiredRule{}, trafficcontrol.ErrUnsupported
	}
	upload, download := view.ActiveUploadBps, view.ActiveDownloadBps
	if view.Lifecycle == StateThrottled {
		upload, download = view.ThrottleUploadBps, view.ThrottleDownloadBps
	}
	if !view.Enabled || upload == 0 || download == 0 || nodeKey == "" || iface == "" || mark == 0 {
		return trafficcontrol.DesiredRule{}, trafficcontrol.ErrUnsupported
	}
	return trafficcontrol.DesiredRule{NodeKey: nodeKey, ClientKey: view.ClientEmail, Interface: iface, Mark: mark, UploadRateBps: upload, DownloadRateBps: download, Selectors: selectors}, nil
}

func DesiredRuleFor(view View) (trafficcontrol.DesiredRule, error) {
	return DesiredRule(view, "", "", 0, nil)
}

// ReconcileRemote is the only Stage A remote enforcement adapter. It refuses
// generic users before making a remote mutation; a future proven attribution
// implementation can supply the rule without changing the node transport.
func ReconcileRemote(ctx context.Context, remote trafficcontrol.RemoteTransport, view View) (trafficcontrol.Status, error) {
	if !trafficcontrol.StatusView().UserAttribution {
		return trafficcontrol.Status{Capabilities: trafficcontrol.Capabilities{State: "unsupported", Reason: "generic Xray user attribution is not proven"}}, trafficcontrol.ErrUnsupported
	}
	rule, err := DesiredRuleFor(view)
	if err != nil {
		return trafficcontrol.Status{}, err
	}
	return trafficcontrol.ReconcileRemote(ctx, remote, []trafficcontrol.DesiredRule{rule})
}
