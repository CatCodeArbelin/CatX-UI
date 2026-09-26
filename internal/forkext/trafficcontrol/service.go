package trafficcontrol

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

var service = newService()

type serviceState struct {
	mu      sync.RWMutex
	enabled bool
	backend *Backend
	status  Status
}

func newService() *serviceState {
	return &serviceState{backend: NewBackend(nil, "", configuredInterfaces())}
}

func configuredInterfaces() []string {
	return strings.FieldsFunc(os.Getenv("CATX_TRAFFIC_CONTROL_INTERFACES"), func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
}

func Configure(enabled bool) {
	service.mu.Lock()
	service.enabled = enabled
	service.backend = NewBackend(nil, "", configuredInterfaces())
	service.status = Status{}
	service.mu.Unlock()
}

func SetBackendForTests(b *Backend) {
	service.mu.Lock()
	service.backend = b
	service.enabled = true
	service.mu.Unlock()
}

func DetectCapabilities(ctx context.Context) Capabilities {
	service.mu.RLock()
	b, enabled := service.backend, service.enabled
	service.mu.RUnlock()
	if !enabled {
		return Capabilities{Platform: "disabled", State: State(stateDisabled), Reason: "traffic_control.enabled is false"}
	}
	return b.Capabilities(ctx)
}

func Reconcile(ctx context.Context, rules []DesiredRule) (Status, error) {
	service.mu.RLock()
	b, enabled := service.backend, service.enabled
	service.mu.RUnlock()
	if !enabled {
		return Status{Capabilities: Capabilities{Platform: "disabled", State: State(stateDisabled), Reason: "traffic_control.enabled is false"}}, nil
	}
	status, err := b.Reconcile(ctx, rules)
	service.mu.Lock()
	service.status = status
	service.mu.Unlock()
	return status, err
}

func (s *serviceState) Capabilities(ctx context.Context) Capabilities {
	return DetectCapabilities(ctx)
}

func (s *serviceState) Reconcile(ctx context.Context, rules []DesiredRule) (Status, error) {
	return Reconcile(ctx, rules)
}

func (s *serviceState) ApplyClientLimit(ctx context.Context, rule DesiredRule) error {
	service.mu.RLock()
	b, enabled := service.backend, service.enabled
	service.mu.RUnlock()
	if !enabled {
		return ErrUnsupported
	}
	return b.ApplyClientLimit(ctx, rule)
}

func (s *serviceState) RemoveClientLimit(ctx context.Context, nodeKey, clientKey string) error {
	service.mu.RLock()
	b, enabled := service.backend, service.enabled
	service.mu.RUnlock()
	if !enabled {
		return ErrUnsupported
	}
	return b.RemoveClientLimit(ctx, nodeKey, clientKey)
}

// GlobalShaper exposes the single WP-6A substrate to fork-owned producers.
// It does not create another backend or reconciliation state store.
func GlobalShaper() Shaper { return service }

func StatusView() Status {
	service.mu.RLock()
	status, initialized := service.status, service.status.State != ""
	service.mu.RUnlock()
	if !initialized {
		status.Capabilities = DetectCapabilities(context.Background())
	}
	return status
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	g := api.Group("/traffic-control")
	g.GET("/status", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true, "obj": StatusView()}) })
	g.GET("/capabilities", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": DetectCapabilities(c.Request.Context())})
	})
	g.POST("/reconcile", func(c *gin.Context) {
		var req ReconcileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid reconcile request"})
			return
		}
		status, err := Reconcile(c.Request.Context(), req.Rules)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "obj": status, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": ReconcileResponse{Status: status}})
	})
}

func RegisterJobs(scheduler *cron.Cron) {
	if scheduler == nil {
		return
	}
	service.mu.RLock()
	enabled := service.enabled
	service.mu.RUnlock()
	if !enabled {
		return
	}
	_, _ = scheduler.AddFunc("@every 30s", func() { _, _ = Reconcile(context.Background(), nil) })
}

func ReconcileRemote(ctx context.Context, remote RemoteTransport, rules []DesiredRule) (Status, error) {
	if remote == nil {
		return Status{Capabilities: Capabilities{State: State(stateUnsupported), Reason: "remote runtime unavailable"}}, ErrUnsupported
	}
	raw, err := remote.TrafficControlCapabilities(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return Status{Capabilities: Capabilities{State: State(stateUnsupported), Reason: "remote node does not support shaping"}}, ErrUnsupported
		}
		return Status{Capabilities: Capabilities{State: State(stateDegraded), Reason: "remote capability probe failed"}, LastError: err.Error()}, err
	}
	var cap Capabilities
	if err := json.Unmarshal(raw, &cap); err != nil {
		return Status{}, err
	}
	if cap.State != State(stateReady) {
		return Status{Capabilities: cap}, ErrUnsupported
	}
	body, err := json.Marshal(ReconcileRequest{Rules: rules})
	if err != nil {
		return Status{}, err
	}
	response, err := remote.TrafficControlReconcile(ctx, body)
	if err != nil {
		return Status{Capabilities: cap, LastError: err.Error()}, err
	}
	var out ReconcileResponse
	if err := json.Unmarshal(response, &out); err != nil {
		return Status{}, err
	}
	return out.Status, nil
}
