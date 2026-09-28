package web

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/fleetupdate"
	"github.com/mhsanaei/3x-ui/v3/internal/web/runtime"
)

type fleetUpdateRemoteExecutor struct{ manager *runtime.Manager }

func (e fleetUpdateRemoteExecutor) remote(n fleetupdate.NodeSnapshot) (*runtime.Remote, error) {
	if e.manager == nil {
		return nil, errors.New("runtime manager unavailable")
	}
	var node model.Node
	if db := database.GetDB(); db == nil || db.First(&node, n.ID).Error != nil {
		return nil, errors.New("node not found")
	}
	if !node.Enable || n.Transitive || node.Id <= 0 {
		return nil, errors.New("node is not a direct eligible node")
	}
	return e.manager.RemoteFor(&node)
}

func (e fleetUpdateRemoteExecutor) Preflight(_ context.Context, n fleetupdate.NodeSnapshot, _ fleetupdate.ReleaseSnapshot) (fleetupdate.PreflightResult, error) {
	if _, err := e.remote(n); err != nil {
		return fleetupdate.PreflightResult{}, err
	}
	return fleetupdate.PreflightResult{Supported: true}, nil
}

func (e fleetUpdateRemoteExecutor) Dispatch(ctx context.Context, n fleetupdate.NodeSnapshot, rel fleetupdate.ReleaseSnapshot) (fleetupdate.DispatchResult, error) {
	remote, err := e.remote(n)
	if err != nil {
		return fleetupdate.DispatchResult{}, err
	}
	start, err := remote.StartUpdate(ctx, rel.Channel == "dev")
	if err != nil {
		if ambiguousRemoteError(err) {
			return fleetupdate.DispatchResult{Evidence: "post-outcome-ambiguous", Ambiguous: true}, fleetupdate.ErrAmbiguousDispatch
		}
		return fleetupdate.DispatchResult{}, err
	}
	return fleetupdate.DispatchResult{RunID: start.RunID, Evidence: "node-accepted-run:" + start.RunID}, nil
}

func (e fleetUpdateRemoteExecutor) Reconcile(ctx context.Context, n fleetupdate.NodeSnapshot, _ fleetupdate.ReleaseSnapshot) (fleetupdate.ConvergenceResult, error) {
	remote, err := e.remote(n)
	if err != nil {
		return fleetupdate.ConvergenceResult{}, err
	}
	status, err := remote.GetUpdateStatus(ctx)
	if err != nil {
		return fleetupdate.ConvergenceResult{}, err
	}
	conv, err := correlatePanelUpdateStatus(n.ExpectedRunID, status)
	if err != nil {
		return fleetupdate.ConvergenceResult{}, err
	}
	conv.FreshHeartbeat = freshHeartbeat(n.LastHeartbeat)
	conv.PanelVersion = n.PanelVersion
	conv.NodeStatus = n.Status
	conv.XrayState = n.XrayState
	return conv, nil
}

func correlatePanelUpdateStatus(expected string, status runtime.PanelUpdateStatus) (fleetupdate.ConvergenceResult, error) {
	if expected == "" || status.RunID == "" || status.RunID != expected {
		return fleetupdate.ConvergenceResult{}, fleetupdate.ErrStaleStatus
	}
	return fleetupdate.ConvergenceResult{
		RunID:       status.RunID,
		UpdateState: status.State, ExitCode: status.ExitCode, FinishedAt: status.FinishedAt,
		RolledBack: status.RolledBack, RollbackHealthy: status.RollbackHealthy,
		UpdateEvidence: status.State == "success" && status.ExitCode == 0 && !status.RolledBack,
	}, nil
}

func freshHeartbeat(last int64) bool {
	return last > 0 && time.Since(time.Unix(last, 0)) <= 30*time.Second
}

func ambiguousRemoteError(err error) bool {
	var ne net.Error
	return errors.Is(err, context.DeadlineExceeded) || errors.As(err, &ne)
}
