package handlers

import (
	"encoding/json"
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// ConfigureIntegrationProvider stores non-OAuth credentials for a provider definition.
// When installation_id is provided the credentials on that installation are updated.
// When omitted a new installation is created and its ID is returned in the response.
func (h *Handler) ConfigureIntegrationProvider(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	payload, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationConfigPayload, IntegrationConfigResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, err := h.resolveActiveDefinition(ctx, payload.DefinitionID, openapiCtx)
	if err != nil {
		return err
	}

	logger := logx.FromContext(requestCtx).With().Str("definition_id", def.ID).Logger()

	userInputProvided := len(payload.UserInput) > 0

	if err := validateDefinitionUserInput(def, payload.UserInput); err != nil {
		if errors.Is(err, ErrInvalidInput) {
			return h.InvalidInput(ctx, err, openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	credentialInput := payload.Body
	credentialProvided := !isNullOrEmptyJSON(payload.Body)

	if !credentialProvided {
		if !userInputProvided || payload.InstallationID == "" {
			return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
		}
	}

	var providerData json.RawMessage
	var credentialRegistration types.CredentialRegistration
	if credentialProvided {
		credentialRegistration, err = resolveCredentialRegistration(def, payload.CredentialRef)
		if err != nil {
			return h.BadRequest(ctx, err, openapiCtx)
		}

		if len(credentialRegistration.Schema) == 0 {
			// A missing credential schema is a definition registration error, not a user input error.
			logger.Error().Str("credential_ref", payload.CredentialRef.String()).Msg("credential registration has no schema")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		credentialValidation, err := jsonx.ValidateSchema(credentialRegistration.Schema, credentialInput)
		if err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		if !credentialValidation.Valid() {
			return h.InvalidInput(ctx, ErrInvalidInput, openapiCtx)
		}

		providerData = jsonx.CloneRawMessage(credentialInput)
	}

	installationRec, _, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, payload.InstallationID, def)
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	logger = logger.With().Str("installation_id", installationRec.ID).Logger()

	if userInputProvided {
		if err := h.persistInstallationUserInput(requestCtx, installationRec, payload.UserInput); err != nil {
			logger.Error().Err(err).Msg("failed to persist user input")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	if credentialProvided {
		credential := types.CredentialSet{ProviderData: providerData}

		if len(def.CredentialRegistrations) > 1 {
			// Multi-credential definitions gather credentials one slot at a time.
			// Save each slot as it arrives; attempt a health check after each save
			// but do not fail — the integration remains Pending until all required
			// slots are filled and the health check passes with full credentials.
			if err := h.IntegrationsRuntime.SaveCredential(requestCtx, installationRec, credentialRegistration.Ref, credential); err != nil {
				logger.Error().Err(err).Str("credential_ref", credentialRegistration.Ref.String()).Msg("failed to save credential")
				return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
			}

			healthOperation, healthErr := h.IntegrationsRuntime.Registry().Operation(def.ID, types.HealthDefaultOperation)
			if healthErr == nil {
				if _, healthErr = h.IntegrationsRuntime.ExecuteOperation(requestCtx, installationRec, healthOperation, nil, nil); healthErr == nil {
					// MarkConnected failure is non-fatal here: the credential was saved and
					// the health check passed; status can be reconciled on next credential update.
					_ = h.IntegrationsRuntime.MarkConnected(requestCtx, installationRec)
				} else {
					logger.Info().Err(healthErr).Str("credential_ref", credentialRegistration.Ref.String()).Msg("credential slot saved; integration remains pending until all required credentials are configured")
				}
			}
		} else if err := h.finalizeIntegrationConnection(ctx, openapiCtx, installationRec, def, credentialRegistration, credential, nil); err != nil {
			return err
		}

	}

	if err := h.IntegrationsRuntime.SyncWebhooks(requestCtx, installationRec, ""); err != nil {
		logger.Error().Err(err).Msg("failed to sync webhooks")
	}

	return h.Success(ctx, IntegrationConfigResponse{
		Reply:          rout.Reply{Success: true},
		Provider:       def.Slug,
		InstallationID: installationRec.ID,
	})
}
