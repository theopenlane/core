package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// integrationEnvelope constrains the envelope types carrying the integration installation as the operation entity
type integrationEnvelope interface {
	operations.Envelope | operations.WebhookEnvelope
}

// integrationNotFoundCancel builds a Cancel predicate that cancels and logs when the envelope's integration no longer exists
func integrationNotFoundCancel[T integrationEnvelope](message string) func(context.Context, T, error) bool {
	return func(ctx context.Context, envelope T, err error) bool {
		if !ent.IsNotFound(err) {
			return false
		}

		var integrationID string

		switch e := any(envelope).(type) {
		case operations.Envelope:
			integrationID = e.EntityID
		case operations.WebhookEnvelope:
			integrationID = e.EntityID
		}

		logx.FromContext(ctx).Error().Err(err).Str(intobvs.FieldIntegrationID, integrationID).Msg(message)

		return true
	}
}

// registerListeners registers the operation, ingest, webhook, reconcile, and definition-provided gala listeners
func (r *Runtime) registerListeners() error {
	runtime := r.Gala()
	if runtime == nil {
		return operations.ErrGalaRequired
	}

	reg := r.Registry()

	for _, operation := range reg.Listeners() {
		if _, err := gala.Register(runtime, gala.Definition[operations.Envelope]{
			Topic:  gala.Topic[operations.Envelope]{Name: operation.Topic, Kind: gala.IntegrationRun.Kind()},
			Name:   operation.Name,
			Cancel: integrationNotFoundCancel[operations.Envelope]("integration not found, cancelling operation"),
			Handle: func(ctx gala.HandlerContext, envelope operations.Envelope) error {
				return r.HandleOperation(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	if err := entityops.RegisterIngestListeners(runtime, r.resolveIngestIntegration); err != nil {
		return err
	}

	for _, event := range reg.WebhookListeners() {
		if _, err := gala.Register(runtime, gala.Definition[operations.WebhookEnvelope]{
			Topic:  gala.Topic[operations.WebhookEnvelope]{Name: event.Topic, Kind: gala.IntegrationWebhook.Kind()},
			Name:   event.Name,
			Cancel: integrationNotFoundCancel[operations.WebhookEnvelope]("integration not found, cancelling webhook event"),
			Handle: func(ctx gala.HandlerContext, envelope operations.WebhookEnvelope) error {
				return r.HandleWebhookEvent(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	for _, listener := range reg.GalaListeners() {
		if _, err := listener.Register(runtime, r); err != nil {
			return err
		}
	}

	_, err := gala.Register(runtime, r.reconcileDefinition(gala.Schedule{}))

	return err
}

// resolveIngestIntegration loads the installation an ingest command targets, cancelling the job when it was removed
func (r *Runtime) resolveIngestIntegration(ctx context.Context, _ *ent.Client, oc gala.OperationContext) (*ent.Integration, error) {
	if oc.EntityID == "" {
		return nil, operations.ErrIngestIntegrationUnresolved
	}

	integration, err := r.resolveCurrentIntegration(ctx, IntegrationLookup{IntegrationID: oc.EntityID})
	if ent.IsNotFound(err) {
		return nil, river.JobCancel(fmt.Errorf("%w: %w", operations.ErrIngestIntegrationRemoved, err))
	}

	return integration, err
}

// reconcileDefinition builds the gala listener definition driving every recurring operation cycle
func (r *Runtime) reconcileDefinition(schedule gala.Schedule) gala.Definition[operations.ReconcileEnvelope] {
	reg := r.Registry()

	return gala.Definition[operations.ReconcileEnvelope]{
		Topic: operations.ReconcileTopic,
		Cancel: func(ctx context.Context, e operations.ReconcileEnvelope, err error) bool {
			return reconcileShouldCancel(ctx, reg, e, err)
		},
		OnExhausted: r.markReconcileExhausted,
		Schedule: &gala.ScheduleSpec[operations.ReconcileEnvelope]{
			Schedule: schedule,
			Handle:   r.HandleReconcile,
			State:    func(e operations.ReconcileEnvelope) gala.ScheduleState { return e.Schedule },
			Wrap: func(e operations.ReconcileEnvelope, s gala.ScheduleState) operations.ReconcileEnvelope {
				return operations.ReconcileEnvelope{
					OperationContext: e.OperationContext,
					Schedule:         s,
				}
			},
			PrepareEmit: func(ctx context.Context, e operations.ReconcileEnvelope) (context.Context, gala.Headers) {
				return intobvs.EmitContext(ctx, e.OperationContext)
			},
			Override: func(e operations.ReconcileEnvelope) *gala.Schedule {
				src := types.IntegrationSourceFrom(e.OperationContext)

				var opSchedule *gala.Schedule

				if op, err := reg.Operation(src.DefinitionID, e.Operation); err == nil {
					opSchedule = op.Schedule
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

// reconcileShouldCancel classifies one cycle error, reporting whether the recurring loop should stop instead of backing off
func reconcileShouldCancel(ctx context.Context, reg *registry.Registry, e operations.ReconcileEnvelope, err error) bool {
	src := types.IntegrationSourceFrom(e.OperationContext)

	if src.IntegrationID != "" && ent.IsNotFound(err) {
		logx.FromContext(ctx).Error().Err(err).Msg("integration not found, not queuing")

		return true
	}

	if errors.Is(err, registry.ErrDefinitionNotFound) || errors.Is(err, registry.ErrOperationNotFound) {
		registered := lo.Map(reg.Definitions(), func(d types.Definition, _ int) string {
			return d.ID
		})

		logx.FromContext(ctx).Error().Err(err).Strs("registered_definition_ids", registered).Msg("operation no longer registered, stopping cycle")

		return true
	}

	if errors.Is(err, operations.ErrOperationDisabled) {
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
