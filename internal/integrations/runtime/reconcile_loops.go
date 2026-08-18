package runtime

import (
	"context"
	"errors"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
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

	ctx, headers := intobvs.EmitContext(ctx, oc)

	if _, err := r.Gala().EmitWithHeaders(ctx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, headers); err != nil {
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

		// the keys must match the emitted GetPropertiesForOperationContext projection:
		// installation-bound contexts promote the integration as the operation's entity
		// (entityId), and the reconcile run type keeps the match disjoint from one-shot
		// event jobs sharing the same installation and operation
		fragment, err := types.PropertiesFragment(map[string]string{
			"entityId":  installation.ID,
			"operation": op.Name,
			"runType":   enums.IntegrationRunTypeReconcile.String(),
		})
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
			purged, err := r.Gala().PurgeActiveJobsWithMetadata(opCtx, fragment)
			if err != nil {
				logx.FromContext(opCtx).Error().Err(err).Msg("failed purging duplicate reconcile jobs")
				errs = append(errs, err)

				continue
			}

			logx.FromContext(opCtx).Info().Int("purged", purged).Msg("purged duplicate reconcile jobs")
		}

		if err := r.emitReconcileLoop(opCtx, installation, op.Name); err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed emitting reconcile loop after reset")
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
