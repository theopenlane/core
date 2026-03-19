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

	credentialInput := payload.Body.RawMessage()
	credentialProvided := !payload.Body.IsNullOrEmptyObject()

	if !credentialProvided {
		if !userInputProvided || payload.InstallationID == "" {
			return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
		}
	}

	var providerData json.RawMessage
	if credentialProvided {
		if def.Credentials == nil || len(def.Credentials.Schema) == 0 {
			return h.BadRequest(ctx, rout.MissingField("credentialsSchema"), openapiCtx)
		}

		credentialValidation, err := jsonx.ValidateSchema(def.Credentials.Schema, credentialInput)
		if err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		if !credentialValidation.Valid() {
			return h.InvalidInput(ctx, ErrInvalidInput, openapiCtx)
		}

		providerData = jsonx.CloneRawMessage(credentialInput)
	}

	installationRec, _, err := h.resolveOrCreateDefinitionIntegration(requestCtx, caller.OrganizationID, payload.InstallationID, def)
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	logger = logger.With().Str("installation_id", installationRec.ID).Logger()

	if userInputProvided {
		if err := h.persistInstallationUserInput(requestCtx, installationRec, payload.UserInput.RawMessage()); err != nil {
			logger.Error().Err(err).Msg("failed to persist user input")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	if credentialProvided {
		credential := types.CredentialSet{ProviderData: providerData}

		if err := h.finalizeIntegrationConnection(ctx, openapiCtx, installationRec, def, credential, nil); err != nil {
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
