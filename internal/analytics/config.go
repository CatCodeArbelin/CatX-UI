package analytics

import (
	"context"
	"time"
)

type Status struct {
	Enabled         bool            `json:"enabled"`
	DNSIntelligence bool            `json:"dnsIntelligence"`
	Retention       RetentionPolicy `json:"retention"`
}

// SetRetentionPolicy changes only the in-process analytics policy. Persistence
// is owned by the fork settings facade; this keeps analytics independent from
// the upstream settings service.
func SetRetentionPolicy(policy RetentionPolicy) {
	configured.Lock()
	defer configured.Unlock()
	configured.retention = policy
}

func CurrentStatus() Status {
	configured.RLock()
	defer configured.RUnlock()
	return Status{Enabled: configured.enabled, DNSIntelligence: configured.enabled && configured.evidenceEnabled, Retention: configured.retention}
}

func pruneLoop(ctx context.Context, repo Repository) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			configured.RLock()
			policy := configured.retention
			configured.RUnlock()
			if _, err := repo.Prune(ctx, now, policy); err != nil {
				// Pruning is best-effort and isolated from Xray/panel startup.
				continue
			}
		}
	}
}
