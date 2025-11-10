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
	"github.com/theopenlane/core/internal/keystore"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/utils/rout"
)

// ConfigureIntegrationProvider stores non-OAuth credentials/configuration for a provider.
func (h *Handler) ConfigureIntegrationProvider(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	payload, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, models.ExampleIntegrationConfigPayload, models.IntegrationConfigResponse{}, openapiCtx.Registry)
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

	rt, ok := h.IntegrationRegistry[providerKey]
	if !ok || rt == nil {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if len(rt.Spec.CredentialsSchema) == 0 {
		return h.BadRequest(ctx, rout.MissingField("credentialsSchema"), openapiCtx)
	}

	attrs := payload.Body.ToMap()
	if len(attrs) == 0 {
		return h.BadRequest(ctx, rout.MissingField("payload"), openapiCtx)
	}

	schemaLoader := gojsonschema.NewGoLoader(rt.Spec.CredentialsSchema)
	documentLoader := gojsonschema.NewGoLoader(attrs)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	if fieldErrs := jsonschemautil.FieldErrorsFromResult(result); len(fieldErrs) > 0 {
		// MKA fix this
		return h.BadRequest(ctx, nil, openapiCtx)
	}

	if err := h.persistCredentialConfiguration(requestCtx, orgID, providerKey, attrs); err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	out := models.IntegrationConfigResponse{
		Reply:    rout.Reply{Success: true},
		Provider: providerKey,
	}

	return h.Success(ctx, out)
}

func (h *Handler) persistCredentialConfiguration(ctx context.Context, orgID, provider string, data map[string]any) error {
	if h.IntegrationStore == nil {
		return errIntegrationStoreNotConfigured
	}

	alias := strings.TrimSpace(fmt.Sprint(data["alias"]))
	if alias == "" {
		if sa, ok := data["serviceAccountEmail"]; ok {
			alias = strings.TrimSpace(fmt.Sprint(sa))
		}
	}

	helper := keystore.NewHelper(provider, alias)
	integrationName := helper.Name()
	description := helper.Description()

	secrets := make([]keystore.SecretRecord, 0, len(data))
	for key, value := range data {
		if value == nil {
			continue
		}
		stringValue := stringifyValue(value)
		if stringValue == "" {
			continue
		}
		record := keystore.SecretRecord{
			Name:        helper.SecretName(key),
			DisplayName: helper.SecretDisplayName(integrationName, key),
			Description: helper.SecretDescription(key),
			Kind:        keystore.MetadataKind,
			Value:       stringValue,
		}
		secrets = append(secrets, record)
	}

	if len(secrets) == 0 {
		return rout.MissingField("req")
	}

	req := keystore.SaveRequest{
		OrgID:                  orgID,
		Provider:               provider,
		IntegrationName:        integrationName,
		IntegrationDescription: description,
		Secrets:                secrets,
	}

	_, err := h.IntegrationStore.UpsertIntegration(ctx, req)
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
