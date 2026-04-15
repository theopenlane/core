package operations

import (
	"context"

	"github.com/riverqueue/river"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterRuntimeListeners registers all Gala listeners needed by the integration runtime
func RegisterRuntimeListeners(runtime *gala.Gala, reg *registry.Registry, operationHandle func(context.Context, Envelope) error, webhookHandle func(context.Context, WebhookEnvelope) error, reconcileHandle ReconcileHandler, reconcileSchedule gala.Schedule) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, operation := range reg.Listeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic: gala.Topic[Envelope]{Name: operation.Topic},
			Name:  operation.Name,
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error {
				err := operationHandle(ctx.Context, envelope)
				if ent.IsNotFound(err) {
					logx.FromContext(ctx.Context).Error().Err(err).Str("integration_id", envelope.IntegrationID).Msg("integration not found, cancelling operation")
					return river.JobCancel(err)
				}

				return err
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
				err := webhookHandle(ctx.Context, envelope)
				if ent.IsNotFound(err) {
					logx.FromContext(ctx.Context).Error().Err(err).Str("integration_id", envelope.IntegrationID).Msg("integration not found, cancelling webhook event")
					return river.JobCancel(err)
				}

				return err
			},
		}); err != nil {
			return err
		}
	}

	if err := RegisterReconcileListener(runtime, reg, reconcileHandle, reconcileSchedule); err != nil {
		return err
	}

	for _, listener := range reg.GalaListeners() {
		if _, err := listener.Register(runtime.Registry()); err != nil {
			return err
		}
	}

	return nil
}
