package fleetupdate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrMutationDisabled = errors.New("fleet update execution is disabled in Stage A")

type PreflightResult struct {
	Supported bool
	Reason    string
}
type (
	DispatchResult    struct{ Evidence string }
	ConvergenceResult struct {
		FreshHeartbeat   bool
		PanelVersion     string
		NodeStatus       string
		XrayState        string
		RollbackEvidence string
	}
)

type UpdateExecutor interface {
	Preflight(context.Context, NodeSnapshot, ReleaseSnapshot) (PreflightResult, error)
	Dispatch(context.Context, NodeSnapshot, ReleaseSnapshot) (DispatchResult, error)
	Reconcile(context.Context, NodeSnapshot, ReleaseSnapshot) (ConvergenceResult, error)
}

// DisabledExecutor is the safe default until the real mutation gate is
// explicitly enabled in a later work package.
type DisabledExecutor struct{}

func (DisabledExecutor) Preflight(context.Context, NodeSnapshot, ReleaseSnapshot) (PreflightResult, error) {
	return PreflightResult{Supported: true}, nil
}

func (DisabledExecutor) Dispatch(context.Context, NodeSnapshot, ReleaseSnapshot) (DispatchResult, error) {
	return DispatchResult{}, ErrMutationDisabled
}

func (DisabledExecutor) Reconcile(_ context.Context, node NodeSnapshot, release ReleaseSnapshot) (ConvergenceResult, error) {
	return ConvergenceResult{FreshHeartbeat: node.LastHeartbeat > 0, PanelVersion: node.PanelVersion, NodeStatus: node.Status, XrayState: node.XrayState}, nil
}

// FakeExecutor is a complete deterministic executor for tests and dry-run
// simulations. It records every dispatch and can simulate health convergence.
type FakeExecutor struct {
	mu          sync.Mutex
	Dispatches  []int
	Nodes       map[int]ConvergenceResult
	DispatchErr map[int]error
}

func (f *FakeExecutor) Preflight(context.Context, NodeSnapshot, ReleaseSnapshot) (PreflightResult, error) {
	return PreflightResult{Supported: true}, nil
}

func (f *FakeExecutor) Dispatch(_ context.Context, n NodeSnapshot, _ ReleaseSnapshot) (DispatchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Dispatches = append(f.Dispatches, n.ID)
	if err := f.DispatchErr[n.ID]; err != nil {
		return DispatchResult{}, err
	}
	return DispatchResult{Evidence: fmt.Sprintf("fake-dispatch-node-%d", n.ID)}, nil
}

func (f *FakeExecutor) Reconcile(_ context.Context, n NodeSnapshot, _ ReleaseSnapshot) (ConvergenceResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.Nodes[n.ID]; ok {
		return r, nil
	}
	return ConvergenceResult{}, nil
}

func healthyConvergence(c ConvergenceResult, release ReleaseSnapshot) bool {
	if !c.FreshHeartbeat || c.PanelVersion != release.Tag {
		return false
	}
	if c.NodeStatus != "online" {
		return false
	}
	return c.XrayState == "running" || c.XrayState == "healthy" || c.XrayState == "ok"
}
func leaseExpired(until time.Time) bool { return until.IsZero() || time.Now().After(until) }
