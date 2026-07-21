package runtime

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// handleScheduledCycle executes one runtime-bound scheduled operation cycle inline
func (r *Runtime) handleScheduledCycle(ctx context.Context, envelope operations.ReconcileEnvelope) (int, error) {
	operation, err := r.Registry().Operation(envelope.DefinitionID, envelope.Operation)
	if err != nil {
		return 0, err
	}

	if operation.DisabledForAll || (operation.Disabled != nil && operation.Disabled(nil)) {
		return 0, operations.ErrOperationDisabled
	}

	logx.FromContext(ctx).Info().Msg("scheduled operation cycle started")

	response, err := r.executeOperationInline(ctx, nil, envelope.DefinitionID, operation, nil, nil)
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
				logx.FromContext(ctx).Info().Str("definition_id", def.ID).Str("operation", op.Name).Msg("scheduled operation disabled, skipping seed")

				continue
			}

			metadata := types.ExecutionMetadata{
				DefinitionID: def.ID,
				Operation:    op.Name,
				RunType:      enums.IntegrationRunTypeScheduled,
				Runtime:      true,
			}

			if err := r.seedScheduledOperation(ctx, metadata); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// seedScheduledOperation emits one scheduled operation cycle envelope when no active job exists for it
func (r *Runtime) seedScheduledOperation(ctx context.Context, metadata types.ExecutionMetadata) error {
	fragment, err := scheduledMetadataFragment(metadata)
	if err != nil {
		return err
	}

	active, err := r.Gala().HasActiveJobWithMetadata(ctx, fragment)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("definition_id", metadata.DefinitionID).Str("operation", metadata.Operation).Msg("failed to check for active scheduled operation job")

		return err
	}

	if active {
		logx.FromContext(ctx).Debug().Str("definition_id", metadata.DefinitionID).Str("operation", metadata.Operation).Msg("scheduled operation already active, skipping seed")

		return nil
	}

	logx.FromContext(ctx).Info().Str("definition_id", metadata.DefinitionID).Str("operation", metadata.Operation).Msg("seeding scheduled operation")

	receipt := r.Gala().EmitWithHeaders(
		types.WithExecutionMetadata(ctx, metadata),
		operations.ReconcileTopic,
		operations.ReconcileEnvelope{ExecutionMetadata: metadata},
		gala.Headers{
			Properties: metadata.Properties(),
			Tags:       types.GetTagsForExecutionMetadata(metadata),
		},
	)

	return receipt.Err
}

// scheduledMetadataFragment builds the JSONB fragment for active-job checks; the run type
// keeps it disjoint from installation-bound reconcile cycles sharing the topic
func scheduledMetadataFragment(metadata types.ExecutionMetadata) (string, error) {
	props := metadata.Properties()

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
