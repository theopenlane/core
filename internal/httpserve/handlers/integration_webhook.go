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
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// integrationWebhookSignatureHeader is the HTTP header carrying the HMAC-SHA256 webhook signature
	integrationWebhookSignatureHeader = "X-Webhook-Signature-256"
	// maxIntegrationWebhookBodyBytes defines the maximum size of webhook payloads to prevent hex0rz
	maxIntegrationWebhookBodyBytes = int64(1024 * 1024)
)

// IntegrationWebhookHandler verifies and dispatches one inbound integration webhook event.
// The endpoint is addressed by the stable endpoint_id generated at webhook creation time,
// which survives integration record replacement so external callers are not disrupted
func (h *Handler) IntegrationWebhookHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return nil
	}

	endpointID := ctx.PathParam("endpointID")
	req := ctx.Request()

	payload, err := readIntegrationWebhookPayload(ctx)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	// Webhook deliveries are unauthenticated — set privacy bypass and a synthetic caller
	// so that ent queries succeed against privacy-policy-protected tables
	webhookCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	webhookCtx = auth.WithCaller(webhookCtx, auth.NewWebhookCaller(""))

	persistedWebhook, err := h.IntegrationsRuntime.ResolveWebhookByEndpoint(webhookCtx, endpointID)
	if err != nil {
		if !ent.IsNotFound(err) {
			// not finding a record vs. failing to query are different so branching that
			logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to query integration webhook")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	integration, err := h.IntegrationsRuntime.ResolveIntegration(webhookCtx, integrationsruntime.IntegrationLookup{IntegrationID: persistedWebhook.IntegrationID})
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to resolve integration")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	// Re-set the caller now that the owning organization is known
	webhookCtx = auth.WithCaller(webhookCtx, auth.NewWebhookCaller(integration.OwnerID))

	webhookReg, err := h.IntegrationsRuntime.Registry().Webhook(integration.DefinitionID, persistedWebhook.Name)
	if err != nil {
		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	if err := verifyIntegrationWebhook(webhookReg, req, payload, persistedWebhook, integration); err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("webhook signature verification failed")

		return h.BadRequest(ctx, err, openapiCtx)
	}

	return h.handleResolvedIntegrationWebhook(webhookCtx, ctx, openapiCtx, integration, webhookReg, persistedWebhook, payload)
}

// readIntegrationWebhookPayload reads the request body up to a defined maximum size and returns an error if the body is empty or exceeds the limit
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

// verifyWebhookHMACSHA256 validates an inbound webhook request header X-Webhook-Signature-256 header using HMAC-SHA256 (format is "sha256=<hex-encoded HMAC>")
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

// verifyIntegrationWebhook delegates verification to the registration's Verify func when present,
// otherwise falls back to the framework HMAC-SHA256 verification
func verifyIntegrationWebhook(reg types.WebhookRegistration, req *http.Request, payload []byte, persistedWebhook *ent.IntegrationWebhook, integration *ent.Integration) error {
	if reg.Verify != nil {
		return reg.Verify(types.WebhookInboundRequest{
			Integration: integration,
			Webhook:     persistedWebhook,
			Request:     req,
			Payload:     payload,
		})
	}

	return verifyWebhookHMACSHA256(req, payload, persistedWebhook.SecretToken)
}

// handleResolvedIntegrationWebhook processes a webhook with a known integration and webhook registration.
// Callers must verify the request before calling this function
func (h *Handler) handleResolvedIntegrationWebhook(requestCtx context.Context, ctx echo.Context, openapiCtx *OpenAPIContext, integration *ent.Integration, webhook types.WebhookRegistration, persistedWebhook *ent.IntegrationWebhook, payload []byte) error {
	requestCtx = logx.WithFields(requestCtx, logx.LogFields{
		"integration_id": integration.ID,
		"webhook":        webhook.Name,
	})

	if webhook.Event == nil {
		logx.FromContext(requestCtx).Error().Msg("webhook registration missing event resolver")

		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	event, err := webhook.Event(types.WebhookInboundRequest{
		Integration: integration,
		Webhook:     persistedWebhook,
		Request:     ctx.Request(),
		Payload:     payload,
	})
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	if event.Name == "" || len(persistedWebhook.AllowedEvents) > 0 && !lo.Contains(persistedWebhook.AllowedEvents, event.Name) {
		logx.FromContext(requestCtx).Debug().Str("event", event.Name).Msg("webhook event not in allowed list, skipped")
		// event name is our contract for what events we accept; not having one means we return early with 200 to not trigger retries (even types unsupported)
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	duplicate, err := h.IntegrationsRuntime.PrepareWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("delivery_id", event.DeliveryID).Msg("failed to register webhook delivery")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if duplicate {
		logx.FromContext(requestCtx).Debug().Str("delivery_id", event.DeliveryID).Msg("duplicate webhook delivery skipped")

		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(requestCtx, integration, integration.DefinitionID, webhook.Name, event); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("event", event.Name).Msg("failed to dispatch webhook event")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "accepted", nil); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("event", event.Name).Msg("failed to finalize webhook delivery")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
}

// IntegrationStaticWebhookHandler returns a handler for webhooks addressed by a fixed
// definition-level route rather than a per-installation endpoint ID.
// When ResolveIntegration is set, the handler resolves a DB-backed integration and
// delegates to the full handleResolvedIntegrationWebhook pipeline (gala dispatch,
// idempotency tracking, delivery finalization).
// When ResolveIntegration is nil, the webhook is runtime-owned: the handler resolves
// the event, finds the registered handler, and calls it inline with no DB integration
func (h *Handler) IntegrationStaticWebhookHandler(definitionID, webhookName string) func(echo.Context, *OpenAPIContext) error {
	return func(ctx echo.Context, openapiCtx *OpenAPIContext) error {
		if isRegistrationContext(ctx) {
			return nil
		}

		req := ctx.Request()

		payload, err := readIntegrationWebhookPayload(ctx)
		if err != nil {
			return h.BadRequest(ctx, err, openapiCtx)
		}

		webhookReg, err := h.IntegrationsRuntime.Registry().Webhook(definitionID, webhookName)
		if err != nil {
			return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
		}

		webhookCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
		webhookCtx = auth.WithCaller(webhookCtx, auth.NewWebhookCaller(""))

		if webhookReg.Verify != nil {
			if err := webhookReg.Verify(types.WebhookInboundRequest{
				Request: req,
				Payload: payload,
			}); err != nil {
				logx.FromContext(webhookCtx).Error().Err(err).Msg("static webhook verification failed")

				return h.BadRequest(ctx, err, openapiCtx)
			}
		}

		if webhookReg.ResolveIntegration != nil {
			return h.handleStaticWebhookWithIntegration(webhookCtx, ctx, openapiCtx, webhookName, webhookReg, req, payload)
		}

		return h.handleStaticWebhookRuntime(webhookCtx, ctx, openapiCtx, definitionID, webhookName, webhookReg, req, payload)
	}
}

// handleStaticWebhookWithIntegration handles static webhooks backed by a DB integration record.
// Used by definitions like GitHub App where webhooks resolve to a per-customer installation
func (h *Handler) handleStaticWebhookWithIntegration(webhookCtx context.Context, ctx echo.Context, openapiCtx *OpenAPIContext, webhookName string, webhookReg types.WebhookRegistration, req *http.Request, payload []byte) error {
	integration, err := webhookReg.ResolveIntegration(webhookCtx, h.DBClient, types.WebhookInboundRequest{
		Request: req,
		Payload: payload,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
		}

		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to resolve integration for static webhook")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	webhookCtx = auth.WithCaller(webhookCtx, auth.NewWebhookCaller(integration.OwnerID))

	persistedWebhook, err := h.IntegrationsRuntime.EnsureWebhook(webhookCtx, integration, webhookName, "")
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to ensure webhook")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.handleResolvedIntegrationWebhook(webhookCtx, ctx, openapiCtx, integration, webhookReg, persistedWebhook, payload)
}

// handleStaticWebhookRuntime handles static webhooks for runtime-owned definitions
// that have no DB integration record. The event is resolved and dispatched through
// the runtime which owns the DB client and handler lifecycle
func (h *Handler) handleStaticWebhookRuntime(webhookCtx context.Context, ctx echo.Context, openapiCtx *OpenAPIContext, definitionID string, webhookName string, webhookReg types.WebhookRegistration, req *http.Request, payload []byte) error {
	if webhookReg.Event == nil {
		logx.FromContext(webhookCtx).Error().Msg("webhook registration missing event resolver")

		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured, openapiCtx)
	}

	event, err := webhookReg.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: payload,
	})
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	if event.Name == "" {
		return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(webhookCtx, nil, definitionID, webhookName, event); err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Str("event", event.Name).Msg("runtime webhook event dispatch failed")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.Success(ctx, rout.Reply{Success: true}, openapiCtx)
}
