package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// resetReconcileLoop collapses one operation to exactly one recurring loop, purging duplicates
func (r *Runtime) resetReconcileLoop(ctx context.Context, installation *ent.Integration, op types.OperationRegistration) error {
	oc := types.NewOperationContext(installation.OwnerID, op.Name, types.IntegrationSource{
		IntegrationID: installation.ID,
		DefinitionID:  installation.DefinitionID,
		RunType:       enums.IntegrationRunTypeReconcile,
	})

	ctx, headers := intobvs.EmitContext(ctx, oc)

	fragment, err := reconcileLoopFragment(installation.ID, op.Name)
	if err != nil {
		return err
	}

	count, err := r.Gala().CountActiveJobsWithMetadata(ctx, fragment)
	if err != nil {
		return err
	}

	switch {
	case count == 1:
		return nil
	case count > 1:
		purged, err := r.Gala().PurgeActiveJobsWithMetadata(ctx, fragment)
		if err != nil {
			return err
		}

		logx.FromContext(ctx).Info().Int("purged", purged).Msg("purged duplicate reconcile jobs")
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

// reconcileLoopFragment builds the JSONB fragment matching one operation's recurring loop jobs
func reconcileLoopFragment(integrationID, operationName string) (string, error) {
	return types.PropertiesFragment(map[string]string{
		"entityId":  integrationID,
		"operation": operationName,
		"runType":   enums.IntegrationRunTypeReconcile.String(),
	})
}

// clientUnresolvedReasonFmt formats the reason recorded when a client can't be established
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

// ResetReconcileLoops collapses every runnable reconcile operation to exactly one recurring loop
func (r *Runtime) ResetReconcileLoops(ctx context.Context, installation *ent.Integration) error {
	if !lo.Contains(enums.IntegrationOperationalStatuses, installation.Status) {
		return nil
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
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

	var errs []error

	for _, op := range def.Operations {
		_, failing := installation.Health.UnhealthyOperations[op.Name]
		if !op.Policy.Reconcile || failing || op.DisabledFor(installation.Config.ClientConfig) {
			continue
		}

		if err := r.resetReconcileLoop(ctx, installation, op); err != nil {
			errs = append(errs, fmt.Errorf("reset reconcile loop %s: %w", op.Name, err))
		}
	}

	return errors.Join(errs...)
}
