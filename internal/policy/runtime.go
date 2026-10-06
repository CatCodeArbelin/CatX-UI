package policy

import "sync"

const (
	RuntimeStateActive          = "active"
	RuntimeStateRestartRequired = "restart_required"
	RuntimeStateError           = "error"
)

var runtimeApply struct {
	sync.RWMutex
	pending  bool
	lastErr  string
	onChange func()
}

// SetRuntimeChangeCallback connects policy persistence to the existing Xray
// restart boundary. The policy package only requests the boundary; it never
// owns a second Xray controller.
func SetRuntimeChangeCallback(fn func()) {
	runtimeApply.Lock()
	runtimeApply.onChange = fn
	runtimeApply.Unlock()
}

// MarkRuntimeApplyRequired records that durable policy state is newer than
// the effective Xray configuration and asks the upstream owner to apply it.
func MarkRuntimeApplyRequired() {
	runtimeApply.Lock()
	runtimeApply.pending = true
	runtimeApply.lastErr = ""
	callback := runtimeApply.onChange
	runtimeApply.Unlock()
	if callback != nil {
		callback()
	}
}

// MarkRuntimeApplied is called only after the existing Xray restart/apply
// path has validated and health-checked its candidate configuration.
func MarkRuntimeApplied() {
	runtimeApply.Lock()
	runtimeApply.pending = false
	runtimeApply.lastErr = ""
	runtimeApply.Unlock()
}

func MarkRuntimeApplyError(err error) {
	runtimeApply.Lock()
	runtimeApply.pending = true
	if err == nil {
		runtimeApply.lastErr = ""
	} else {
		runtimeApply.lastErr = err.Error()
	}
	runtimeApply.Unlock()
}

type RuntimeStatusView struct {
	State           string `json:"state"`
	RestartRequired bool   `json:"restartRequired"`
	LastError       string `json:"lastError,omitempty"`
}

func RuntimeStatus() RuntimeStatusView {
	runtimeApply.RLock()
	pending, lastErr := runtimeApply.pending, runtimeApply.lastErr
	runtimeApply.RUnlock()
	if lastErr != "" {
		return RuntimeStatusView{State: RuntimeStateError, RestartRequired: true, LastError: lastErr}
	}
	if pending {
		return RuntimeStatusView{State: RuntimeStateRestartRequired, RestartRequired: true}
	}
	return RuntimeStatusView{State: RuntimeStateActive}
}
