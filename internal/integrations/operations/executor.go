package operations

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterListeners attaches one Gala listener per registered operation topic
func RegisterListeners(runtime *gala.Gala, reg *registry.Registry, handle func(context.Context, Envelope) error) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, operation := range reg.Listeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic: gala.Topic[Envelope]{
				Name: operation.Topic,
			},
			Name: operation.Name,
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error {
				return handle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

// RegisterWebhookListeners attaches one Gala listener per registered webhook event topic
func RegisterWebhookListeners(runtime *gala.Gala, reg *registry.Registry, handle func(context.Context, WebhookEnvelope) error) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, event := range reg.WebhookListeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[WebhookEnvelope]{
			Topic: gala.Topic[WebhookEnvelope]{
				Name: event.Topic,
			},
			Name: event.Name,
			Handle: func(ctx gala.HandlerContext, envelope WebhookEnvelope) error {
				return handle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

// RegisterRuntimeListeners registers all Gala listeners needed by the integration runtime.
// When operationHandle is nil, operation execution listeners are skipped (use when a
// workflow engine registers its own wrapping listeners on the same topics).
func RegisterRuntimeListeners(runtime *gala.Gala, reg *registry.Registry, operationHandle func(context.Context, Envelope) error, webhookHandle func(context.Context, WebhookEnvelope) error) error {
	if operationHandle != nil {
		if err := RegisterListeners(runtime, reg, operationHandle); err != nil {
			return err
		}
	}

	if err := RegisterIngestListeners(runtime); err != nil {
		return err
	}

	return RegisterWebhookListeners(runtime, reg, webhookHandle)
}
