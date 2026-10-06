package forkext

import "sync"

// RuntimeState is the small, stable vocabulary shared by fork settings and
// feature-owned API/UI surfaces. It describes the relationship between the
// persisted setting and the in-process runtime; it is not an upstream status
// replacement.
type RuntimeState string

const (
	RuntimeStateFeatureOff      RuntimeState = "feature_off"
	RuntimeStateActive          RuntimeState = "active"
	RuntimeStateInitializing    RuntimeState = "initializing"
	RuntimeStateRestartRequired RuntimeState = "restart_required"
	RuntimeStateError           RuntimeState = "error"
)

type runtimeSnapshot struct {
	initialized    bool
	config         runtimeConfig
	preparedConfig runtimeConfig
	prepared       bool
	err            string
}

var runtimeSnapshotState struct {
	sync.RWMutex
	snapshot runtimeSnapshot
}

func publishRuntimeSnapshot(config runtimeConfig, initialized bool, runtimeErr string) {
	runtimeSnapshotState.Lock()
	runtimeSnapshotState.snapshot = runtimeSnapshot{
		initialized:    initialized,
		config:         config,
		preparedConfig: config,
		prepared:       initialized,
		err:            runtimeErr,
	}
	runtimeSnapshotState.Unlock()
}

func publishPreparedRuntime(config runtimeConfig) {
	runtimeSnapshotState.Lock()
	snapshot := runtimeSnapshotState.snapshot
	snapshot.preparedConfig = config
	snapshot.prepared = true
	snapshot.err = ""
	runtimeSnapshotState.snapshot = snapshot
	runtimeSnapshotState.Unlock()
}

func publishRuntimeError(runtimeErr string) {
	runtimeSnapshotState.Lock()
	snapshot := runtimeSnapshotState.snapshot
	snapshot.prepared = false
	snapshot.err = runtimeErr
	runtimeSnapshotState.snapshot = snapshot
	runtimeSnapshotState.Unlock()
}

func currentRuntimeSnapshot() runtimeSnapshot {
	runtimeSnapshotState.RLock()
	snapshot := runtimeSnapshotState.snapshot
	runtimeSnapshotState.RUnlock()
	return snapshot
}

func runtimePreparedFor(config runtimeConfig) bool {
	snapshot := currentRuntimeSnapshot()
	return snapshot.prepared && snapshot.preparedConfig == config
}

func runtimeFlagEnabled(config runtimeConfig, flag Flag) bool {
	switch flag {
	case FlagAnalytics:
		return config.analytics
	case FlagDNSIntelligence:
		return config.dns
	case FlagPolicies:
		return config.policies
	case FlagTrafficControl:
		return config.trafficControl
	case FlagSecurityAnomaly:
		return config.security
	case FlagAudit:
		return config.audit
	case FlagSelfService:
		return config.selfService
	case FlagFleetUpdates:
		return config.fleetUpdates
	case FlagFleetMutation:
		return config.fleetMutation
	case FlagSponsors:
		return config.sponsors
	default:
		return false
	}
}

func featureRuntimeStatus(flag Flag, configured bool) (active bool, state RuntimeState) {
	snapshot := currentRuntimeSnapshot()
	active = runtimeFlagEnabled(snapshot.config, flag)
	if snapshot.err != "" && configured {
		return active, RuntimeStateError
	}
	if !snapshot.initialized {
		if configured {
			return active, RuntimeStateInitializing
		}
		return active, RuntimeStateFeatureOff
	}
	if active != configured {
		return active, RuntimeStateRestartRequired
	}
	if active {
		return active, RuntimeStateActive
	}
	return active, RuntimeStateFeatureOff
}
