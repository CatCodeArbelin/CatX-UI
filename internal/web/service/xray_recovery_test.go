package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/forkrecovery"
	"github.com/mhsanaei/3x-ui/v3/internal/util/json_util"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func setupXrayRecoveryTest(t *testing.T) (*XrayService, *xray.Config) {
	t.Helper()
	setupSettingTestDB(t)
	t.Setenv("XUI_BIN_FOLDER", t.TempDir())
	t.Setenv("XUI_LOG_FOLDER", t.TempDir())
	preserveXrayTestGlobals(t)
	isManuallyStopped.Store(false)
	isNeedXrayRestart.Store(false)

	svc := &XrayService{}
	candidate, err := svc.GetXrayConfig()
	if err != nil {
		t.Fatalf("GetXrayConfig: %v", err)
	}
	knownGood, err := forkrecovery.CloneXrayConfig(candidate)
	if err != nil {
		t.Fatalf("clone config: %v", err)
	}
	knownGood.LogConfig = json_util.RawMessage(`{"loglevel":"warning","recoveryTest":"known-good"}`)
	xrayState.replace(xray.NewProcess(knownGood))
	knownGoodFile, err := json.Marshal(knownGood)
	if err != nil {
		t.Fatalf("marshal known-good config: %v", err)
	}
	if err := os.WriteFile(xray.GetConfigPath(), knownGoodFile, 0o600); err != nil {
		t.Fatalf("write known-good config: %v", err)
	}
	return svc, knownGood
}

func preserveXrayTestGlobals(t *testing.T) {
	t.Helper()
	previousProcess, previousResult := xrayState.snapshot()
	previousHeldBack := xrayState.heldBackReason()
	previousManual := isManuallyStopped.Load()
	previousNeedRestart := isNeedXrayRestart.Load()
	t.Cleanup(func() {
		xrayState.mu.Lock()
		xrayState.process = previousProcess
		xrayState.result = previousResult
		xrayState.heldBack = previousHeldBack
		xrayState.mu.Unlock()
		isManuallyStopped.Store(previousManual)
		isNeedXrayRestart.Store(previousNeedRestart)
	})
}

func TestRestartXrayValidationFailureNeverActivatesCandidate(t *testing.T) {
	svc, knownGood := setupXrayRecoveryTest(t)
	started := 0
	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error { return errors.New("invalid candidate") },
		func(context.Context, *xray.Process) error { return nil },
	)
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error {
		started++
		return nil
	})
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err := svc.RestartXray(true)
	if err == nil || !strings.Contains(err.Error(), "candidate validation failed") {
		t.Fatalf("RestartXray error = %v, want validation failure", err)
	}
	if started != 0 {
		t.Fatalf("process starts = %d, invalid candidate must never activate", started)
	}
	if got := currentXrayProcess().GetConfig(); !got.Equals(knownGood) {
		t.Fatal("invalid candidate replaced the known-good config")
	}
}

func TestRestartXrayStartupFailureRollsBack(t *testing.T) {
	svc, knownGood := setupXrayRecoveryTest(t)
	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error { return nil },
		func(context.Context, *xray.Process) error { return nil },
	)
	starts := 0
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error {
		starts++
		if starts == 1 {
			return errors.New("exec failed")
		}
		return nil
	})
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err := svc.RestartXray(true)
	var failure *forkrecovery.Failure
	if !errors.As(err, &failure) || failure.RollbackFailed() {
		t.Fatalf("RestartXray error = %v, want candidate failure with successful rollback", err)
	}
	if starts != 2 {
		t.Fatalf("process starts = %d, want candidate then rollback", starts)
	}
	if got := currentXrayProcess().GetConfig(); !got.Equals(knownGood) {
		t.Fatal("startup failure did not restore known-good config")
	}
}

func TestRestartXrayHealthcheckFailureRollsBack(t *testing.T) {
	svc, knownGood := setupXrayRecoveryTest(t)
	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error { return nil },
		func(_ context.Context, process *xray.Process) error {
			if process.GetConfig().Equals(knownGood) {
				return nil
			}
			return errors.New("candidate API unavailable")
		},
	)
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error { return nil })
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err := svc.RestartXray(true)
	var failure *forkrecovery.Failure
	if !errors.As(err, &failure) || failure.RollbackFailed() {
		t.Fatalf("RestartXray error = %v, want health failure with successful rollback", err)
	}
	if got := currentXrayProcess().GetConfig(); !got.Equals(knownGood) {
		t.Fatal("healthcheck failure did not restore known-good config")
	}
}

func TestRestartXrayReportsRollbackFailureSeparately(t *testing.T) {
	svc, knownGood := setupXrayRecoveryTest(t)
	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error { return nil },
		func(_ context.Context, process *xray.Process) error {
			if !process.GetConfig().Equals(knownGood) {
				return errors.New("candidate unhealthy")
			}
			return nil
		},
	)
	starts := 0
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error {
		starts++
		if starts == 1 {
			if err := os.WriteFile(xray.GetConfigPath(), []byte(`{"candidate":"on-disk"}`), 0o600); err != nil {
				t.Fatalf("write simulated candidate config: %v", err)
			}
		}
		if starts == 2 {
			return errors.New("known-good restart failed")
		}
		return nil
	})
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err := svc.RestartXray(true)
	var failure *forkrecovery.Failure
	if !errors.As(err, &failure) || !failure.RollbackFailed() {
		t.Fatalf("RestartXray error = %v, want distinguishable rollback failure", err)
	}
	if !strings.Contains(err.Error(), "known-good restart failed") {
		t.Fatalf("error = %q, want rollback cause", err)
	}
	restoredFile, readErr := os.ReadFile(xray.GetConfigPath())
	if readErr != nil {
		t.Fatalf("read restored config: %v", readErr)
	}
	var restoredConfig xray.Config
	if unmarshalErr := json.Unmarshal(restoredFile, &restoredConfig); unmarshalErr != nil {
		t.Fatalf("restored config is invalid: %v", unmarshalErr)
	}
	if !restoredConfig.Equals(knownGood) {
		t.Fatal("rollback process failure did not restore exact known-good config file")
	}
}
