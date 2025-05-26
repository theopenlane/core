package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/models"
)

func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,

			mutation *generated.ScheduledJobMutation) (generated.Value, error) {
			cfg, ok := mutation.Configuration()
			if !ok {
				return nil, errors.New("configuration required")
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

func HookControlScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ControlScheduledJobMutation) (generated.Value, error) {

			cadence, hasCadence := mutation.Cadence()
			cron, hasCron := mutation.Cron()
			fmt.Println(cadence, cron, hasCadence, hasCron)

			if err := validateCadenceOrCron(&cadence, hasCadence, cron, hasCron); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

func validateCadenceOrCron(cadence *models.JobCadence, hasCadence bool, cron string, hasCron bool) error {

	if !hasCadence && (!hasCron || cron == "") {
		return errors.New("either cadence or cron must be specified")
	}

	if hasCadence && hasCron {
		return errors.New("only one of cadence or cron must be specified")
	}

	if hasCadence {
		if err := cadence.Validate(); err != nil {
			return fmt.Errorf("cadence: %w", err)
		}
	}

	if hasCron {
		if err := models.ValidateCronExpression(cron); err != nil {
			return err
		}
	}

	return nil
}
