package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
	"github.com/theopenlane/core/internal/integrations/types"
	credentialmodels "github.com/theopenlane/core/pkg/models"
	openapi "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/utils/rout"
)

// ConfigureIntegrationProvider stores non-OAuth credentials/configuration for a provider.
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

	orgID, err := auth.GetOrganizationIDFromContext(requestCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
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

	if len(spec.CredentialsSchema) == 0 {
		return h.BadRequest(ctx, rout.MissingField("credentialsSchema"), openapiCtx)
	}

	attrs := payload.Body.ToMap()
	if len(attrs) == 0 {
		return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
	}

	schemaLoader := gojsonschema.NewGoLoader(spec.CredentialsSchema)
	documentLoader := gojsonschema.NewGoLoader(attrs)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	if fieldErrs := jsonschemautil.FieldErrorsFromResult(result); len(fieldErrs) > 0 {
		// MKA fix this
		return h.BadRequest(ctx, nil, openapiCtx)
	}

	if err := h.persistCredentialConfiguration(requestCtx, orgID, providerType, attrs); err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	out := openapi.IntegrationConfigResponse{
		Reply:    rout.Reply{Success: true},
		Provider: string(providerType),
	}

	return h.Success(ctx, out)
}

func (h *Handler) persistCredentialConfiguration(ctx context.Context, orgID string, provider types.ProviderType, data map[string]any) error {
	if h.IntegrationStore == nil {
		return errIntegrationStoreNotConfigured
	}

	payload, err := types.NewCredentialBuilder(provider).
		With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(credentialmodels.CredentialSet{
				ProviderData: cloneProviderData(data),
			}),
		).
		Build()
	if err != nil {
		return err
	}

	_, err = h.IntegrationStore.SaveCredential(ctx, orgID, payload)
	return err
}

func stringifyValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func cloneProviderData(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(data))
	for key, value := range data {
		cloned[key] = stringifyValue(value)
	}
	return cloned
}
