package handlers

import (
	"fmt"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// ConfigureIntegrationProvider stores non-OAuth credentials for a provider definition.
// When installation_id is provided the credentials on that installation are updated.
// When omitted a new installation is created and its ID is returned in the response
func (h *Handler) ConfigureIntegrationProvider(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	payload, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationConfigPayload, IntegrationConfigResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(payload.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	installationRec, isNewInstallation, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, payload.IntegrationID, def)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("payload", payload).Msg("failed to resolve installation")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	var credential *types.CredentialSet

	if !jsonx.IsEmptyRawMessage(payload.Body) {
		credential = &types.CredentialSet{Data: jsonx.CloneRawMessage(payload.Body)}
	}

	if err := h.IntegrationsRuntime.Reconcile(requestCtx, installationRec, payload.UserInput, payload.CredentialRef, credential, nil); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("payload", payload).Msg("reconcile failed")

		return h.BadRequest(ctx, ErrProcessingRequest, openapiCtx)
	}

	systemCtx := privacy.DecisionContext(requestCtx, privacy.Allow)

	if len(def.CredentialRegistrations) == 0 && installationRec.Status == enums.IntegrationStatusPending {
		if err := h.IntegrationsRuntime.DB().Integration.UpdateOneID(installationRec.ID).
			SetStatus(enums.IntegrationStatusConnected).
			Exec(systemCtx); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to mark credential-less installation connected")

			return h.BadRequest(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationRec.Status = enums.IntegrationStatusConnected
	}

	resp := IntegrationConfigResponse{
		Reply:                rout.Reply{Success: true},
		Provider:             def.ID,
		IntegrationID:        installationRec.ID,
		HealthStatus:         "ok",
		HealthSummary:        "Validation passed during configuration",
		InstallationMetadata: installationRec.InstallationMetadata.Attributes,
	}

	var primaryWebhookURL string
	var primaryWebhookSecret string

	for i, registration := range def.Webhooks {
		webhook, webhookErr := h.IntegrationsRuntime.EnsureWebhook(systemCtx, installationRec, registration.Name, "")
		if webhookErr != nil {
			logx.FromContext(requestCtx).Error().Err(webhookErr).Str("installation_id", installationRec.ID).Str("webhook", registration.Name).Msg("failed to ensure installation webhook")

			return h.BadRequest(ctx, ErrProcessingRequest, openapiCtx)
		}

		if i == 0 && webhook != nil {
			primaryWebhookURL = absoluteEndpointURL(ctx, lo.FromPtr(webhook.EndpointURL))
			primaryWebhookSecret = webhook.SecretToken
		}
	}

	if isNewInstallation {
		resp.WebhookEndpointURL = primaryWebhookURL
		resp.WebhookSecret = primaryWebhookSecret
	}

	return h.Success(ctx, resp)
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
