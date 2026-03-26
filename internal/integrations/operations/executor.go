package operations

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterRuntimeListeners registers all Gala listeners needed by the integration runtime
func RegisterRuntimeListeners(runtime *gala.Gala, reg *registry.Registry, operationHandle func(context.Context, Envelope) error, webhookHandle func(context.Context, WebhookEnvelope) error) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, operation := range reg.Listeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic:  gala.Topic[Envelope]{Name: operation.Topic},
			Name:   operation.Name,
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error { return operationHandle(ctx.Context, envelope) },
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

	return nil
}
