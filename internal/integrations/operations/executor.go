package operations

import (
	"context"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// MutationListenerFunc handles a mutation event for a registered mutation listener
type MutationListenerFunc func(ctx context.Context, listener types.MutationListenerRegistration, payload eventqueue.MutationGalaPayload) error

// RegisterRuntimeListeners registers all Gala listeners needed by the integration runtime
func RegisterRuntimeListeners(runtime *gala.Gala, reg *registry.Registry, operationHandle func(context.Context, Envelope) error, webhookHandle func(context.Context, WebhookEnvelope) error, reconcileHandle ReconcileHandler, reconcileSchedule gala.Schedule, mutationHandle MutationListenerFunc) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, operation := range reg.Listeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic: gala.Topic[Envelope]{Name: operation.Topic},
			Name:  operation.Name,
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error {
				return operationHandle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	if err := RegisterIngestListeners(runtime); err != nil {
		return err
	}

	for _, event := range reg.WebhookListeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[WebhookEnvelope]{
			Topic: gala.Topic[WebhookEnvelope]{Name: event.Topic},
			Name:  event.Name,
			Handle: func(ctx gala.HandlerContext, envelope WebhookEnvelope) error {
				return webhookHandle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	if err := RegisterReconcileListener(runtime, reconcileHandle, reconcileSchedule); err != nil {
		return err
	}

	for _, listener := range reg.MutationListeners() {
		topic := eventqueue.MutationTopic(eventqueue.MutationConcernDirect, listener.SchemaType)

		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: topic,
			Name:  listener.Name,
			Handle: func(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
				return mutationHandle(ctx.Context, listener, payload)
			},
		}); err != nil {
			return err
		}
	}

	for _, listener := range reg.GalaListeners() {
		if _, err := listener.Register(runtime.Registry()); err != nil {
			return err
		}
	}

	return nil
}
