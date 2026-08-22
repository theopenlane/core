package operations

import (
	"context"
	"errors"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// ReconcileEnvelope is the durable payload for one recurring operation cycle, either
// installation-bound (IntegrationID set) or runtime-bound (Runtime true); the type name
// is the durable topic identity and must not change
type ReconcileEnvelope struct {
	gala.OperationContext
	// Schedule is the adaptive scheduling state carried across cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// ReconcileUniqueKey derives the insert-time uniqueness key for one recurring loop, so any
// emitter of the topic collapses to at most one live loop per installation (or runtime
// definition) and operation
func ReconcileUniqueKey(e ReconcileEnvelope) string {
	src := types.IntegrationSourceFrom(e.OperationContext)

	return gala.IntegrationReconcile.Key(src.IntegrationID, src.DefinitionID, e.Operation)
}

// ReconcileTopic is the durable reconcile topic: the name derives from the envelope type
// under the reconcile namespace, and every emission carries the loop uniqueness key
var ReconcileTopic = gala.NamespacedTopicFor(gala.IntegrationReconcile, gala.WithUniqueKey(ReconcileUniqueKey))

// LegacyTopicRenames maps the historical reconcile topic to its designated topic
func LegacyTopicRenames() map[gala.TopicName]gala.TopicName {
	return map[gala.TopicName]gala.TopicName{
		// the pre-namespace reconcile topic name
		"integration.ReconcileEnvelope": ReconcileTopic.Name,
	}
}

// ReconcileDefinition builds the Gala listener definition driving every recurring operation
// cycle: installation-bound reconciliation and runtime-bound scheduled operations
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
			// log fields are snapshotted at emit, so a cycle re-emitted without them stays anonymous
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

				// runtime-bound sweeps have no installation to mark unhealthy and no reseed
				// path besides startup, so they back off forever instead of exhausting
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

// reconcileShouldCancel classifies one cycle error, reporting whether the recurring loop should
// stop instead of scheduling another cycle with backoff
func reconcileShouldCancel(ctx context.Context, reg *registry.Registry, e ReconcileEnvelope, err error) bool {
	src := types.IntegrationSourceFrom(e.OperationContext)

	// not-found is terminal only for installation-bound cycles; runtime sweeps
	// surface joined per-item errors that may wrap not-found
	if src.IntegrationID != "" && ent.IsNotFound(err) {
		logx.FromContext(ctx).Error().Err(err).Msg("integration not found, not queuing")
		return true
	}

	if errors.Is(err, registry.ErrDefinitionNotFound) || errors.Is(err, registry.ErrOperationNotFound) {
		// what is registered separates a missing definition from an empty definition id
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

	return false
}
