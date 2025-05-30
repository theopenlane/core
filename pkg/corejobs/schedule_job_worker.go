package corejobs

import (
	"context"

	"github.com/riverqueue/river"
)

// ScheduledJobArgs represents the arguments for the scheduled job worker
type ScheduledJobArgs struct{}

func (ScheduledJobArgs) Kind() string { return "scheduled_jobs" }

// ScheduledJobConfig contains the configuration for the scheduling job worker
type ScheduledJobConfig struct {
}

// ScheduledJobWorker is a queue worker that schedules job that can be executed by agents
type ScheduledJobWorker struct {
	river.WorkerDefaults[ScheduledJobArgs]

	Config ScheduledJobConfig `koanf:"config" json:"config" jsonschema:"description=the scheduled job worker configuration"`
}

// Work evaluates the available jobs and marks them as ready to be executed by agents if needed
func (s *ScheduledJobWorker) Work(ctx context.Context, _ *river.Job[ScheduledJobArgs]) error {
	// prevent the "runs" table from being bloated
	// any item that should have been scheduled should be removed.
	// The results would be in the "job results". but the logs not needed here
	return nil
}
