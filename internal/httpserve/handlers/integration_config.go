package handlers

import (
	"encoding/json"
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
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

	def, ok := h.IntegrationsRuntime.Registry().Definition(payload.DefinitionID)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	userInputProvided := len(payload.UserInput) > 0
	userInput := json.RawMessage(payload.UserInput)

	if err := validateDefinitionUserInput(def, payload.UserInput); err != nil {
		if errors.Is(err, ErrInvalidInput) {
			return h.InvalidInput(ctx, err, openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	attrs := payload.Body.ToMap()
	credentialProvided := len(attrs) > 0

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

		credentialValidation, err := jsonx.ValidateSchema(def.Credentials.Schema, attrs)
		if err != nil {
			if !credentialValidation.Valid() {
				return h.InvalidInput(ctx, ErrInvalidInput, openapiCtx)
			}
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		providerData, err = json.Marshal(attrs)
		if err != nil {
			return h.InternalServerError(ctx, err, openapiCtx)
		}
	}

	installationRec, _, err := h.resolveOrCreateDefinitionIntegration(requestCtx, caller.OrganizationID, payload.InstallationID, def)
	if err != nil {
		if errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch) {
			return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
		}
		if errors.Is(err, integrationsruntime.ErrInstallationNotFound) {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", payload.InstallationID).Msg("installation not found")

			return h.NotFound(ctx, ErrIntegrationNotFound, openapiCtx)
		}

		logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", def.ID).Msg("failed to resolve installation")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	installationID := installationRec.ID
	if userInputProvided {
		config := installationRec.Config
		config.ClientConfig = userInput

		update := h.DBClient.Integration.UpdateOneID(installationRec.ID).SetConfig(config)

		var inputMap map[string]any
		if err := json.Unmarshal(userInput, &inputMap); err == nil {
			if name, ok := inputMap["name"].(string); ok && name != "" {
				update.SetName(name)
				installationRec.Name = name
			}
		}

		if err := update.Exec(requestCtx); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to persist integration user input")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationRec.Config = config
	}

	if credentialProvided {
		credential := types.CredentialSet{ProviderData: providerData}

		healthOperation, err := h.IntegrationsRuntime.Registry().Operation(def.ID, "health.default")
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", def.ID).Msg("health operation not registered")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		if _, err := h.IntegrationsRuntime.ExecuteOperation(requestCtx, installationRec, healthOperation, credential, nil); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationID).Msg("provider health check failed")
			return h.BadRequest(ctx, wrapIntegrationError("validate", ErrProviderHealthCheckFailed), openapiCtx)
		}

		if err := h.IntegrationsRuntime.SaveCredential(requestCtx, installationRec, credential); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to save credential")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		if def.Installation != nil {
			metadata, ok, err := def.Installation.Resolve(requestCtx, types.InstallationRequest{
				Installation: installationRec,
				Credential:   credential,
				Config:       installationRec.Config,
			})
			if err != nil {
				logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to resolve installation metadata")
				return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
			}

			if ok {
				if err := integrationsruntime.SaveInstallationMetadata(requestCtx, installationRec, metadata); err != nil {
					logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to save installation metadata")
					return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
				}
			}
		}

		if err := h.DBClient.Integration.UpdateOneID(installationRec.ID).
			SetStatus(enums.IntegrationStatusConnected).
			Exec(requestCtx); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to update integration status")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationRec.Status = enums.IntegrationStatusConnected
	}

	if err := h.IntegrationsRuntime.SyncWebhooks(requestCtx, installationRec, ""); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to sync integration webhooks")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.Success(ctx, IntegrationConfigResponse{
		Reply:          rout.Reply{Success: true},
		Provider:       def.Slug,
		InstallationID: installationID,
	})
}
