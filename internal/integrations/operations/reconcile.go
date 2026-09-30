package operations

import (
	"context"
	"errors"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ReconcileEnvelope is the durable payload for one recurring operation cycle
type ReconcileEnvelope struct {
	gala.OperationContext
	// Schedule is the adaptive scheduling state carried across cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// ReconcileUniqueKey derives the insert-time uniqueness key for one recurring loop
func ReconcileUniqueKey(e ReconcileEnvelope) string {
	src := types.IntegrationSourceFrom(e.OperationContext)

	return gala.IntegrationReconcile.Key(src.IntegrationID, src.DefinitionID, e.Operation)
}

// ReconcileTopic is the durable reconcile topic derived from the envelope type
var ReconcileTopic = gala.NamespacedTopicFor(gala.IntegrationReconcile, gala.WithUniqueKey(ReconcileUniqueKey))

// ReconcileDefinition builds the gala listener definition for recurring operation cycles
func ReconcileDefinition(reg *registry.Registry, handle func(context.Context, ReconcileEnvelope) (int, error), onExhausted func(context.Context, ReconcileEnvelope, error), schedule gala.Schedule) gala.Definition[ReconcileEnvelope] {
	return gala.Definition[ReconcileEnvelope]{
		Topic: ReconcileTopic,
		Cancel: func(ctx context.Context, e ReconcileEnvelope, err error) bool {
			return reconcileShouldCancel(ctx, reg, e, err)
		},
		OnExhausted: onExhausted,
		Schedule: &gala.ScheduleSpec[ReconcileEnvelope]{
			Schedule: schedule,
			Handle:   handle,
			State:    func(e ReconcileEnvelope) gala.ScheduleState { return e.Schedule },
			Wrap: func(e ReconcileEnvelope, s gala.ScheduleState) ReconcileEnvelope {
				return ReconcileEnvelope{
					OperationContext: e.OperationContext,
					Schedule:         s,
				}
			},
			PrepareEmit: func(ctx context.Context, e ReconcileEnvelope) (context.Context, gala.Headers) {
				return intobvs.EmitContext(ctx, e.OperationContext)
			},
			Override: func(e ReconcileEnvelope) *gala.Schedule {
				src := types.IntegrationSourceFrom(e.OperationContext)

				var opSchedule *gala.Schedule

				if reg != nil {
					if op, err := reg.Operation(src.DefinitionID, e.Operation); err == nil {
						opSchedule = op.Schedule
					}
				}

				if !src.Runtime {
					return opSchedule
				}

				override := schedule
				if opSchedule != nil {
					override = *opSchedule
				}

				override.MaxErrorStreak = gala.UnlimitedErrorStreak

				return &override
			},
		},
	}
}

// reconcileShouldCancel reports whether the recurring loop should stop instead of retrying
func reconcileShouldCancel(ctx context.Context, reg *registry.Registry, e ReconcileEnvelope, err error) bool {
	src := types.IntegrationSourceFrom(e.OperationContext)

	if src.IntegrationID != "" && ent.IsNotFound(err) {
		logx.FromContext(ctx).Error().Err(err).Msg("integration not found, not queuing")
		return true
	}

	if errors.Is(err, registry.ErrDefinitionNotFound) || errors.Is(err, registry.ErrOperationNotFound) {
		var registered []string
		if reg != nil {
			registered = lo.Map(reg.Definitions(), func(d types.Definition, _ int) string {
				return d.ID
			})
		}

		logx.FromContext(ctx).Error().Err(err).Strs("registered_definition_ids", registered).Msg("operation no longer registered, stopping cycle")

		return true
	}

	if errors.Is(err, ErrOperationDisabled) {
		logx.FromContext(ctx).Info().Msg("operation disabled, stopping cycle")

		return true
	}

	if unhealthy, ok := types.UnhealthyFrom(err); ok {
		logx.FromContext(ctx).Error().Err(err).Str("reason", unhealthy.Reason).Msg("integration unhealthy, stopping cycle")

		return true
	}

	if degraded, ok := types.DegradedFrom(err); ok {
		logx.FromContext(ctx).Error().Err(err).Str("reason", degraded.Reason).Msg("operation unhealthy, stopping cycle")

		return true
	}

	return false
}
