package handlers

import (
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party providers.
func (h *Handler) ListIntegrationProviders(ctx echo.Context, _ *OpenAPIContext) error {
	if h.IntegrationRegistry == nil {
		return h.InternalServerError(ctx, errIntegrationRegistryNotConfigured, nil)
	}

	catalog := h.IntegrationRegistry.ProviderMetadataCatalog()
	result := make([]openapi.IntegrationProviderMetadata, 0, len(catalog))
	for providerType, meta := range catalog {
		spec, ok := h.IntegrationRegistry.Config(providerType)
		if !ok {
			continue
		}

		entry := openapi.IntegrationProviderMetadata{
			Name:             defaultProviderName(providerType, spec.Name),
			DisplayName:      meta.DisplayName,
			Category:         meta.Category,
			AuthType:         keystore.AuthType(meta.Auth),
			Active:           spec.Active,
			LogoURL:          meta.LogoURL,
			DocsURL:          meta.DocsURL,
			Persistence:      spec.Persistence,
			WorkloadIdentity: spec.WorkloadIdentity,
			GitHubApp:        spec.GitHubApp,
			Labels:           spec.Labels,
			CredentialsSchema: func() map[string]any {
				if len(meta.Schema) > 0 {
					return meta.Schema
				}
				return spec.CredentialsSchema
			}(),
		}

		if spec.OAuth != nil && (spec.AuthType == types.AuthKindOAuth2 || spec.AuthType == types.AuthKindOIDC) {
			entry.OAuth = &openapi.IntegrationOAuthMetadata{
				AuthURL:     spec.OAuth.AuthURL,
				TokenURL:    spec.OAuth.TokenURL,
				RedirectURI: spec.OAuth.RedirectURI,
				Scopes:      append([]string{}, spec.OAuth.Scopes...),
				UsePKCE:     spec.OAuth.UsePKCE,
				AuthParams:  spec.OAuth.AuthParams,
				TokenParams: spec.OAuth.TokenParams,
			}
		}

		result = append(result, entry)
	}

	resp := openapi.IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Schema:    keystore.Schema(),
		Providers: result,
	}

	return h.Success(ctx, resp)
}

func defaultProviderName(provider types.ProviderType, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	if provider == types.ProviderUnknown {
		return ""
	}
	return strings.ToLower(string(provider))
}
