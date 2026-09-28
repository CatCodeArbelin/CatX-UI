package web

import (
	"context"
	"errors"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/fleetupdate"
	"github.com/mhsanaei/3x-ui/v3/internal/web/runtime"
)

// fleetUpdateRemoteExecutor is the Stage A production capability boundary. It
// uses the existing authenticated Remote and cannot dispatch an update yet.
type fleetUpdateRemoteExecutor struct{ manager *runtime.Manager }

func (e fleetUpdateRemoteExecutor) remote(n fleetupdate.NodeSnapshot) (*runtime.Remote, error) {
	if e.manager == nil {
		return nil, errors.New("runtime manager unavailable")
	}
	return e.manager.RemoteFor(&model.Node{Id: n.ID, Guid: n.Guid, Name: n.Name, Address: n.Address, Port: n.Port, Scheme: n.Scheme, Enable: n.Enabled})
}

func (e fleetUpdateRemoteExecutor) Preflight(_ context.Context, n fleetupdate.NodeSnapshot, _ fleetupdate.ReleaseSnapshot) (fleetupdate.PreflightResult, error) {
	if _, err := e.remote(n); err != nil {
		return fleetupdate.PreflightResult{}, err
	}
	return fleetupdate.PreflightResult{Supported: true}, nil
}

func (e fleetUpdateRemoteExecutor) Dispatch(context.Context, fleetupdate.NodeSnapshot, fleetupdate.ReleaseSnapshot) (fleetupdate.DispatchResult, error) {
	return fleetupdate.DispatchResult{}, fleetupdate.ErrMutationDisabled
}

func (e fleetUpdateRemoteExecutor) Reconcile(_ context.Context, n fleetupdate.NodeSnapshot, _ fleetupdate.ReleaseSnapshot) (fleetupdate.ConvergenceResult, error) {
	return fleetupdate.ConvergenceResult{FreshHeartbeat: n.LastHeartbeat > 0, PanelVersion: n.PanelVersion, NodeStatus: n.Status, XrayState: n.XrayState}, nil
}
