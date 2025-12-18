package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/jobtemplate"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

// HookScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// a cron and the configuration matches what is expected
func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ScheduledJobMutation,
		) (generated.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			cron, _ := mutation.Cron()
			jobID, _ := mutation.JobID()

			// if the cron is not set on create, attempt to inherit from the job template
			// let the schema do the validation
			if jobID != "" {
				job, err := mutation.Client().JobTemplate.Query().
					Where(jobtemplate.ID(jobID)).
					Select(jobtemplate.FieldCron).
					Only(ctx)
				if err != nil {
					return nil, err
				}

				if cron == "" && job.Cron != nil {
					cron = *job.Cron
					mutation.SetCron(cron)
				}
			}

			// validate the job runner is in the organization
			jobRunnerID, hasJobRunnerID := mutation.JobRunnerID()
			if hasJobRunnerID && jobRunnerID != "" {
				exists, err := mutation.Client().JobRunner.Query().
					Where(jobrunner.ID(jobRunnerID)).
					Exist(ctx)
				if err != nil {
					return nil, err
				}

				if !exists {
					logx.FromContext(ctx).Debug().Str("job_runner_id", jobRunnerID).Msg("requested job runner not found")

					return nil, &generated.NotFoundError{}
				}
			}

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return retVal, err
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}
