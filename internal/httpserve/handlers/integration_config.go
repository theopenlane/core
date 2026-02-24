package handlers

import (
	"context"
	"errors"
	"maps"

	echo "github.com/theopenlane/echox"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/iam/auth"

	intauth "github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
	"github.com/theopenlane/core/internal/integrations/activation"
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
	if orgID == "" {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	if h.IntegrationRegistry == nil {
		return h.InternalServerError(ctx, errIntegrationRegistryNotConfigured, openapiCtx)
	}

	providerType := types.ProviderTypeFromString(providerKey)
	if providerType == types.ProviderUnknown {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	spec, ok := h.IntegrationRegistry.Config(providerType)
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

	if err := h.persistCredentialConfiguration(requestCtx, orgID, providerType, attrs); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Msg("error persisting credential configuration")

		switch {
		case errors.Is(err, activation.ErrHealthCheckFailed):
			return h.BadRequest(ctx, wrapIntegrationError("validate", err), openapiCtx)
		case errors.Is(err, activation.ErrOperationsRequired):
			return h.InternalServerError(ctx, errIntegrationOperationsNotConfigured, openapiCtx)
		default:
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	if record, err := h.IntegrationStore.EnsureIntegration(requestCtx, orgID, providerType); err == nil {
		if err := h.updateIntegrationProviderMetadata(requestCtx, record.ID, providerType); err != nil {
			logx.FromContext(requestCtx).Warn().Err(err).Str("provider", string(providerType)).Msg("failed to update integration provider metadata")
		}
	} else {
		logx.FromContext(requestCtx).Warn().Err(err).Str("provider", string(providerType)).Msg("failed to ensure integration record for metadata update")
	}

	out := openapi.IntegrationConfigResponse{
		Reply:    rout.Reply{Success: true},
		Provider: string(providerType),
	}

	return h.Success(ctx, out)
}

// persistCredentialConfiguration saves the provider credential configuration for the organization
func (h *Handler) persistCredentialConfiguration(ctx context.Context, orgID string, provider types.ProviderType, data map[string]any) error {
	if h.IntegrationActivation == nil {
		return errActivationNotConfigured
	}

	_, err := h.IntegrationActivation.Configure(ctx, activation.ConfigureRequest{
		OrgID:        orgID,
		Provider:     provider,
		ProviderData: maps.Clone(data),
		Validate:     true,
	})

	return err
}
