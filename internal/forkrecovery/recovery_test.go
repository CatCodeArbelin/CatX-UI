package forkrecovery

import (
	"errors"
	"strings"
	"testing"
)

func TestFailureReportsSuccessfulRollback(t *testing.T) {
	candidate := errors.New("candidate unhealthy")
	err := NewFailure("Xray activation", candidate, nil)
	var failure *Failure
	if !errors.As(err, &failure) {
		t.Fatalf("error type = %T, want *Failure", err)
	}
	if failure.RollbackFailed() {
		t.Fatal("successful rollback reported as failed")
	}
	if !errors.Is(err, candidate) {
		t.Fatal("candidate error is not preserved for errors.Is")
	}
	if !strings.Contains(err.Error(), "automatic rollback succeeded") {
		t.Fatalf("error = %q, want successful rollback report", err)
	}
}

func TestFailureDistinguishesFailedRollback(t *testing.T) {
	err := NewFailure("database import", errors.New("candidate invalid"), errors.New("restore failed"))
	var failure *Failure
	if !errors.As(err, &failure) || !failure.RollbackFailed() {
		t.Fatalf("error = %v, want distinguishable failed rollback", err)
	}
	if !strings.Contains(err.Error(), "automatic rollback failed: restore failed") {
		t.Fatalf("error = %q, want rollback failure detail", err)
	}
}
