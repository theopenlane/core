package runtime

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
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

			if op.DisabledForAll || (op.Disabled != nil && op.Disabled(nil)) {
				logx.FromContext(ctx).Info().Str("definition_id", def.ID).Str(intobvs.FieldOperation, op.Name).Msg("scheduled operation disabled, skipping seed")

				continue
			}

			oc := types.NewOperationContext("", op.Name, types.IntegrationSource{
				DefinitionID: def.ID,
				RunType:      enums.IntegrationRunTypeScheduled,
				Runtime:      true,
			})

			if err := r.seedScheduledOperation(ctx, oc); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// seedScheduledOperation emits one scheduled operation cycle envelope when no active job exists for it
func (r *Runtime) seedScheduledOperation(ctx context.Context, oc gala.OperationContext) error {
	src := types.IntegrationSourceFrom(oc)

	fragment, err := scheduledMetadataFragment(oc)
	if err != nil {
		return err
	}

	active, err := r.Gala().HasActiveJobWithMetadata(ctx, fragment)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("definition_id", src.DefinitionID).Str(intobvs.FieldOperation, oc.Operation).Msg("failed to check for active scheduled operation job")

		return err
	}

	if active {
		logx.FromContext(ctx).Debug().Str("definition_id", src.DefinitionID).Str(intobvs.FieldOperation, oc.Operation).Msg("scheduled operation already active, skipping seed")

		return nil
	}

	logx.FromContext(ctx).Info().Str("definition_id", src.DefinitionID).Str(intobvs.FieldOperation, oc.Operation).Msg("seeding scheduled operation")

	receipt := r.Gala().EmitWithHeaders(
		gala.WithOperationContext(ctx, oc),
		operations.ReconcileTopic,
		operations.ReconcileEnvelope{OperationContext: oc},
		gala.Headers{
			Properties: types.GetPropertiesForOperationContext(oc),
			Tags:       types.GetTagsForOperationContext(oc),
		},
	)

	return receipt.Err
}

// scheduledMetadataFragment builds the JSONB fragment for active-job checks; the run type
// keeps it disjoint from installation-bound reconcile cycles sharing the topic
func scheduledMetadataFragment(oc gala.OperationContext) (string, error) {
	props := types.GetPropertiesForOperationContext(oc)

	b, err := json.Marshal(map[string]map[string]string{
		"properties": {
			"definitionId": props["definitionId"],
			"operation":    props["operation"],
			"runType":      props["runType"],
		},
	})
	if err != nil {
		return "", err
	}

	return string(b), nil
}
