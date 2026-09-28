package web

import (
	"errors"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/forkext/fleetupdate"
	"github.com/mhsanaei/3x-ui/v3/internal/web/runtime"
)

func TestCorrelatePanelUpdateStatusRejectsStaleRun(t *testing.T) {
	_, err := correlatePanelUpdateStatus("run-new", runtime.PanelUpdateStatus{RunID: "run-old", State: "success"})
	if !errors.Is(err, fleetupdate.ErrStaleStatus) {
		t.Fatalf("error = %v, want stale status", err)
	}
}

func TestCorrelatePanelUpdateStatusPreservesRollbackEvidence(t *testing.T) {
	got, err := correlatePanelUpdateStatus("run-1", runtime.PanelUpdateStatus{
		RunID: "run-1", State: "failed", ExitCode: 7, RolledBack: true, RollbackHealthy: true,
	})
	if err != nil {
		t.Fatalf("correlate: %v", err)
	}
	if got.UpdateEvidence || !got.RolledBack || !got.RollbackHealthy || got.ExitCode != 7 {
		t.Fatalf("convergence = %+v, want failed update with healthy rollback evidence", got)
	}
}

func TestCorrelatePanelUpdateStatusRejectsNonzeroSuccessEvidence(t *testing.T) {
	got, err := correlatePanelUpdateStatus("run-1", runtime.PanelUpdateStatus{
		RunID: "run-1", State: "success", ExitCode: 7,
	})
	if err != nil {
		t.Fatalf("correlate: %v", err)
	}
	if got.UpdateEvidence {
		t.Fatalf("convergence = %+v, want no success evidence for nonzero exit", got)
	}
}
