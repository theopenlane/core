package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const maxIntegrationWebhookBodyBytes = int64(1024 * 1024)

var (
	errIntegrationWebhookNotConfigured = errors.New("integration webhook not configured")
)

// IntegrationWebhookHandler verifies and dispatches one inbound integration webhook event.
// The endpoint is addressed by the stable endpoint_id generated at webhook creation time,
// which survives integration record replacement so external callers are not disrupted.
func (h *Handler) IntegrationWebhookHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	var endpointID string
	if err := echo.PathParamsBinder(ctx).
		String("endpointID", &endpointID).
		BindError(); err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}
	if endpointID == "" {
		return h.BadRequest(ctx, rout.MissingField("endpointID"), openapiCtx)
	}

	req := ctx.Request()
	requestCtx := req.Context()
	logger := logx.FromContext(requestCtx).With().Str("endpoint_id", endpointID).Logger()

	payload, err := readIntegrationWebhookPayload(ctx)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	persistedWebhook, err := h.DBClient.IntegrationWebhook.Query().
		Where(
			integrationwebhook.EndpointIDEQ(endpointID),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		Only(requestCtx)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error().Err(err).Msg("failed to query integration webhook")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	installation, err := h.IntegrationsRuntime.ResolveInstallation(requestCtx, "", persistedWebhook.IntegrationID, "")
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	webhookReg, err := h.IntegrationsRuntime.Registry().Webhook(installation.DefinitionID, persistedWebhook.Name)
	if err != nil {
		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	return h.handleResolvedIntegrationWebhook(ctx, openapiCtx, installation, webhookReg, persistedWebhook, payload, false)
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

func (h *Handler) handleResolvedIntegrationWebhook(ctx echo.Context, openapiCtx *OpenAPIContext, installation *ent.Integration, webhook types.WebhookRegistration, persistedWebhook *ent.IntegrationWebhook, payload []byte, skipVerify bool) error {
	req := ctx.Request()
	requestCtx := req.Context()
	logger := logx.FromContext(requestCtx).With().
		Str("integration_id", installation.ID).
		Str("webhook", webhook.Name).
		Logger()

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
			logger.Warn().Err(err).Str("event", event.Name).Msg("failed to finalize ignored webhook delivery")
		}

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	duplicate, err := h.IntegrationsRuntime.PrepareWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID)
	if err != nil {
		logger.Error().Err(err).Str("delivery_id", event.DeliveryID).Msg("failed to register webhook delivery")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if duplicate {
		if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "duplicate", nil); err != nil {
			logger.Warn().Err(err).Str("delivery_id", event.DeliveryID).Msg("failed to finalize duplicate webhook delivery")
		}

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(requestCtx, installation, webhook.Name, event); err != nil {
		_ = h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "failed", err)
		logger.Error().Err(err).Str("event", event.Name).Msg("failed to dispatch webhook event")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "accepted", nil); err != nil {
		logger.Warn().Err(err).Str("event", event.Name).Msg("failed to finalize webhook delivery")
	}

	return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
}
