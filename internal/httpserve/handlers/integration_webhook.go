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

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationsruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// integrationWebhookSignatureHeader is the HTTP header carrying the HMAC-SHA256 webhook signature
	integrationWebhookSignatureHeader = "X-Webhook-Signature-256"
	// maxIntegrationWebhookBodyBytes defines the maximum size of webhook payloads to prevent hex0rz
	maxIntegrationWebhookBodyBytes = int64(1024 * 1024)
)

// IntegrationWebhookHandler verifies and dispatches one inbound integration webhook event, addressed by the stable endpoint_id generated at webhook creation time so it survives integration record replacement without disrupting external callers
func (h *Handler) IntegrationWebhookHandler(ctx echo.Context) error {
	endpointID := ctx.PathParam("endpointID")
	req := ctx.Request()

	payload, err := readIntegrationWebhookPayload(ctx)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// deliveries are unauthenticated and the owning org is not known until the endpoint resolves, so the lookup is a cross-org internal read
	webhookCtx := auth.WithInternalReadCrossOrgContext(req.Context())

	persistedWebhook, err := h.IntegrationsRuntime.ResolveWebhookByEndpoint(webhookCtx, endpointID)
	if err != nil {
		if !ent.IsNotFound(err) {
			logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to query integration webhook")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	integration, err := h.IntegrationsRuntime.ResolveIntegration(webhookCtx, integrationsruntime.IntegrationLookup{IntegrationID: persistedWebhook.IntegrationID})
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to resolve integration")

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	// Re-set the caller now that the owning organization is known
	webhookCtx = auth.WithOrgInternalCaller(webhookCtx, integration.OwnerID)

	webhookReg, found := resolveIntegrationWebhook(h.IntegrationsRuntime.Registry(), integration.DefinitionID, persistedWebhook.Name)
	if !found {
		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured)
	}

	if err := verifyIntegrationWebhook(webhookReg, req, payload, persistedWebhook, integration); err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("webhook signature verification failed")

		return h.BadRequest(ctx, err)
	}

	return h.handleResolvedIntegrationWebhook(webhookCtx, ctx, integration, webhookReg, persistedWebhook, payload)
}

// resolveIntegrationWebhook finds the webhook registration a persisted row is named after, falling back to the registration whose ref replaces that name when the row predates a rename
func resolveIntegrationWebhook(reg *registry.Registry, definitionID, name string) (types.WebhookRegistration, bool) {
	webhookReg, err := reg.Webhook(definitionID, name)
	if err == nil {
		return webhookReg, true
	}

	def, ok := reg.Definition(definitionID)
	if !ok {
		return types.WebhookRegistration{}, false
	}

	return def.WebhookReplacing(name)
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

// verifyIntegrationWebhook delegates verification to the registration's Verify func when present, otherwise falls back to the framework HMAC-SHA256 verification
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

// handleResolvedIntegrationWebhook processes a webhook with a known integration and webhook registration; callers must verify the request before calling this function
func (h *Handler) handleResolvedIntegrationWebhook(requestCtx context.Context, ctx echo.Context, integration *ent.Integration, webhook types.WebhookRegistration, persistedWebhook *ent.IntegrationWebhook, payload []byte) error {
	requestCtx = logx.WithFields(requestCtx, logx.LogFields{
		"integration_id": integration.ID,
		"webhook":        webhook.Name,
	})

	if webhook.Event == nil {
		logx.FromContext(requestCtx).Error().Msg("webhook registration missing event resolver")

		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured)
	}

	event, err := webhook.Event(types.WebhookInboundRequest{
		Integration: integration,
		Webhook:     persistedWebhook,
		Request:     ctx.Request(),
		Payload:     payload,
	})
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	if event.Name == "" || len(persistedWebhook.AllowedEvents) > 0 && !lo.Contains(persistedWebhook.AllowedEvents, event.Name) {
		logx.FromContext(requestCtx).Debug().Str("event", event.Name).Msg("webhook event not in allowed list, skipped")
		return h.Success(ctx, rout.Reply{Success: true})
	}

	duplicate, err := h.IntegrationsRuntime.PrepareWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("delivery_id", event.DeliveryID).Msg("failed to register webhook delivery")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if duplicate {
		logx.FromContext(requestCtx).Debug().Str("delivery_id", event.DeliveryID).Msg("duplicate webhook delivery skipped")

		return h.Success(ctx, rout.Reply{Success: true})
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(requestCtx, integration, integration.DefinitionID, webhook.Name, event); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("event", event.Name).Msg("failed to dispatch webhook event")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if err := h.IntegrationsRuntime.FinalizeWebhookDelivery(requestCtx, persistedWebhook, event.DeliveryID, "accepted", nil); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("event", event.Name).Msg("failed to finalize webhook delivery")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, rout.Reply{Success: true})
}

// IntegrationStaticWebhookHandler returns a handler for webhooks addressed by a fixed definition-level route rather than a per-installation endpoint ID, delegating to the full handleResolvedIntegrationWebhook pipeline when ResolveIntegration is set or resolving and dispatching the event inline with no DB integration when it is runtime-owned
func (h *Handler) IntegrationStaticWebhookHandler(definitionID, webhookName string) func(echo.Context) error {
	return func(ctx echo.Context) error {
		req := ctx.Request()

		payload, err := readIntegrationWebhookPayload(ctx)
		if err != nil {
			return h.BadRequest(ctx, err)
		}

		webhookReg, err := h.IntegrationsRuntime.Registry().Webhook(definitionID, webhookName)
		if err != nil {
			return h.BadRequest(ctx, errIntegrationWebhookNotConfigured)
		}

		// the owning org, if any, is not known until the integration resolves, so this is a cross-org internal context
		webhookCtx := auth.WithInternalCrossOrgContext(req.Context())

		if webhookReg.Verify != nil {
			if err := webhookReg.Verify(types.WebhookInboundRequest{
				Request: req,
				Payload: payload,
			}); err != nil {
				logx.FromContext(webhookCtx).Error().Err(err).Msg("static webhook verification failed")

				return h.BadRequest(ctx, err)
			}
		}

		if webhookReg.ResolveIntegration != nil {
			return h.handleStaticWebhookWithIntegration(webhookCtx, ctx, webhookName, webhookReg, req, payload)
		}

		return h.handleStaticWebhookRuntime(webhookCtx, ctx, definitionID, webhookName, webhookReg, req, payload)
	}
}

// handleStaticWebhookWithIntegration handles static webhooks backed by a DB integration record, used by definitions like GitHub App where webhooks resolve to a per-customer installation
func (h *Handler) handleStaticWebhookWithIntegration(webhookCtx context.Context, ctx echo.Context, webhookName string, webhookReg types.WebhookRegistration, req *http.Request, payload []byte) error {
	integration, err := webhookReg.ResolveIntegration(webhookCtx, h.DBClient, types.WebhookInboundRequest{
		Request: req,
		Payload: payload,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return h.Success(ctx, rout.Reply{Success: true})
		}

		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to resolve integration for static webhook")

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	webhookCtx = auth.WithOrgInternalCaller(webhookCtx, integration.OwnerID)

	persistedWebhook, err := h.IntegrationsRuntime.EnsureWebhook(webhookCtx, integration, webhookName, "")
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to ensure webhook")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.handleResolvedIntegrationWebhook(webhookCtx, ctx, integration, webhookReg, persistedWebhook, payload)
}

// handleStaticWebhookRuntime handles static webhooks for runtime-owned definitions that have no DB integration record, resolving and dispatching the event through the runtime which owns the DB client and handler lifecycle
func (h *Handler) handleStaticWebhookRuntime(webhookCtx context.Context, ctx echo.Context, definitionID string, webhookName string, webhookReg types.WebhookRegistration, req *http.Request, payload []byte) error {
	if webhookReg.Event == nil {
		logx.FromContext(webhookCtx).Error().Msg("webhook registration missing event resolver")

		return h.BadRequest(ctx, errIntegrationWebhookNotConfigured)
	}

	event, err := webhookReg.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: payload,
	})
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	if event.Name == "" {
		return h.Success(ctx, rout.Reply{Success: true})
	}

	if err := h.IntegrationsRuntime.DispatchWebhookEvent(webhookCtx, nil, definitionID, webhookName, event); err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Str("event", event.Name).Msg("runtime webhook event dispatch failed")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, rout.Reply{Success: true})
}
