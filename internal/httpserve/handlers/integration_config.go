package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
	"github.com/theopenlane/shared/integrations/types"
	"github.com/theopenlane/shared/logx"
	credentialmodels "github.com/theopenlane/shared/models"
	openapi "github.com/theopenlane/shared/openapi"
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
	if !spec.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if len(spec.CredentialsSchema) == 0 {
		return h.BadRequest(ctx, rout.MissingField("credentialsSchema"), openapiCtx)
	}

	attrs := payload.Body.ToMap()
	if len(attrs) == 0 {
		return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
	}

	if key, ok := attrs["serviceAccountKey"].(string); ok && strings.TrimSpace(key) != "" {
		attrs["serviceAccountKey"] = normalizeServiceAccountKey(key)
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

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.runIntegrationHealthCheck(requestCtx, orgID, providerType); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Msg("error running integration health check")

		switch {
		case errors.Is(err, errIntegrationOperationsNotConfigured),
			errors.Is(err, errIntegrationRegistryNotConfigured):
			return h.InternalServerError(ctx, err, openapiCtx)
		default:
			return h.BadRequest(ctx, wrapIntegrationError("validate", err), openapiCtx)
		}
	}

	out := openapi.IntegrationConfigResponse{
		Reply:    rout.Reply{Success: true},
		Provider: string(providerType),
	}

	return h.Success(ctx, out)
}

// persistCredentialConfiguration saves the provider credential configuration for the organization
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

func cloneProviderData(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(data))
	for key, value := range data {
		if key == "serviceAccountKey" {
			cloned[key] = normalizeServiceAccountKey(stringifyValue(value))
			continue
		}
		cloned[key] = cloneProviderValue(value)
	}

	return cloned
}

func cloneProviderValue(value any) any {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			out[i] = strings.TrimSpace(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneProviderValue(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = cloneProviderValue(item)
		}
		return out
	default:
		return v
	}
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

func normalizeServiceAccountKey(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return strings.TrimSpace(decoded)
	}

	return trimmed
}

const defaultHealthOperation types.OperationName = "health.default"

// runIntegrationHealthCheck performs a health check operation for the given provider if supported
func (h *Handler) runIntegrationHealthCheck(ctx context.Context, orgID string, provider types.ProviderType) error {
	if !h.providerHasHealthOperation(provider) {
		return nil
	}

	result, err := h.IntegrationOperations.Run(ctx, types.OperationRequest{
		OrgID:    orgID,
		Provider: provider,
		Name:     defaultHealthOperation,
		Force:    true,
	})
	if err != nil {
		return err
	}

	if result.Status != types.OperationStatusOK {
		summary := strings.TrimSpace(result.Summary)
		if summary == "" {
			return ErrProviderHealthCheckFailed
		}
		return fmt.Errorf("%w: %s", ErrProviderHealthCheckFailed, summary)
	}

	return nil
}

// providerHasHealthOperation checks if the provider has a health check operation defined
func (h *Handler) providerHasHealthOperation(provider types.ProviderType) bool {
	for _, descriptor := range h.IntegrationRegistry.OperationDescriptors(provider) {
		if descriptor.Name == defaultHealthOperation {
			return true
		}
	}

	return false
}
