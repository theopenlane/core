package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// HookScheduledJobCreate verifies a scheduled job has
// a cadence set or a cron and the configuration matches what is expected
func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ScheduledJobMutation) (generated.Value, error) {
			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			cfg, ok := mutation.Configuration()
			if !ok {
				return nil, errors.New("configuration required") // nolint:err113
			}

			jobType, _ := mutation.JobType()

			if err := cfg.Validate(jobType); err != nil {
				return nil, err
			}

			cadence, hasCadence := mutation.Cadence()
			cron, hasCron := mutation.Cron()

			if err := validateCadenceOrCron(&cadence, hasCadence, cron, hasCron); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

// HookControlScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// a cadence set or a cron and the configuration matches what is expected
func HookControlScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ControlScheduledJobMutation) (generated.Value, error) {
			cadence, hasCadence := mutation.Cadence()
			cron, hasCron := mutation.Cron()

			if entx.CheckIsSoftDelete(ctx) {
				handle, ok := mutation.JobHandle()
				if ok {
					mutation.Job.GetRiverClient().PeriodicJobs().Remove(rivertype.PeriodicJobHandle(handle))
				}

				return next.Mutate(ctx, mutation)
			}

			if err := validateCadenceOrCron(&cadence, hasCadence, cron, hasCron); err != nil {
				return nil, err
			}

			v, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			job := v.(*generated.ControlScheduledJob)

			if mutation.Op() == ent.OpCreate && (hasCadence || hasCron) {
				var scheduler river.PeriodicSchedule

				if hasCadence {
					scheduler = cadence
				} else {
					scheduler = cron
				}

				handle := mutation.Job.GetRiverClient().PeriodicJobs().Add(
					river.NewPeriodicJob(scheduler, func() (river.JobArgs, *river.InsertOpts) {
						return corejobs.ScheduledJobArgs{
							JobID: job.ID,
						}, nil
					}, nil))

				err := mutation.Client().ControlScheduledJob.
					UpdateOne(job).
					SetJobHandle(int(handle)).
					Exec(ctx)
				if err != nil {
					return nil, err
				}
			}

			return job, err
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

func validateCadenceOrCron(cadence *models.JobCadence, hasCadence bool, cron models.Cron, hasCron bool) error {
	if !hasCadence && (!hasCron || cron == "") {
		return nil
	}

	if hasCadence && hasCron {
		return ErrEitherCadenceOrCron
	}

	if hasCadence {
		if err := cadence.Validate(); err != nil {
			return fmt.Errorf("cadence: %w", err) // nolint:err113
		}
	}

	if hasCron {
		return cron.Validate()
	}

	return nil
}
