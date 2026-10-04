package forkext

import "testing"

func TestFeatureRuntimeStatusDistinguishesSavedAndActiveState(t *testing.T) {
	publishRuntimeSnapshot(runtimeConfig{}, true, "")
	t.Cleanup(func() { publishRuntimeSnapshot(runtimeConfig{}, false, "") })
	if active, state := featureRuntimeStatus(FlagAnalytics, false); active || state != RuntimeStateFeatureOff {
		t.Fatalf("disabled status = active:%v state:%q", active, state)
	}
	if active, state := featureRuntimeStatus(FlagAnalytics, true); active || state != RuntimeStateRestartRequired {
		t.Fatalf("pending enable status = active:%v state:%q", active, state)
	}
	publishRuntimeSnapshot(runtimeConfig{analytics: true}, true, "")
	if active, state := featureRuntimeStatus(FlagAnalytics, true); !active || state != RuntimeStateActive {
		t.Fatalf("active status = active:%v state:%q", active, state)
	}
	publishRuntimeSnapshot(runtimeConfig{}, true, "runtime failed")
	if active, state := featureRuntimeStatus(FlagAnalytics, true); active || state != RuntimeStateError {
		t.Fatalf("error status = active:%v state:%q", active, state)
	}
}
