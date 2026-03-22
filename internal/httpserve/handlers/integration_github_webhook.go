package handlers

import (
	"context"
	"encoding/json"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// GitHubAppWebhookHandler verifies and dispatches one inbound GitHub App webhook delivery
func (h *Handler) GitHubAppWebhookHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	req := ctx.Request()
	requestCtx := req.Context()

	payload, err := readIntegrationWebhookPayload(ctx)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	webhook, err := h.IntegrationsRuntime.Registry().Webhook(githubapp.DefinitionID.ID(), githubapp.InstallationEventsWebhook.Name())
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	// GitHub signs deliveries at the app webhook level using the operator-configured secret,
	// so verify the signature before resolving any local state
	if webhook.Verify != nil {
		if err := webhook.Verify(types.WebhookVerifyRequest{Request: req, Payload: payload}); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Msg("webhook verification failed")
			return h.BadRequest(ctx, err, openapiCtx)
		}
	}

	// GitHub App webhook deliveries are unauthenticated — set privacy bypass and a synthetic caller
	// so that ent queries succeed against privacy-policy-protected tables
	webhookCtx := privacy.DecisionContext(requestCtx, privacy.Allow)
	webhookCtx = auth.WithCaller(webhookCtx, auth.NewWebhookCaller(""))

	installation, err := h.resolveGitHubAppWebhookInstallation(webhookCtx, payload)
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to resolve github app webhook installation")
		return h.BadRequest(ctx, err, openapiCtx)
	}

	webhookCtx = logx.WithFields(webhookCtx, logx.LogFields{
		"integration_id": installation.ID,
		"webhook":        webhook.Name,
	})

	persistedWebhook, err := h.IntegrationsRuntime.EnsureWebhook(webhookCtx, installation, webhook.Name, "")
	if err != nil {
		logx.FromContext(webhookCtx).Error().Err(err).Msg("failed to ensure webhook")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.handleResolvedIntegrationWebhook(ctx, webhookCtx, openapiCtx, installation, webhook, persistedWebhook, payload, true)
}

func (h *Handler) resolveGitHubAppWebhookInstallation(ctx context.Context, payload []byte) (*ent.Integration, error) {
	providerInstallationID, err := extractGitHubAppWebhookProviderInstallationID(payload)
	if err != nil {
		return nil, err
	}

	return h.lookupGitHubAppIntegrationByProviderInstallationID(ctx, "", providerInstallationID)
}

func (h *Handler) lookupGitHubAppIntegrationByProviderInstallationID(ctx context.Context, ownerID, providerInstallationID string) (*ent.Integration, error) {
	query := h.DBClient.Integration.Query().
		Where(
			integration.DefinitionIDEQ(githubapp.DefinitionID.ID()),
			func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(integration.FieldInstallationMetadata, providerInstallationID, sqljson.Path("attributes", "installationId")))
			},
		)
	if ownerID != "" {
		query.Where(integration.OwnerIDEQ(ownerID))
	}

	return query.Only(ctx)
}

func extractGitHubAppWebhookProviderInstallationID(payload []byte) (string, error) {
	var envelope struct {
		Installation *struct {
			ID int64 `json:"id"`
		} `json:"installation"`
	}

	if err := json.Unmarshal(payload, &envelope); err != nil {
		return "", githubapp.ErrWebhookPayloadInvalid
	}

	if envelope.Installation == nil || envelope.Installation.ID == 0 {
		return "", githubapp.ErrInstallationIDMissing
	}

	return strconv.FormatInt(envelope.Installation.ID, 10), nil
}
