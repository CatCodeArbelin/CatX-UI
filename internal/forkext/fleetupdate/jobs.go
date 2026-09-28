package fleetupdate

import (
	"context"

	"github.com/robfig/cron/v3"
)

type reconcileJob struct{}

func (reconcileJob) Run() {
	s := Current()
	if s == nil || !s.Enabled() {
		return
	}
	campaigns, err := s.List(context.Background())
	if err != nil {
		return
	}
	for _, c := range campaigns {
		if c.State == StateReady || c.State == StatePreflight || c.State == StateDispatching || c.State == StateHealthCheck || c.State == StateSoaking || c.State == StateUnknown {
			_ = s.Reconcile(context.Background(), c.ID, "scheduler")
		}
	}
}

func RegisterJobs(scheduler *cron.Cron) {
	if scheduler != nil {
		s := Current()
		if s == nil || !s.Enabled() {
			return
		}
		_, _ = scheduler.AddJob("@every 10s", reconcileJob{})
	}
}
