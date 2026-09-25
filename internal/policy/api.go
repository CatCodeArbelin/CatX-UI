package policy

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var state struct {
	sync.RWMutex
	repo    *Repository
	enabled bool
}

func Configure(db *gorm.DB, enabled bool) {
	state.Lock()
	defer state.Unlock()
	state.enabled = enabled && db != nil
	if state.enabled {
		state.repo = NewRepository(db)
	} else {
		state.repo = nil
	}
}

func Enabled() bool { state.RLock(); defer state.RUnlock(); return state.enabled }
func current() (*Repository, bool) {
	state.RLock()
	defer state.RUnlock()
	return state.repo, state.enabled
}

func response(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": obj})
}
func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"success": false, "msg": msg})
}
func unavailable(c *gin.Context) { fail(c, http.StatusConflict, "policies are disabled") }
func parseID(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || n == 0 {
		fail(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return uint(n), true
}
func repoOrUnavailable(c *gin.Context) (*Repository, bool) {
	r, ok := current()
	if !ok {
		unavailable(c)
	}
	return r, ok
}
func mapError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fail(c, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
	}
}

type policyInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Spec        json.RawMessage `json:"spec"`
	Priority    *int            `json:"priority"`
	Enabled     *bool           `json:"enabled"`
}

func (in policyInput) model(id uint) (Policy, error) {
	spec := string(in.Spec)
	if len(in.Spec) == 0 {
		spec = `{}`
	}
	p := Policy{ID: id, Name: in.Name, Description: in.Description, Spec: spec, Enabled: true}
	if in.Priority != nil {
		p.Priority = *in.Priority
	}
	if in.Enabled != nil {
		p.Enabled = *in.Enabled
	}
	return p, nil
}

type assignmentInput struct {
	PolicyID   uint   `json:"policyId"`
	TargetType string `json:"targetType"`
	TargetRef  string `json:"targetRef"`
	Priority   int    `json:"priority"`
	Enabled    *bool  `json:"enabled"`
}
type overrideInput struct {
	PolicyID   uint            `json:"policyId"`
	TargetType string          `json:"targetType"`
	TargetRef  string          `json:"targetRef"`
	Scope      string          `json:"scope"`
	Value      json.RawMessage `json:"value"`
	Priority   int             `json:"priority"`
	Enabled    *bool           `json:"enabled"`
}
type temporaryInput struct {
	PolicyID   uint            `json:"policyId"`
	TargetType string          `json:"targetType"`
	TargetRef  string          `json:"targetRef"`
	Scope      string          `json:"scope"`
	Value      json.RawMessage `json:"value"`
	Priority   int             `json:"priority"`
	StartsAt   int64           `json:"startsAt"`
	ExpiresAt  int64           `json:"expiresAt"`
	Enabled    *bool           `json:"enabled"`
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/policies/status", func(c *gin.Context) { response(c, gin.H{"enabled": Enabled(), "enforcement": false}) })
	api.GET("/policies", func(c *gin.Context) {
		r, ok := current()
		if !ok {
			response(c, gin.H{"enabled": false, "items": []Policy{}})
			return
		}
		rows, err := r.ListPolicies(c.Request.Context())
		if err != nil {
			fail(c, 500, "policy data unavailable")
			return
		}
		response(c, gin.H{"enabled": true, "items": rows})
	})
	api.POST("/policies", createPolicy)
	api.GET("/policies/:id", getPolicy)
	api.PUT("/policies/:id", updatePolicy)
	api.DELETE("/policies/:id", deletePolicy)
	api.GET("/policies/assignments", listAssignments)
	api.POST("/policies/assignments", createAssignment)
	api.DELETE("/policies/assignments/:id", deleteAssignment)
	api.PUT("/policies/assignments/:id", updateAssignment)
	api.GET("/policies/overrides", listOverrides)
	api.POST("/policies/overrides", createOverride)
	api.DELETE("/policies/overrides/:id", deleteOverride)
	api.PUT("/policies/overrides/:id", updateOverride)
	api.GET("/policies/temporary-overrides", listTemporary)
	api.POST("/policies/temporary-overrides", createTemporary)
	api.DELETE("/policies/temporary-overrides/:id", deleteTemporary)
	api.PUT("/policies/temporary-overrides/:id", updateTemporary)
	api.GET("/policies/resolve", resolve)
}

func createPolicy(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	var in policyInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid policy")
		return
	}
	p, _ := in.model(0)
	if err := r.CreatePolicy(c.Request.Context(), &p); err != nil {
		mapError(c, err)
		return
	}
	response(c, p)
}
func getPolicy(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	p, err := r.GetPolicy(c.Request.Context(), id)
	if err != nil {
		mapError(c, err)
		return
	}
	response(c, p)
}
func updatePolicy(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in policyInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid policy")
		return
	}
	p, _ := in.model(id)
	if err := r.UpdatePolicy(c.Request.Context(), &p); err != nil {
		mapError(c, err)
		return
	}
	response(c, p)
}
func deletePolicy(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := r.DeletePolicy(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response(c, gin.H{"deleted": id})
}

func createAssignment(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	var in assignmentInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid assignment")
		return
	}
	a := PolicyAssignment{PolicyID: in.PolicyID, TargetType: in.TargetType, TargetRef: in.TargetRef, Priority: in.Priority, Enabled: true}
	if in.Enabled != nil {
		a.Enabled = *in.Enabled
	}
	if err := r.CreateAssignment(c.Request.Context(), &a); err != nil {
		mapError(c, err)
		return
	}
	response(c, a)
}
func listAssignments(c *gin.Context) {
	r, ok := current()
	if !ok {
		response(c, gin.H{"enabled": false, "items": []PolicyAssignment{}})
		return
	}
	rows, err := r.ListAssignments(c.Request.Context())
	if err != nil {
		fail(c, 500, "policy data unavailable")
		return
	}
	response(c, gin.H{"enabled": true, "items": rows})
}
func deleteAssignment(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := r.DeleteAssignment(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response(c, gin.H{"deleted": id})
}
func updateAssignment(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in assignmentInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid assignment")
		return
	}
	a := PolicyAssignment{ID: id, PolicyID: in.PolicyID, TargetType: in.TargetType, TargetRef: in.TargetRef, Priority: in.Priority, Enabled: true}
	if in.Enabled != nil {
		a.Enabled = *in.Enabled
	}
	if err := r.UpdateAssignment(c.Request.Context(), &a); err != nil {
		mapError(c, err)
		return
	}
	response(c, a)
}

func createOverride(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	var in overrideInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid override")
		return
	}
	o := PolicyOverride{PolicyID: in.PolicyID, TargetType: in.TargetType, TargetRef: in.TargetRef, Scope: in.Scope, Value: string(in.Value), Priority: in.Priority, Enabled: true}
	if in.Enabled != nil {
		o.Enabled = *in.Enabled
	}
	if err := r.CreateOverride(c.Request.Context(), &o); err != nil {
		mapError(c, err)
		return
	}
	response(c, o)
}
func listOverrides(c *gin.Context) {
	r, ok := current()
	if !ok {
		response(c, gin.H{"enabled": false, "items": []PolicyOverride{}})
		return
	}
	rows, err := r.ListOverrides(c.Request.Context())
	if err != nil {
		fail(c, 500, "policy data unavailable")
		return
	}
	response(c, gin.H{"enabled": true, "items": rows})
}
func deleteOverride(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := r.DeleteOverride(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response(c, gin.H{"deleted": id})
}
func updateOverride(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in overrideInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid override")
		return
	}
	o := PolicyOverride{ID: id, PolicyID: in.PolicyID, TargetType: in.TargetType, TargetRef: in.TargetRef, Scope: in.Scope, Value: string(in.Value), Priority: in.Priority, Enabled: true}
	if in.Enabled != nil {
		o.Enabled = *in.Enabled
	}
	if err := r.UpdateOverride(c.Request.Context(), &o); err != nil {
		mapError(c, err)
		return
	}
	response(c, o)
}

func createTemporary(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	var in temporaryInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid temporary override")
		return
	}
	o := TemporaryOverride{PolicyID: in.PolicyID, TargetType: in.TargetType, TargetRef: in.TargetRef, Scope: in.Scope, Value: string(in.Value), Priority: in.Priority, StartsAt: in.StartsAt, ExpiresAt: in.ExpiresAt, Enabled: true}
	if in.Enabled != nil {
		o.Enabled = *in.Enabled
	}
	if err := r.CreateTemporaryOverride(c.Request.Context(), &o); err != nil {
		mapError(c, err)
		return
	}
	response(c, o)
}
func listTemporary(c *gin.Context) {
	r, ok := current()
	if !ok {
		response(c, gin.H{"enabled": false, "items": []TemporaryOverride{}})
		return
	}
	rows, err := r.ListTemporaryOverrides(c.Request.Context(), time.Now().UnixMilli())
	if err != nil {
		fail(c, 500, "policy data unavailable")
		return
	}
	response(c, gin.H{"enabled": true, "items": rows})
}
func deleteTemporary(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := r.DeleteTemporaryOverride(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response(c, gin.H{"deleted": id})
}
func updateTemporary(c *gin.Context) {
	r, ok := repoOrUnavailable(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in temporaryInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid temporary override")
		return
	}
	o := TemporaryOverride{ID: id, PolicyID: in.PolicyID, TargetType: in.TargetType, TargetRef: in.TargetRef, Scope: in.Scope, Value: string(in.Value), Priority: in.Priority, StartsAt: in.StartsAt, ExpiresAt: in.ExpiresAt, Enabled: true}
	if in.Enabled != nil {
		o.Enabled = *in.Enabled
	}
	if err := r.UpdateTemporaryOverride(c.Request.Context(), &o); err != nil {
		mapError(c, err)
		return
	}
	response(c, o)
}

func resolve(c *gin.Context) {
	r, ok := current()
	if !ok {
		response(c, gin.H{"enabled": false, "items": []resolvedCandidate{}})
		return
	}
	at := int64(0)
	if raw := c.Query("at"); raw != "" {
		var err error
		at, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || at <= 0 {
			fail(c, 400, "invalid at")
			return
		}
	}
	result, err := r.Resolve(c.Request.Context(), c.Query("clientEmail"), c.Query("groupName"), at)
	if err != nil {
		fail(c, 500, "policy data unavailable")
		return
	}
	response(c, gin.H{"enabled": true, "enforcement": false, "resolved": result})
}
