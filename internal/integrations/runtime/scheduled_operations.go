package runtime

import (
	"context"
	"errors"

	"github.com/theopenlane/core/common/enums"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/logx"
)

// handleScheduledCycle executes one runtime-bound scheduled operation cycle inline
func (r *Runtime) handleScheduledCycle(ctx context.Context, envelope operations.ReconcileEnvelope) (int, error) {
	src := types.IntegrationSourceFrom(envelope.OperationContext)

	operation, err := r.Registry().Operation(src.DefinitionID, envelope.Operation)
	if err != nil {
		return 0, err
	}

	if operation.DisabledForAll || (operation.Disabled != nil && operation.Disabled(nil)) {
		return 0, operations.ErrOperationDisabled
	}

	logx.FromContext(ctx).Info().Msg("scheduled operation cycle started")

	response, err := r.executeOperationInline(ctx, nil, src.DefinitionID, operation, nil, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("scheduled operation cycle failed")

		return 0, err
	}

	var result types.ScheduledCycleResult
	if err := jsonx.UnmarshalIfPresent(response, &result); err != nil {
		return 0, err
	}

	logx.FromContext(ctx).Info().Int("processed", result.Processed).Msg("scheduled operation cycle completed")

	return result.Processed, nil
}

// SeedScheduledOperations ensures every operation with the Scheduled policy has an active
// polling loop, called once at startup
func (r *Runtime) SeedScheduledOperations(ctx context.Context) error {
	var errs []error

	for _, def := range r.Registry().Definitions() {
		if !def.Active {
			continue
		}

		for _, op := range def.Operations {
			if !op.Policy.Scheduled {
				continue
			}

			oc := types.NewOperationContext("", op.Name, types.IntegrationSource{
				DefinitionID: def.ID,
				RunType:      enums.IntegrationRunTypeScheduled,
				Runtime:      true,
			})

			if op.DisabledForAll || (op.Disabled != nil && op.Disabled(nil)) {
				logx.FromContext(intobvs.WithContext(ctx, oc)).Info().Msg("scheduled operation disabled, skipping seed")

				continue
			}

			if err := r.seedScheduledOperation(ctx, oc); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// seedScheduledOperation emits one scheduled operation cycle envelope unless the loop is
// already live; successor cycles carry per-cycle unique keys the seed's key can't collide with
func (r *Runtime) seedScheduledOperation(ctx context.Context, oc gala.OperationContext) error {
	ctx, headers := intobvs.EmitContext(ctx, oc)

	props := types.GetPropertiesForOperationContext(oc)

	fragment, err := types.PropertiesFragment(map[string]string{
		"definitionId": props["definitionId"],
		"operation":    props["operation"],
		"runType":      props["runType"],
	})
	if err != nil {
		return err
	}

	active, err := r.Gala().HasActiveJobWithMetadata(ctx, fragment)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to check for active scheduled operation job")

		return err
	}

	if active {
		return nil
	}

	logx.FromContext(ctx).Info().Msg("seeding scheduled operation")

	_, err = r.Gala().EmitWithHeaders(ctx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, headers)

	return err
}
