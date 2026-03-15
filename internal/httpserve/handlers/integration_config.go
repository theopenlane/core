package handlers

import (
	"encoding/json"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
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

	requestCtx := ctx.Request().Context()

	if strings.TrimSpace(payload.Provider) == "" {
		return h.BadRequest(ctx, rout.MissingField("provider"), openapiCtx)
	}

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.resolveIntegrationDefinition(payload.Provider)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Spec.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if def.Credentials == nil || len(def.Credentials.Schema) == 0 {
		return h.BadRequest(ctx, rout.MissingField("credentialsSchema"), openapiCtx)
	}

	attrs := payload.Body.ToMap()
	if len(attrs) == 0 {
		return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
	}

	if key, ok := attrs["serviceAccountKey"].(string); ok {
		trimmed := strings.TrimSpace(key)
		var decoded string
		if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
			trimmed = strings.TrimSpace(decoded)
		}
		attrs["serviceAccountKey"] = trimmed
	}

	providerData, err := json.Marshal(attrs)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	userInputProvided := len(payload.UserInput) > 0
	normalizedUserInput := json.RawMessage(nil)
	if userInputProvided {
		if def.UserInput == nil || len(def.UserInput.Schema) == 0 {
			return h.BadRequest(ctx, rout.MissingField("userInputSchema"), openapiCtx)
		}

		normalizedUserInput, err = normalizeUserInput(requestCtx, def.UserInput, json.RawMessage(payload.UserInput))
		if err != nil {
			return h.BadRequest(ctx, err, openapiCtx)
		}
	}

	var (
		installationID  string
		installationRec *ent.Integration
		created         bool
	)

	switch {
	case payload.InstallationID != "":
		// Caller is updating credentials on an existing installation.
		rec, err := h.loadOwnedIntegration(requestCtx, caller.OrganizationID, payload.InstallationID)
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", payload.InstallationID).Msg("installation not found")
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}
		if err := validateInstallationDefinition(rec, def); err != nil {
			return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
		}

		installationID = rec.ID
		installationRec = rec

	default:
		// No installation ID provided — create a new installation.
		name := def.Spec.DisplayName
		if name == "" {
			name = def.Spec.Slug
		}

		rec, err := h.DBClient.Integration.Create().
			SetOwnerID(caller.OrganizationID).
			SetName(name).
			SetDefinitionID(string(def.Spec.ID)).
			SetDefinitionVersion(def.Spec.Version).
			SetDefinitionSlug(def.Spec.Slug).
			SetFamily(def.Spec.Family).
			SetStatus(enums.IntegrationStatusPending).
			Save(requestCtx)
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", string(def.Spec.ID)).Msg("failed to create installation")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationID = rec.ID
		installationRec = rec
		created = true
	}

	previousConfig := installationRec.Config
	if userInputProvided {
		nextConfig := installationRec.Config
		nextConfig.ClientConfig = normalizedUserInput
		installationRec.Config = nextConfig
	}

	credential, err := normalizeCredential(requestCtx, def.Credentials, installationRec, providerData)
	if err != nil {
		if created {
			_ = h.DBClient.Integration.DeleteOneID(installationRec.ID).Exec(requestCtx)
		}

		return h.BadRequest(ctx, err, openapiCtx)
	}

	if _, err := h.IntegrationsRuntime.Executor().ExecuteOperation(requestCtx, installationRec, "health.default", credential, nil); err != nil {
		if created {
			_ = h.DBClient.Integration.DeleteOneID(installationRec.ID).Exec(requestCtx)
		}

		logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationID).Msg("provider health check failed")
		return h.BadRequest(ctx, wrapIntegrationError("validate", ErrProviderHealthCheckFailed), openapiCtx)
	}

	if userInputProvided {
		if err := h.DBClient.Integration.UpdateOneID(installationRec.ID).SetConfig(installationRec.Config).Exec(requestCtx); err != nil {
			if created {
				_ = h.DBClient.Integration.DeleteOneID(installationRec.ID).Exec(requestCtx)
			}

			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to persist integration user input")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	if err := h.IntegrationsRuntime.CredentialStore().SaveCredential(requestCtx, installationRec, credential); err != nil {
		if created {
			_ = h.DBClient.Integration.DeleteOneID(installationRec.ID).Exec(requestCtx)
		} else if userInputProvided {
			_ = h.DBClient.Integration.UpdateOneID(installationRec.ID).SetConfig(previousConfig).Exec(requestCtx)
		}

		logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to save credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.Success(ctx, IntegrationConfigResponse{
		Reply:          rout.Reply{Success: true},
		Provider:       def.Spec.Slug,
		InstallationID: installationID,
	})
}
