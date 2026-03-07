package handlers

import (
	"errors"

	echo "github.com/theopenlane/echox"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/iam/auth"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
	"github.com/theopenlane/core/internal/integrations/activation"
	intauth "github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// ConfigureIntegrationProvider stores non-OAuth credentials/configuration for a provider
func (h *Handler) ConfigureIntegrationProvider(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	payload, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleIntegrationConfigPayload, openapi.IntegrationConfigResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	requestCtx := ctx.Request().Context()

	providerKey := payload.Provider
	if providerKey == "" {
		return h.BadRequest(ctx, rout.MissingField("provider"), openapiCtx)
	}

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	orgID := caller.OrganizationID

	providerType := types.ProviderTypeFromString(providerKey)
	if providerType == types.ProviderUnknown {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	spec, ok := h.IntegrationRuntime.Registry().Config(providerType)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}
	if spec.Active == nil || !*spec.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if len(spec.CredentialsSchema) == 0 {
		return h.BadRequest(ctx, rout.MissingField("credentialsSchema"), openapiCtx)
	}

	attrs := payload.Body.ToMap()
	if len(attrs) == 0 {
		return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
	}

	if key, ok := attrs["serviceAccountKey"].(string); ok {
		attrs["serviceAccountKey"] = intauth.NormalizeServiceAccountKey(key)
	}

	schemaLoader := gojsonschema.NewGoLoader(spec.CredentialsSchema)
	documentLoader := gojsonschema.NewGoLoader(attrs)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	if fieldErrs := jsonschemautil.FieldErrorsFromResult(result); len(fieldErrs) > 0 {
		return h.BadRequest(ctx, fieldErrs[0], openapiCtx)
	}

	if _, err := h.IntegrationRuntime.Activation().Configure(requestCtx, activation.ConfigureRequest{
		OrgID:        orgID,
		Provider:     providerType,
		ProviderData: attrs,
		Validate:     true,
	}); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Msg("error persisting credential configuration")

		if errors.Is(err, activation.ErrHealthCheckFailed) {
			return h.BadRequest(ctx, wrapIntegrationError("validate", ErrProviderHealthCheckFailed), openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if record, err := h.IntegrationRuntime.Store().EnsureIntegration(requestCtx, orgID, providerType); err == nil {
		if err := h.updateIntegrationProviderMetadata(requestCtx, record.ID, providerType); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("provider", string(providerType)).Msg("failed to update integration provider metadata")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	out := openapi.IntegrationConfigResponse{
		Reply:    rout.Reply{Success: true},
		Provider: string(providerType),
	}

	return h.Success(ctx, out)
}
