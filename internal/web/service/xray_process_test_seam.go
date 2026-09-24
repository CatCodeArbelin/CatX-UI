package service

import (
	"context"

	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// SetXrayProcessForTest installs p as the running process and returns the restore func,
// so tests in other packages can observe online state. Never call it in production.
func SetXrayProcessForTest(p *xray.Process) (restore func()) {
	previousProcess, previousResult := xrayState.snapshot()
	xrayState.replace(p)
	return func() {
		xrayState.mu.Lock()
		xrayState.process = previousProcess
		xrayState.result = previousResult
		xrayState.mu.Unlock()
	}
}

// SetXrayRecoveryHooksForTest replaces candidate validation and healthchecking
// and returns a restore function. It exists for cross-package HTTP tests that
// exercise DB import without launching a real core.
func SetXrayRecoveryHooksForTest(
	validate func(context.Context, *xray.Config) error,
	healthcheck func(context.Context, *xray.Process) error,
) (restore func()) {
	previousValidate := validateXrayCandidate
	previousHealthcheck := healthcheckXray
	if validate != nil {
		validateXrayCandidate = validate
	}
	if healthcheck != nil {
		healthcheckXray = healthcheck
	}
	return func() {
		validateXrayCandidate = previousValidate
		healthcheckXray = previousHealthcheck
	}
}

// SetXrayStartHookForTest replaces Process.Start for deterministic startup and
// rollback failure tests.
func SetXrayStartHookForTest(start func(*xray.Process) error) (restore func()) {
	previous := startXrayProcess
	if start != nil {
		startXrayProcess = start
	}
	return func() { startXrayProcess = previous }
}
