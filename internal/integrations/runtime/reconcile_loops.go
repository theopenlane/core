package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// emitReconcileLoop starts one operation's loop unless a live one exists; the metadata guard is
// what stops seeds from spawning parallel chains, since successor cycles change their unique key
func (r *Runtime) emitReconcileLoop(ctx context.Context, installation *ent.Integration, operationName string) error {
	oc := types.NewOperationContext(installation.OwnerID, operationName, types.IntegrationSource{
		IntegrationID: installation.ID,
		DefinitionID:  installation.DefinitionID,
		RunType:       enums.IntegrationRunTypeReconcile,
	})

	ctx, headers := intobvs.EmitContext(ctx, oc)

	fragment, err := types.PropertiesFragment(map[string]string{
		"entityId":  installation.ID,
		"operation": operationName,
		"runType":   enums.IntegrationRunTypeReconcile.String(),
	})
	if err != nil {
		return err
	}

	active, err := r.Gala().HasActiveJobWithMetadata(ctx, fragment)
	if err != nil {
		return err
	}

	if active {
		return nil
	}

	op, err := r.Registry().Operation(installation.DefinitionID, operationName)
	if err != nil {
		return err
	}

	if op.ClientRef.Valid() {
		if _, err := r.BuildClientForIntegration(ctx, installation, op.ClientRef); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("client unresolved, marking unhealthy instead of seeding loop")

			return r.MarkIntegrationUnhealthy(ctx, installation, fmt.Sprintf(clientUnresolvedReasonFmt, err))
		}
	}

	if _, err := r.Gala().EmitWithHeaders(ctx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, headers); err != nil {
		return err
	}

	logx.FromContext(ctx).Info().Msg("reconcile loop emitted")

	return nil
}

// clientUnresolvedReasonFmt formats the actionable reason recorded when an integration cannot establish its client
const clientUnresolvedReasonFmt = "the integration could not establish a connection and needs to be reconnected: %s"

// reconcileExhaustedReasonFmt formats the user-facing reason recorded on the unhealthy installation
const reconcileExhaustedReasonFmt = "repeated sync failures due to %s"

// markReconcileExhausted marks the installation unhealthy when its loop exhausts its error budget
func (r *Runtime) markReconcileExhausted(ctx context.Context, e operations.ReconcileEnvelope, cause error) {
	src := types.IntegrationSourceFrom(e.OperationContext)
	if src.IntegrationID == "" {
		return
	}

	ctx = intobvs.WithContext(ctx, e.OperationContext)

	installation, err := r.ResolveIntegration(ctx, IntegrationLookup{IntegrationID: src.IntegrationID})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed resolving integration after exhausted reconcile loop")

		return
	}

	logx.FromContext(ctx).Error().Err(cause).Msg("reconcile loop exhausted error budget, marking integration unhealthy")

	if err := r.MarkIntegrationUnhealthy(ctx, installation, fmt.Sprintf(reconcileExhaustedReasonFmt, cause)); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed marking integration unhealthy after exhausted reconcile loop")
	}
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
