package handlers

import (
	"fmt"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ConfigureIntegrationProvider stores non-OAuth credentials for a provider definition.
// When installation_id is provided the credentials on that installation are updated.
// When omitted a new installation is created and its ID is returned in the response
func (h *Handler) ConfigureIntegrationProvider(ctx echo.Context) error {
	payload, err := BindAndValidate[ConfigureIntegrationRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(payload.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider)
	}

	installationRec, isNewInstallation, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, payload.IntegrationID, def, payload.UserInput, payload.OperationConfig)
	if err != nil {
		// do not log payload, it can contain secrets
		logx.FromContext(requestCtx).Error().Err(err).Msg("failed to ensure installation")

		return h.BadRequest(ctx, err)
	}

	if payload.HasCredentialBody() {
		if err := h.IntegrationsRuntime.ReconcileCredential(requestCtx, installationRec, payload.CredentialRef, types.CredentialSet{Data: jsonx.CloneRawMessage(payload.Body)}); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Msg("credential reconcile failed")

			return h.BadRequest(ctx, err)
		}
	}

	resp := ConfigureIntegrationResponse{
		Success:              true,
		Provider:             def.ID,
		IntegrationID:        installationRec.ID,
		HealthStatus:         "ok",
		HealthSummary:        "Validation passed during configuration",
		InstallationMetadata: installationRec.InstallationMetadata.Attributes,
	}

	primaryWebhookURL, primaryWebhookSecret, err := h.ensureInstallationWebhooks(ctx, installationRec, def)
	if err != nil {
		return h.BadRequest(ctx, ErrProcessingRequest)
	}

	if isNewInstallation {
		resp.WebhookEndpointURL = primaryWebhookURL
		resp.WebhookSecret = primaryWebhookSecret
	}

	// ensure all reconcile jobs exist after any config update; a previously-disabled
	// operation that was just re-enabled needs a new job seeded - this is a no-op
	// when all jobs are already active
	if lo.Contains(enums.IntegrationOperationalStatuses, installationRec.Status) {
		if err := h.IntegrationsRuntime.ResetReconcileLoops(requestCtx, installationRec); err != nil {
			logx.FromContext(requestCtx).Warn().Err(err).Str("installation_id", installationRec.ID).Msg("failed to seed missing reconcile jobs after config update")
		}
	}

	return h.Success(ctx, resp)
}

// ensureInstallationWebhooks ensures every declared webhook exists for the installation and returns the first webhook's absolute endpoint URL and secret
func (h *Handler) ensureInstallationWebhooks(ctx echo.Context, installationRec *ent.Integration, def types.Definition) (string, string, error) {
	requestCtx := ctx.Request().Context()

	var primaryWebhookURL, primaryWebhookSecret string

	for i, registration := range def.Webhooks {
		webhook, err := h.IntegrationsRuntime.EnsureWebhook(requestCtx, installationRec, registration.Name, "")
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Str("webhook", registration.Name).Msg("failed to ensure installation webhook")

			return "", "", err
		}

		if i == 0 && webhook != nil {
			primaryWebhookURL = absoluteEndpointURL(ctx, lo.FromPtr(webhook.EndpointURL))
			primaryWebhookSecret = webhook.SecretToken
		}
	}

	return primaryWebhookURL, primaryWebhookSecret, nil
}

func absoluteEndpointURL(ctx echo.Context, path string) string {
	if path == "" {
		return ""
	}

	if path[0] != '/' {
		return path
	}

	return fmt.Sprintf("%s://%s%s", ctx.Scheme(), ctx.Request().Host, path)
}
