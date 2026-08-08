package runtime

import (
	"context"
	"errors"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// emitReconcileLoop emits the recurring loop for one operation on an installation; the topic's
// UniqueKey derivation collapses concurrent seeds of the same loop to one job
func (r *Runtime) emitReconcileLoop(ctx context.Context, installation *ent.Integration, operationName string) error {
	oc := types.NewOperationContext(installation.OwnerID, operationName, types.IntegrationSource{
		IntegrationID: installation.ID,
		DefinitionID:  installation.DefinitionID,
		RunType:       enums.IntegrationRunTypeReconcile,
	})

	ctx = intobvs.WithContext(ctx, oc)

	if _, err := r.Gala().Emit(ctx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, gala.WithHeaders(gala.Headers{
		Properties: types.GetPropertiesForOperationContext(oc),
	})); err != nil {
		return err
	}

	logx.FromContext(ctx).Info().Msg("reconcile loop emitted")

	return nil
}

// ResetReconcileLoops collapses each reconcilable operation on the installation to exactly one
// recurring loop: an operation already running a single loop is left untouched (preserving its
// adaptive schedule state), while zero or multiple loops are cancelled and reseeded as one fresh
// unique loop
func (r *Runtime) ResetReconcileLoops(ctx context.Context, installation *ent.Integration) error {
	if installation.Status != enums.IntegrationStatusConnected {
		return nil
	}

	ctx = intobvs.WithInstallation(ctx, installation)

	active, err := r.isOrgSubscriptionActive(ctx, installation.OwnerID)
	if err != nil {
		return err
	}

	if !active {
		logx.FromContext(ctx).Info().Msg("owner subscription is not active, skipping reconcile loop reset")

		return nil
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
		return nil
	}

	var errs []error

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		if op.Disabled != nil && op.Disabled(installation.Config.ClientConfig) {
			continue
		}

		opCtx := intobvs.WithOperation(ctx, op.Name)

		fragment, err := reconcileMetadataFragment(installation.ID, op.Name)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		count, err := r.Gala().CountActiveJobsWithMetadata(opCtx, fragment)
		if err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed counting reconcile jobs for loop reset")
			errs = append(errs, err)

			continue
		}

		if count == 1 {
			continue
		}

		if count > 1 {
			cancelled, err := r.Gala().CancelActiveJobsWithMetadata(opCtx, fragment)
			if err != nil {
				logx.FromContext(opCtx).Error().Err(err).Msg("failed cancelling duplicate reconcile jobs")
				errs = append(errs, err)

				continue
			}

			logx.FromContext(opCtx).Info().Int("cancelled", cancelled).Msg("cancelled duplicate reconcile jobs")
		}

		if err := r.emitReconcileLoop(opCtx, installation, op.Name); err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed emitting reconcile loop after reset")
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
