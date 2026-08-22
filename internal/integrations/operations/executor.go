package operations

import (
	"context"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// integrationEnvelope constrains the envelope types carrying the integration installation as the operation entity
type integrationEnvelope interface {
	Envelope | WebhookEnvelope
}

// integrationNotFoundCancel builds a Cancel predicate that cancels and logs when the envelope's integration no longer exists
func integrationNotFoundCancel[T integrationEnvelope](message string) func(context.Context, T, error) bool {
	return func(ctx context.Context, envelope T, err error) bool {
		if !ent.IsNotFound(err) {
			return false
		}

		var integrationID string

		switch e := any(envelope).(type) {
		case Envelope:
			integrationID = e.EntityID
		case WebhookEnvelope:
			integrationID = e.EntityID
		}

		logx.FromContext(ctx).Error().Err(err).Str("integration_id", integrationID).Msg(message)

		return true
	}
}

// RegisterRuntimeListeners registers the event, webhook, and definition-provided gala listeners for
// the integration runtime. Adaptive-scheduled pollers (reconcile, scheduled operations) register
// themselves at the call site via their own Register*Listener functions
func RegisterRuntimeListeners(runtime *gala.Gala, reg *registry.Registry, services types.RuntimeServices, operationHandle func(context.Context, Envelope) error, webhookHandle func(context.Context, WebhookEnvelope) error) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, operation := range reg.Listeners() {
		if _, err := gala.Register(runtime, gala.Definition[Envelope]{
			Topic:  gala.Topic[Envelope]{Name: operation.Topic, Kind: gala.IntegrationRun.Kind()},
			Name:   operation.Name,
			Cancel: integrationNotFoundCancel[Envelope]("integration not found, cancelling operation"),
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error {
				return operationHandle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	if err := bindIngestPersistence(); err != nil {
		return err
	}

	if err := entityops.RegisterIngestListeners(runtime, func(ctx context.Context, client *ent.Client, _ gala.OperationContext) (*ent.Integration, error) {
		return resolveIngestIntegration(ctx, client)
	}); err != nil {
		return err
	}

	for _, event := range reg.WebhookListeners() {
		if _, err := gala.Register(runtime, gala.Definition[WebhookEnvelope]{
			Topic:  gala.Topic[WebhookEnvelope]{Name: event.Topic, Kind: gala.IntegrationWebhook.Kind()},
			Name:   event.Name,
			Cancel: integrationNotFoundCancel[WebhookEnvelope]("integration not found, cancelling webhook event"),
			Handle: func(ctx gala.HandlerContext, envelope WebhookEnvelope) error {
				return webhookHandle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	for _, listener := range reg.GalaListeners() {
		if _, err := listener.Register(runtime, services); err != nil {
			return err
		}
	}

	return nil
}
