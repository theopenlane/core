package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const maxIntegrationWebhookBodyBytes = int64(1024 * 1024)

var (
	errIntegrationWebhookNotConfigured = errors.New("integration webhook not configured")
	errIntegrationWebhookAmbiguous     = errors.New("integration webhook ambiguous")
)

// IntegrationWebhookHandler verifies and dispatches one inbound integration webhook event
func (h *Handler) IntegrationWebhookHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	var integrationID string
	if err := echo.PathParamsBinder(ctx).String("integrationID", &integrationID).BindError(); err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}
	if integrationID == "" {
		return h.BadRequest(ctx, rout.MissingField("integrationID"), openapiCtx)
	}

	req := ctx.Request()
	requestCtx := req.Context()
	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	payload, err := io.ReadAll(http.MaxBytesReader(ctx.Response().Writer, req.Body, maxIntegrationWebhookBodyBytes))
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}
	if len(payload) == 0 {
		return h.BadRequest(ctx, errPayloadEmpty, openapiCtx)
	}

	installation, err := h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, integrationID, "")
	if err != nil {
		if errors.Is(err, integrationsruntime.ErrInstallationNotFound) {
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		logx.FromContext(requestCtx).Error().Err(err).Str("integration_id", integrationID).Str("organization_id", caller.OrganizationID).Msg("failed to resolve integration webhook installation")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	definition, ok := h.IntegrationsRuntime.Registry().Definition(installation.DefinitionID)
	if !ok {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	webhook, err := resolveIntegrationWebhook(definition)
	if err != nil {
		if errors.Is(err, errIntegrationWebhookNotConfigured) {
			return h.BadRequest(ctx, err, openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

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
	if webhook.Verify != nil {
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

func resolveIntegrationWebhook(definition types.Definition) (types.WebhookRegistration, error) {
	if len(definition.Webhooks) == 0 {
		return types.WebhookRegistration{}, errIntegrationWebhookNotConfigured
	}

	if len(definition.Webhooks) > 1 {
		return types.WebhookRegistration{}, errIntegrationWebhookAmbiguous
	}

	return definition.Webhooks[0], nil
}
