package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// HookScheduledJobCreate verifies a scheduled job has
// the a cadence set or a cron and the configuration matches what is expected
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

// HookScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// the a cadence set or a cron and the configuration matches what is expected
func HookControlScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ControlScheduledJobMutation) (generated.Value, error) {
			cadence, hasCadence := mutation.Cadence()
			cron, hasCron := mutation.Cron()

			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			if err := validateCadenceOrCron(&cadence, hasCadence, cron, hasCron); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

func validateCadenceOrCron(cadence *models.JobCadence, hasCadence bool, cron string, hasCron bool) error {
	if !hasCadence && (!hasCron || cron == "") {
		return errors.New("either cadence or cron must be specified") // nolint:err113
	}

	if hasCadence && hasCron {
		return errors.New("only one of cadence or cron must be specified") // nolint:err113
	}

	if hasCadence {
		if err := cadence.Validate(); err != nil {
			return fmt.Errorf("cadence: %w", err) // nolint:err113
		}
	}

	if hasCron {
		if err := models.ValidateCronExpression(cron); err != nil {
			return err
		}
	}

	return nil
}
