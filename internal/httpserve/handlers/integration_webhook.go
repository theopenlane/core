package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const maxIntegrationWebhookBodyBytes = int64(1024 * 1024)

var (
	errIntegrationWebhookNotConfigured = errors.New("integration webhook not configured")
)

// IntegrationWebhookHandler verifies and dispatches one inbound integration webhook event
func (h *Handler) IntegrationWebhookHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	var (
		integrationID string
		webhookName   string
	)
	if err := echo.PathParamsBinder(ctx).
		String("integrationID", &integrationID).
		String("webhookName", &webhookName).
		BindError(); err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}
	if integrationID == "" {
		return h.BadRequest(ctx, rout.MissingField("integrationID"), openapiCtx)
	}
	if webhookName == "" {
		return h.BadRequest(ctx, rout.MissingField("webhookName"), openapiCtx)
	}

	req := ctx.Request()
	requestCtx := req.Context()
	payload, err := readIntegrationWebhookPayload(ctx)
	if err != nil {
		if errors.Is(err, errPayloadEmpty) {
			return h.BadRequest(ctx, err, openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	installation, err := h.IntegrationsRuntime.ResolveInstallation(requestCtx, "", integrationID, "")
	if err != nil {
		if errors.Is(err, integrationsruntime.ErrInstallationNotFound) {
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		logx.FromContext(requestCtx).Error().Err(err).Str("integration_id", integrationID).Msg("failed to resolve integration webhook installation")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	webhook, err := h.IntegrationsRuntime.Registry().Webhook(installation.DefinitionID, webhookName)
	if err != nil {
		if errors.Is(err, registry.ErrWebhookNotFound) {
			return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.handleResolvedIntegrationWebhook(ctx, openapiCtx, installation, webhook, payload, false)
}

func readIntegrationWebhookPayload(ctx echo.Context) ([]byte, error) {
	payload, err := io.ReadAll(http.MaxBytesReader(ctx.Response().Writer, ctx.Request().Body, maxIntegrationWebhookBodyBytes))
	if err != nil {
		return nil, err
	}

	if len(payload) == 0 {
		return nil, errPayloadEmpty
	}

	return payload, nil
}

func (h *Handler) handleResolvedIntegrationWebhook(ctx echo.Context, openapiCtx *OpenAPIContext, installation *ent.Integration, webhook types.WebhookRegistration, payload []byte, skipVerify bool) error {
	req := ctx.Request()
	requestCtx := req.Context()

	persistedWebhook, err := h.IntegrationsRuntime.EnsureWebhook(requestCtx, installation, webhook.Name)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("integration_id", installation.ID).Str("webhook", webhook.Name).Msg("failed to ensure integration webhook")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	verifyRequest := types.WebhookVerifyRequest{
		Integration: installation,
		Webhook:     persistedWebhook,
		Request:     req,
		Payload:     payload,
	}
	if !skipVerify && webhook.Verify != nil {
		if err := webhook.Verify(requestCtx, verifyRequest); err != nil {
			return h.BadRequest(ctx, err, openapiCtx)
		}
	}

	if webhook.Event == nil {
		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	event, err := webhook.Event(requestCtx, types.WebhookEventRequest{
		Integration: installation,
		Webhook:     persistedWebhook,
		Request:     req,
		Payload:     payload,
	})
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	if event.Name == "" {
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if len(persistedWebhook.AllowedEvents) > 0 && !lo.Contains(persistedWebhook.AllowedEvents, event.Name) {
		if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "ignored", nil); err != nil {
			logx.FromContext(requestCtx).Warn().Err(err).Str("integration_id", installation.ID).Str("webhook", webhook.Name).Str("event", event.Name).Msg("failed to finalize ignored integration webhook delivery")
		}

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	duplicate, err := h.IntegrationsRuntime.PrepareWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("integration_id", installation.ID).Str("webhook", webhook.Name).Str("delivery_id", event.DeliveryID).Msg("failed to register integration webhook delivery")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if duplicate {
		if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "duplicate", nil); err != nil {
			logx.FromContext(requestCtx).Warn().Err(err).Str("integration_id", installation.ID).Str("webhook", webhook.Name).Str("delivery_id", event.DeliveryID).Msg("failed to finalize duplicate integration webhook delivery")
		}

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(requestCtx, installation, webhook.Name, event); err != nil {
		_ = h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "failed", err)
		logx.FromContext(requestCtx).Error().Err(err).Str("integration_id", installation.ID).Str("webhook", webhook.Name).Str("event", event.Name).Msg("failed to dispatch integration webhook event")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "accepted", nil); err != nil {
		logx.FromContext(requestCtx).Warn().Err(err).Str("integration_id", installation.ID).Str("webhook", webhook.Name).Str("event", event.Name).Msg("failed to finalize integration webhook delivery")
	}

	return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
}
