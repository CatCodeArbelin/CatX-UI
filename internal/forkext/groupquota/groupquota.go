package groupquota

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const Scale int64 = 1_000_000

const (
	periodNever   = "never"
	periodHourly  = "hourly"
	periodDaily   = "daily"
	periodWeekly  = "weekly"
	periodMonthly = "monthly"
)

// State contains quota configuration and carry only. It intentionally does
// not contain traffic counters: client_traffics remains authoritative.
type State struct {
	ID                   uint   `gorm:"primaryKey"`
	GroupName            string `gorm:"uniqueIndex;not null"`
	QuotaBytes           int64  `gorm:"not null;default:0"`
	CarryUp              int64  `gorm:"not null;default:0"`
	CarryDown            int64  `gorm:"not null;default:0"`
	ActiveMultiplierPPM  int64  `gorm:"not null;default:1000000"`
	PendingMultiplierPPM int64  `gorm:"not null;default:1000000"`
	ResetPeriod          string `gorm:"not null;default:never"`
	ResetDay             int    `gorm:"not null;default:1"`
	LastResetAt          int64  `gorm:"not null;default:0"`
	Depleted             bool   `gorm:"not null;default:false"`
	UpdatedAt            int64  `gorm:"not null;default:0"`
}

func (State) TableName() string { return "fork_group_quota_states" }

type Membership struct {
	ID              uint   `gorm:"primaryKey"`
	GroupName       string `gorm:"uniqueIndex:idx_fork_group_quota_member,priority:1;not null"`
	ClientEmail     string `gorm:"uniqueIndex:idx_fork_group_quota_member,priority:2;not null"`
	BaseUp          int64  `gorm:"not null;default:0"`
	BaseDown        int64  `gorm:"not null;default:0"`
	DisabledByGroup bool   `gorm:"not null;default:false"`
	UpdatedAt       int64  `gorm:"not null;default:0"`
}

func (Membership) TableName() string { return "fork_group_quota_members" }

type GroupView struct {
	GroupName            string `json:"groupName"`
	QuotaBytes           int64  `json:"quotaBytes"`
	UsedBytes            int64  `json:"usedBytes"`
	RemainingBytes       int64  `json:"remainingBytes"`
	QuotaEnabled         bool   `json:"quotaEnabled"`
	Depleted             bool   `json:"depleted"`
	ActiveMultiplierPPM  int64  `json:"activeMultiplierPpm"`
	PendingMultiplierPPM int64  `json:"pendingMultiplierPpm"`
	ResetPeriod          string `json:"resetPeriod"`
	ResetDay             int    `json:"resetDay"`
}

var (
	db              *gorm.DB
	enabled         bool
	restartCallback func()
)

func Configure(database *gorm.DB, on bool) { db, enabled = database, on }
func SetRestartCallback(fn func())         { restartCallback = fn }

func Migrate(database *gorm.DB) error {
	if database == nil {
		return errors.New("group quota migration requires database")
	}
	return database.AutoMigrate(&State{}, &Membership{})
}

func Models() []any { return []any{&State{}, &Membership{}} }

func Enabled() bool { return enabled && db != nil }

func validateMultiplier(v int64) error {
	if v < 1 || v > 100*Scale {
		return fmt.Errorf("multiplier must be between 0.000001 and 100")
	}
	return nil
}

func clamp(v int64) int64 {
	if v < 0 || v > 9_000_000_000_000_000_000 {
		return 9_000_000_000_000_000_000
	}
	return v
}

func ceilMul(v, ppm int64) int64 {
	if v <= 0 || ppm <= 0 {
		return 0
	}
	if v > math.MaxInt64/ppm {
		return 9_000_000_000_000_000_000
	}
	return clamp((v*ppm + Scale - 1) / Scale)
}

func lockState(tx *gorm.DB, names []string) ([]State, error) {
	sort.Strings(names)
	var out []State
	for _, name := range names {
		var row State
		q := tx.Where("group_name = ?", name)
		if tx.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func ensureMembership(tx *gorm.DB, group, email string, baseUp, baseDown int64) error {
	if group == "" || email == "" {
		return nil
	}
	var row Membership
	if err := tx.Where("group_name = ? AND client_email = ?", group, email).First(&row).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&Membership{GroupName: group, ClientEmail: email, BaseUp: baseUp, BaseDown: baseDown, UpdatedAt: time.Now().UnixMilli()}).Error
}

// ApplyClientDeltas runs in the same transaction as the authoritative
// client_traffics increment. It creates missing membership baselines from the
// pre-delta value, so the current poll is charged exactly once.
func ApplyClientDeltas(tx *gorm.DB, traffics []*xray.ClientTraffic) error {
	if !Enabled() || len(traffics) == 0 {
		return nil
	}
	groups := make([]string, 0, len(traffics))
	seenGroups := make(map[string]struct{}, len(traffics))
	for _, t := range traffics {
		if t == nil || t.Email == "" {
			continue
		}
		var c model.ClientRecord
		if err := tx.Where("email = ?", t.Email).First(&c).Error; err == nil && c.Group != "" {
			if _, seen := seenGroups[c.Group]; !seen {
				seenGroups[c.Group] = struct{}{}
				groups = append(groups, c.Group)
			}
		}
	}
	if _, err := lockState(tx, groups); err != nil {
		return err
	}
	for _, t := range traffics {
		if t == nil || t.Email == "" || (t.Up == 0 && t.Down == 0) {
			continue
		}
		var c model.ClientRecord
		if err := tx.Where("email = ?", t.Email).First(&c).Error; err != nil {
			continue
		}
		var row xray.ClientTraffic
		if err := tx.Where("email = ?", t.Email).First(&row).Error; err != nil {
			return err
		}
		if c.Group != "" {
			if err := ensureMembership(tx, c.Group, c.Email, max(row.Up-t.Up, 0), max(row.Down-t.Down, 0)); err != nil {
				return err
			}
		}
	}
	return nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func memberTotals(tx *gorm.DB, group string) (int64, int64, error) {
	var row struct{ Up, Down int64 }
	err := tx.Table("fork_group_quota_members m").
		Select("COALESCE(SUM(CASE WHEN ct.up > m.base_up THEN ct.up - m.base_up ELSE 0 END),0) AS up, COALESCE(SUM(CASE WHEN ct.down > m.base_down THEN ct.down - m.base_down ELSE 0 END),0) AS down").
		Joins("JOIN client_traffics ct ON ct.email = m.client_email").
		Where("m.group_name = ?", group).Scan(&row).Error
	return row.Up, row.Down, err
}

func viewTx(tx *gorm.DB, state State) (GroupView, error) {
	up, down, err := memberTotals(tx, state.GroupName)
	if err != nil {
		return GroupView{}, err
	}
	up = clamp(up + state.CarryUp)
	down = clamp(down + state.CarryDown)
	used := clamp(ceilMul(up, state.ActiveMultiplierPPM) + ceilMul(down, state.ActiveMultiplierPPM))
	remaining := int64(0)
	if state.QuotaBytes <= 0 || used < state.QuotaBytes {
		remaining = state.QuotaBytes - used
		if state.QuotaBytes <= 0 {
			remaining = 0
		}
	}
	return GroupView{GroupName: state.GroupName, QuotaBytes: state.QuotaBytes, UsedBytes: used, RemainingBytes: remaining, QuotaEnabled: state.QuotaBytes > 0, Depleted: state.QuotaBytes > 0 && used >= state.QuotaBytes, ActiveMultiplierPPM: state.ActiveMultiplierPPM, PendingMultiplierPPM: state.PendingMultiplierPPM, ResetPeriod: state.ResetPeriod, ResetDay: state.ResetDay}, nil
}

func View(tx *gorm.DB, group string) (GroupView, error) {
	if !Enabled() {
		return GroupView{}, gorm.ErrRecordNotFound
	}
	var state State
	if err := tx.Where("group_name = ?", strings.TrimSpace(group)).First(&state).Error; err != nil {
		return GroupView{}, err
	}
	return viewTx(tx, state)
}

func ListViews(tx *gorm.DB) ([]GroupView, error) {
	if !Enabled() {
		return nil, nil
	}
	var states []State
	if err := tx.Order("group_name ASC").Find(&states).Error; err != nil {
		return nil, err
	}
	out := make([]GroupView, 0, len(states))
	for _, state := range states {
		v, err := viewTx(tx, state)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func UpsertConfig(tx *gorm.DB, group string, quota, pendingMultiplier int64, period string, day int) error {
	if !Enabled() {
		return errors.New("group quota is disabled")
	}
	group = strings.TrimSpace(group)
	if group == "" || quota < 0 {
		return errors.New("invalid group quota configuration")
	}
	if err := validateMultiplier(pendingMultiplier); err != nil {
		return err
	}
	if period != periodNever && period != periodHourly && period != periodDaily && period != periodWeekly && period != periodMonthly {
		return errors.New("invalid group quota reset period")
	}
	if day < 1 || day > 31 {
		return errors.New("group quota reset day must be between 1 and 31")
	}
	now := time.Now().UnixMilli()
	var row State
	err := tx.Where("group_name = ?", group).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.Create(&State{GroupName: group, QuotaBytes: quota, ActiveMultiplierPPM: Scale, PendingMultiplierPPM: pendingMultiplier, ResetPeriod: period, ResetDay: day, UpdatedAt: now}).Error; err != nil {
			return err
		}
		return syncMemberships(tx, group)
	}
	if err != nil {
		return err
	}
	if _, err := lockState(tx, []string{group}); err != nil {
		return err
	}
	if err := tx.Model(&row).Updates(map[string]any{"quota_bytes": quota, "pending_multiplier_ppm": pendingMultiplier, "reset_period": period, "reset_day": day, "updated_at": now}).Error; err != nil {
		return err
	}
	if err := syncMemberships(tx, group); err != nil {
		return err
	}
	return recoverGroup(tx, group)
}

func syncMemberships(tx *gorm.DB, group string) error {
	var clients []struct {
		Email    string
		Up, Down int64
	}
	if err := tx.Table("clients c").Select("c.email, COALESCE(ct.up,0) AS up, COALESCE(ct.down,0) AS down").Joins("LEFT JOIN client_traffics ct ON ct.email = c.email").Where("c.group_name = ?", group).Find(&clients).Error; err != nil {
		return err
	}
	for _, c := range clients {
		if err := ensureMembership(tx, group, c.Email, c.Up, c.Down); err != nil {
			return err
		}
	}
	return nil
}

func Reset(tx *gorm.DB, group string) ([]string, error) {
	if !Enabled() {
		return nil, nil
	}
	states, err := lockState(tx, []string{group})
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	state := states[0]
	if err := tx.Where("group_name = ?", group).First(&state).Error; err != nil {
		return nil, err
	}
	var members []Membership
	if err := tx.Where("group_name = ?", group).Find(&members).Error; err != nil {
		return nil, err
	}
	owned := make([]string, 0)
	for _, m := range members {
		var ct xray.ClientTraffic
		if err := tx.Where("email = ?", m.ClientEmail).First(&ct).Error; err == nil {
			if m.DisabledByGroup {
				owned = append(owned, m.ClientEmail)
			}
			if err := tx.Model(&m).Updates(map[string]any{"base_up": ct.Up, "base_down": ct.Down, "disabled_by_group": false, "updated_at": time.Now().UnixMilli()}).Error; err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Model(&state).Updates(map[string]any{"carry_up": 0, "carry_down": 0, "active_multiplier_ppm": state.PendingMultiplierPPM, "depleted": false, "last_reset_at": time.Now().UnixMilli(), "updated_at": time.Now().UnixMilli()}).Error; err != nil {
		return nil, err
	}
	return owned, nil
}

// DepletedEmails evaluates quota state after authoritative counters have been
// updated. It records ownership but leaves lifecycle/runtime application to
// the upstream disable/reconcile path.
func DepletedEmails(tx *gorm.DB) ([]string, error) {
	if !Enabled() {
		return nil, nil
	}
	var names []string
	if err := tx.Model(&State{}).Order("group_name ASC").Pluck("group_name", &names).Error; err != nil {
		return nil, err
	}
	states, err := lockState(tx, names)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for _, state := range states {
		v, err := viewTx(tx, state)
		if err != nil {
			return nil, err
		}
		if err := tx.Model(&state).Update("depleted", v.Depleted).Error; err != nil {
			return nil, err
		}
		if !v.Depleted {
			continue
		}
		var members []Membership
		if err := tx.Where("group_name = ?", state.GroupName).Order("client_email ASC").Find(&members).Error; err != nil {
			return nil, err
		}
		for _, m := range members {
			var ct xray.ClientTraffic
			if err := tx.Where("email = ?", m.ClientEmail).First(&ct).Error; err != nil || !ct.Enable {
				continue
			}
			if !m.DisabledByGroup {
				if err := tx.Model(&m).Update("disabled_by_group", true).Error; err != nil {
					return nil, err
				}
			}
			if !seen[m.ClientEmail] {
				seen[m.ClientEmail] = true
				out = append(out, m.ClientEmail)
			}
		}
	}
	return out, nil
}

// recoverGroup releases only ownership established by this group quota. It
// deliberately does not touch clients lacking the marker, so unrelated manual,
// individual-quota, expiry, policy, or quarantine disables remain disabled.
func recoverGroup(tx *gorm.DB, group string) error {
	state, err := func() (State, error) {
		rows, e := lockState(tx, []string{group})
		if e != nil || len(rows) == 0 {
			return State{}, e
		}
		return rows[0], nil
	}()
	if err != nil {
		return err
	}
	v, err := viewTx(tx, state)
	if err != nil || v.Depleted {
		return err
	}
	var members []Membership
	if err := tx.Where("group_name = ? AND disabled_by_group = ?", group, true).Find(&members).Error; err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	emails := make([]string, 0, len(members))
	for _, m := range members {
		emails = append(emails, m.ClientEmail)
	}
	if err := tx.Model(&Membership{}).Where("group_name = ? AND disabled_by_group = ?", group, true).Update("disabled_by_group", false).Error; err != nil {
		return err
	}
	if err := tx.Model(&xray.ClientTraffic{}).Where("email IN ?", emails).Update("enable", true).Error; err != nil {
		return err
	}
	return tx.Model(&model.ClientRecord{}).Where("email IN ?", emails).Update("enable", true).Error
}

func IsBlocked(tx *gorm.DB, email string) (bool, error) {
	if !Enabled() {
		return false, nil
	}
	var m Membership
	if err := tx.Where("client_email = ? AND disabled_by_group = ?", email, true).First(&m).Error; err == nil {
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	var c model.ClientRecord
	if err := tx.Where("email = ?", email).First(&c).Error; err != nil {
		return false, nil
	}
	if c.Group == "" {
		return false, nil
	}
	var state State
	if err := tx.Where("group_name = ?", c.Group).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	v, err := viewTx(tx, state)
	return v.Depleted, err
}

func IsGroupDepleted(tx *gorm.DB, group string) (bool, error) {
	if !Enabled() || strings.TrimSpace(group) == "" {
		return false, nil
	}
	var state State
	if err := tx.Where("group_name = ?", strings.TrimSpace(group)).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	v, err := viewTx(tx, state)
	return v.Depleted, err
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	g := api.Group("/clients/groups/quota")
	g.GET("/:name", func(c *gin.Context) {
		v, err := View(db, c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": v})
	})
	g.PUT("/:name", func(c *gin.Context) {
		var req struct {
			QuotaBytes    int64  `json:"quotaBytes"`
			MultiplierPPM int64  `json:"multiplierPpm"`
			ResetPeriod   string `json:"resetPeriod"`
			ResetDay      int    `json:"resetDay"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		if req.MultiplierPPM == 0 {
			req.MultiplierPPM = Scale
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := UpsertConfig(tx, c.Param("name"), req.QuotaBytes, req.MultiplierPPM, req.ResetPeriod, req.ResetDay); err != nil {
				return err
			}
			return ReconcileEnabled(tx)
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		if restartCallback != nil {
			restartCallback()
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"name": c.Param("name")}})
	})
	g.POST("/:name/reset", func(c *gin.Context) {
		var owned []string
		err := db.Transaction(func(tx *gorm.DB) error {
			var e error
			owned, e = Reset(tx, c.Param("name"))
			if e != nil {
				return e
			}
			if len(owned) > 0 {
				if e = tx.Model(&xray.ClientTraffic{}).Where("email IN ?", owned).Updates(map[string]any{"enable": true}).Error; e != nil {
					return e
				}
				if e = tx.Model(&model.ClientRecord{}).Where("email IN ?", owned).Updates(map[string]any{"enable": true}).Error; e != nil {
					return e
				}
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		if len(owned) > 0 && restartCallback != nil {
			restartCallback()
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"name": c.Param("name"), "reenabled": owned}})
	})
}

// PreserveReset records consumed usage before the caller changes the
// authoritative counter. newUp/newDown are the post-operation values.
func PreserveReset(tx *gorm.DB, email string, newUp, newDown int64) error {
	if !Enabled() {
		return nil
	}
	var c model.ClientRecord
	if err := tx.Where("email = ?", email).First(&c).Error; err != nil {
		return err
	}
	if c.Group == "" {
		return nil
	}
	states, err := lockState(tx, []string{c.Group})
	if err != nil {
		return err
	}
	if len(states) == 0 {
		return nil
	}
	state := states[0]
	var m Membership
	if err := tx.Where("group_name = ? AND client_email = ?", c.Group, email).First(&m).Error; err != nil {
		return err
	}
	var old xray.ClientTraffic
	if err := tx.Where("email = ?", email).First(&old).Error; err != nil {
		return err
	}
	if err := tx.Model(&state).Updates(map[string]any{"carry_up": gorm.Expr("carry_up + ?", max(old.Up-m.BaseUp, 0)), "carry_down": gorm.Expr("carry_down + ?", max(old.Down-m.BaseDown, 0)), "updated_at": time.Now().UnixMilli()}).Error; err != nil {
		return err
	}
	// The caller is about to write the authoritative post-operation counter.
	// Rebase to that value, never to the pre-reset value, so the consumed
	// amount now held in carry cannot be counted again or disappear.
	return tx.Model(&m).Updates(map[string]any{
		"base_up":    newUp,
		"base_down":  newDown,
		"updated_at": time.Now().UnixMilli(),
	}).Error
}

func RebaselineClient(tx *gorm.DB, email string, newUp, newDown int64) error {
	if !Enabled() {
		return nil
	}
	var c model.ClientRecord
	if err := tx.Where("email = ?", email).First(&c).Error; err != nil {
		return err
	}
	if c.Group == "" {
		return nil
	}
	var old xray.ClientTraffic
	if err := tx.Where("email = ?", email).First(&old).Error; err != nil {
		return err
	}
	if newUp < old.Up || newDown < old.Down {
		return PreserveReset(tx, email, newUp, newDown)
	}
	return nil
}

func ClearOwnership(tx *gorm.DB, email string) error {
	if !Enabled() {
		return nil
	}
	return tx.Model(&Membership{}).Where("client_email = ?", email).Update("disabled_by_group", false).Error
}

func RemoveMembershipByEmail(tx *gorm.DB, email string) error {
	if !Enabled() || email == "" {
		return nil
	}
	var members []Membership
	if err := tx.Where("client_email = ?", email).Find(&members).Error; err != nil {
		return err
	}
	groups := make([]string, 0, len(members))
	for _, m := range members {
		groups = append(groups, m.GroupName)
	}
	states, err := lockState(tx, groups)
	if err != nil {
		return err
	}
	byGroup := make(map[string]State, len(states))
	for _, state := range states {
		byGroup[state.GroupName] = state
	}
	var ct xray.ClientTraffic
	if err := tx.Where("email = ?", email).First(&ct).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, m := range members {
		state, ok := byGroup[m.GroupName]
		if !ok {
			continue
		}
		if err := tx.Model(&state).Updates(map[string]any{
			"carry_up":   gorm.Expr("carry_up + ?", max(ct.Up-m.BaseUp, 0)),
			"carry_down": gorm.Expr("carry_down + ?", max(ct.Down-m.BaseDown, 0)),
			"updated_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}
	}
	return tx.Where("client_email = ?", email).Delete(&Membership{}).Error
}

func ChangeMembership(tx *gorm.DB, email, oldGroup, newGroup string) error {
	if !Enabled() {
		return nil
	}
	if oldGroup == newGroup {
		return nil
	}
	if _, err := lockState(tx, []string{oldGroup, newGroup}); err != nil {
		return err
	}
	if oldGroup != "" {
		var m Membership
		if err := tx.Where("group_name = ? AND client_email = ?", oldGroup, email).First(&m).Error; err == nil {
			var ct xray.ClientTraffic
			if err := tx.Where("email = ?", email).First(&ct).Error; err == nil {
				var state State
				if err := tx.Where("group_name = ?", oldGroup).First(&state).Error; err == nil {
					if err := tx.Model(&state).Updates(map[string]any{"carry_up": gorm.Expr("carry_up + ?", max(ct.Up-m.BaseUp, 0)), "carry_down": gorm.Expr("carry_down + ?", max(ct.Down-m.BaseDown, 0))}).Error; err != nil {
						return err
					}
				}
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Where("group_name = ? AND client_email = ?", oldGroup, email).Delete(&Membership{}).Error; err != nil {
			return err
		}
	}
	if newGroup != "" {
		var ct xray.ClientTraffic
		if err := tx.Where("email = ?", email).First(&ct).Error; err == nil {
			if err := ensureMembership(tx, newGroup, email, ct.Up, ct.Down); err != nil {
				return err
			}
		}
	}
	return nil
}

func RenameGroup(tx *gorm.DB, oldName, newName string) error {
	if !Enabled() {
		return nil
	}
	if oldName == newName {
		return nil
	}
	if _, err := lockState(tx, []string{oldName, newName}); err != nil {
		return err
	}
	if newName == "" {
		var members []Membership
		if err := tx.Where("group_name = ?", oldName).Find(&members).Error; err != nil {
			return err
		}
		for _, m := range members {
			if err := PreserveReset(tx, m.ClientEmail, 0, 0); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if err := tx.Where("group_name = ?", oldName).Delete(&Membership{}).Error; err != nil {
			return err
		}
		return tx.Where("group_name = ?", oldName).Delete(&State{}).Error
	}
	if err := tx.Model(&State{}).Where("group_name = ?", oldName).Update("group_name", newName).Error; err != nil {
		return err
	}
	return tx.Model(&Membership{}).Where("group_name = ?", oldName).Update("group_name", newName).Error
}

func ReconcileEnabled(tx *gorm.DB) error {
	if !Enabled() {
		return nil
	}
	emails, err := DepletedEmails(tx)
	if err != nil {
		return err
	}
	if len(emails) > 0 {
		if err := tx.Model(&xray.ClientTraffic{}).Where("email IN ?", emails).Update("enable", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ClientRecord{}).Where("email IN ?", emails).Update("enable", false).Error; err != nil {
			return err
		}
	}
	var groups []string
	if err := tx.Model(&State{}).Where("depleted = ?", false).Pluck("group_name", &groups).Error; err != nil {
		return err
	}
	for _, group := range groups {
		if err := recoverGroup(tx, group); err != nil {
			return err
		}
	}
	return nil
}

func RegisterJobs(scheduler *cron.Cron) {
	if !Enabled() || scheduler == nil {
		return
	}
	_, _ = scheduler.AddFunc("@every 1m", func() {
		if db == nil {
			return
		}
		var states []State
		if err := db.Where("reset_period <> ? AND quota_bytes > 0", periodNever).Find(&states).Error; err != nil {
			return
		}
		now := time.Now()
		for _, state := range states {
			if !resetDue(state, now) {
				continue
			}
			ownedCount := 0
			resetErr := db.Transaction(func(tx *gorm.DB) error {
				owned, err := Reset(tx, state.GroupName)
				if err != nil {
					return err
				}
				ownedCount = len(owned)
				if len(owned) == 0 {
					return nil
				}
				if err := tx.Model(&xray.ClientTraffic{}).Where("email IN ?", owned).Updates(map[string]any{"enable": true}).Error; err != nil {
					return err
				}
				return tx.Model(&model.ClientRecord{}).Where("email IN ?", owned).Updates(map[string]any{"enable": true}).Error
			})
			if resetErr == nil && ownedCount > 0 && restartCallback != nil {
				restartCallback()
			}
		}
	})
}

func resetDue(state State, now time.Time) bool {
	if state.LastResetAt == 0 {
		return true
	}
	last := time.UnixMilli(state.LastResetAt)
	switch state.ResetPeriod {
	case periodHourly:
		return now.After(last.Add(time.Hour))
	case periodDaily:
		return now.After(last.Add(24 * time.Hour))
	case periodWeekly:
		return now.After(last.Add(7 * 24 * time.Hour))
	case periodMonthly:
		return now.Year() != last.Year() || now.Month() != last.Month()
	default:
		return false
	}
}
