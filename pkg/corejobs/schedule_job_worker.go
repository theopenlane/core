package corejobs

import (
	"bytes"
	"context"
	"errors"
	"text/template"
	"time"

	"github.com/riverqueue/river"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/controlscheduledjob"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

var (
	// ErrNoEntClient is an error to identify a scheduled worker without an ent client
	ErrNoEntClient = errors.New("please configure the ent/db client")
)

// ScheduledJobArgs represents the arguments for the scheduled job worker
type ScheduledJobArgs struct {
	JobID string
}

func (ScheduledJobArgs) Kind() string { return "scheduled_jobs" }

// ScheduledJobConfig contains the configuration for the scheduling job worker
type ScheduledJobConfig struct {
}

// ScheduledJobWorker is a queue worker that schedules job that can be executed by agents
type ScheduledJobWorker struct {
	river.WorkerDefaults[ScheduledJobArgs]

	Config ScheduledJobConfig `koanf:"config" json:"config" jsonschema:"description=the scheduled job worker configuration"`

	client *generated.Client
}

// Work evaluates the available jobs and marks them as ready to be executed by agents if needed
func (s *ScheduledJobWorker) Work(ctx context.Context, job *river.Job[ScheduledJobArgs]) error {
	// prevent the "runs" table from being bloated
	// any item that should have been scheduled should be removed.
	// The results would be in the "job results". but the logs not needed here
	if s.client == nil {
		return ErrNoEntClient
	}

	if job.Args.JobID == "" {
		return newMissingRequiredArg("job_id", ScheduledJobArgs{}.Kind())
	}

	scheduledJob, err := s.client.ControlScheduledJob.Query().
		Where(controlscheduledjob.ID(job.Args.JobID)).
		Only(ctx)
	if err != nil {
		return err
	}

	jobDefinition, err := scheduledJob.QueryJob().Only(ctx)
	if err != nil {
		return err
	}

	script, err := parseConfigIntoScript(jobDefinition.JobType, scheduledJob.Configuration, jobDefinition.Script)
	if err != nil {
		return err
	}

	return s.client.ScheduledJobRun.Create().
		SetScript(script).
		SetExpectedExecutionTime(job.ScheduledAt).
		SetCreatedAt(time.Now()).
		SetJobRunnerID(scheduledJob.JobRunnerID).
		SetStatus(enums.ScheduledJobRunStatusPending).
		Exec(ctx)
}

func parseConfigIntoScript(jobType enums.JobType, cfg models.JobConfiguration, script string) (string, error) {
	tmpl, err := template.New("script").Parse(script)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	switch jobType {
	case enums.JobTypeSsl:
		if err := tmpl.Execute(&buf, cfg.SSL); err != nil {
			return "", err
		}

		return buf.String(), nil

	default:
		// return as-is for now.
		// TODO: when we support "others" job type, make sure to parse the entire cfg object here
		return script, nil
	}
}

// WithEntClient configures the worker with the configured db client
func (s *ScheduledJobWorker) WithEntClient(client *generated.Client) { s.client = client }
