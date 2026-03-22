package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const maxIntegrationWebhookBodyBytes = int64(1024 * 1024)

const (
	// integrationWebhookSignatureHeader is the HTTP header carrying the HMAC-SHA256 webhook signature
	integrationWebhookSignatureHeader = "X-Webhook-Signature-256"
)

// IntegrationWebhookHandler verifies and dispatches one inbound integration webhook event.
// The endpoint is addressed by the stable endpoint_id generated at webhook creation time,
// which survives integration record replacement so external callers are not disrupted
func (h *Handler) IntegrationWebhookHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	endpointID := ctx.PathParam("endpointID")
	if endpointID == "" {
		return h.BadRequest(ctx, rout.MissingField("endpointID"), openapiCtx)
	}

	req := ctx.Request()

	payload, err := readIntegrationWebhookPayload(ctx)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	// Webhook deliveries are unauthenticated — set privacy bypass and a synthetic caller
	// so that ent queries succeed against privacy-policy-protected tables
	webhookCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	webhookCtx = auth.WithCaller(webhookCtx, auth.NewWebhookCaller(""))
	webhookCtx = logx.WithField(webhookCtx, "endpoint_id", endpointID)

	persistedWebhook, err := h.IntegrationsRuntime.ResolveWebhookByEndpoint(webhookCtx, endpointID)
	if err != nil {
		if !ent.IsNotFound(err) {
			logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to query integration webhook")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	// Verify the webhook signature using the persisted secret token before any further processing
	if err := verifyWebhookHMACSHA256(req, payload, persistedWebhook.SecretToken); err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("webhook signature verification failed")
		return h.BadRequest(ctx, err, openapiCtx)
	}

	installation, err := h.IntegrationsRuntime.ResolveInstallation(webhookCtx, "", persistedWebhook.IntegrationID, "")
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	webhookReg, err := h.IntegrationsRuntime.Registry().Webhook(installation.DefinitionID, persistedWebhook.Name)
	if err != nil {
		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	return h.handleResolvedIntegrationWebhook(ctx, webhookCtx, openapiCtx, installation, webhookReg, persistedWebhook, payload, false)
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

// verifyWebhookHMACSHA256 validates an inbound webhook request using HMAC-SHA256.
// The expected header format is "sha256=<hex-encoded HMAC>" in the X-Webhook-Signature-256 header
func verifyWebhookHMACSHA256(req *http.Request, payload []byte, secret string) error {
	if secret == "" {
		return errIntegrationWebhookSecretMissing
	}

	signature := req.Header.Get(integrationWebhookSignatureHeader)
	if signature == "" {
		return errIntegrationWebhookSignatureMissing
	}

	sigHex, found := strings.CutPrefix(signature, "sha256=")
	if !found {
		return errIntegrationWebhookSignatureMismatch
	}

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return errIntegrationWebhookSignatureMismatch
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)

	if !hmac.Equal(sigBytes, mac.Sum(nil)) {
		return errIntegrationWebhookSignatureMismatch
	}

	return nil
}

func (h *Handler) handleResolvedIntegrationWebhook(ctx echo.Context, requestCtx context.Context, openapiCtx *OpenAPIContext, installation *ent.Integration, webhook types.WebhookRegistration, persistedWebhook *ent.IntegrationWebhook, payload []byte, skipVerify bool) error {
	requestCtx = logx.WithFields(requestCtx, logx.LogFields{
		"integration_id": installation.ID,
		"webhook":        webhook.Name,
	})

	if !skipVerify && webhook.Verify != nil {
		if err := webhook.Verify(types.WebhookVerifyRequest{
			Integration: installation,
			Webhook:     persistedWebhook,
			Request:     ctx.Request(),
			Payload:     payload,
		}); err != nil {
			return h.BadRequest(ctx, err, openapiCtx)
		}
	}

	if webhook.Event == nil {
		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	event, err := webhook.Event(types.WebhookEventRequest{
		Integration: installation,
		Webhook:     persistedWebhook,
		Request:     ctx.Request(),
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
			logx.FromContext(requestCtx).Warn().Err(err).Str("event", event.Name).Msg("failed to finalize ignored webhook delivery")
		}

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if event.DeliveryID == "" {
		logx.FromContext(requestCtx).Warn().Str("event", event.Name).Msg("webhook delivery missing idempotency key, skipping duplicate check")
	}

	duplicate, err := h.IntegrationsRuntime.PrepareWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("delivery_id", event.DeliveryID).Msg("failed to register webhook delivery")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if duplicate {
		if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "duplicate", nil); err != nil {
			logx.FromContext(requestCtx).Warn().Err(err).Str("delivery_id", event.DeliveryID).Msg("failed to finalize duplicate webhook delivery")
		}

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(requestCtx, installation, webhook.Name, event); err != nil {
		_ = h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "failed", err)
		logx.FromContext(requestCtx).Error().Err(err).Str("event", event.Name).Msg("failed to dispatch webhook event")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "accepted", nil); err != nil {
		logx.FromContext(requestCtx).Warn().Err(err).Str("event", event.Name).Msg("failed to finalize webhook delivery")
	}

	return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
}
