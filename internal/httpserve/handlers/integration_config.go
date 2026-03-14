package handlers

import (
	"encoding/json"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
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

	requestCtx := ctx.Request().Context()

	definitionID := types.DefinitionID(payload.Provider)
	if string(definitionID) == "" {
		return h.BadRequest(ctx, rout.MissingField("provider"), openapiCtx)
	}

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(definitionID)
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

	result, err := jsonx.ValidateSchema(def.Credentials.Schema, attrs)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	if fieldErrs := jsonschemautil.FieldErrorsFromResult(result); len(fieldErrs) > 0 {
		return h.BadRequest(ctx, fieldErrs[0], openapiCtx)
	}

	providerData, err := json.Marshal(attrs)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	credential := types.CredentialSet{ProviderData: providerData}

	var (
		installationID  string
		installationRec *ent.Integration
	)

	switch {
	case payload.InstallationID != "":
		// Caller is updating credentials on an existing installation.
		rec, err := h.DBClient.Integration.Get(requestCtx, payload.InstallationID)
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", payload.InstallationID).Msg("installation not found")
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		if err := h.IntegrationsRuntime.CredentialStore().SaveCredential(requestCtx, rec, credential); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", payload.InstallationID).Msg("failed to save credential")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
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
			logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", string(definitionID)).Msg("failed to create installation")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		if err := h.IntegrationsRuntime.CredentialStore().SaveCredential(requestCtx, rec, credential); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", rec.ID).Msg("failed to save credential")
			// Best-effort cleanup — installation was created but credential save failed.
			_ = h.DBClient.Integration.DeleteOneID(rec.ID).Exec(requestCtx)
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationID = rec.ID
		installationRec = rec
	}

	if _, err := h.IntegrationsRuntime.Executor().ExecuteOperation(requestCtx, installationRec, "health.default", credential, nil); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationID).Msg("provider health check failed")
		return h.BadRequest(ctx, wrapIntegrationError("validate", ErrProviderHealthCheckFailed), openapiCtx)
	}

	return h.Success(ctx, IntegrationConfigResponse{
		Reply:          rout.Reply{Success: true},
		Provider:       payload.Provider,
		InstallationID: installationID,
	})
}
