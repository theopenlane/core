package handlers

import (
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party providers.
func (h *Handler) ListIntegrationProviders(ctx echo.Context, _ *OpenAPIContext) error {
	result := make([]openapi.IntegrationProviderMetadata, 0, len(h.IntegrationRegistry))
	for name, rt := range h.IntegrationRegistry {
		sp := rt.Spec
		meta := openapi.IntegrationProviderMetadata{
			Name:             sp.Name,
			DisplayName:      sp.DisplayName,
			Category:         sp.Category,
			AuthType:         sp.AuthType,
			Active:           sp.Active,
			LogoURL:          sp.LogoURL,
			DocsURL:          sp.DocsURL,
			Persistence:      sp.Persistence,
			WorkloadIdentity: sp.WorkloadIdentity,
			GitHubApp:        sp.GitHubApp,
			Labels:           sp.Labels,
		}

		if h.IntegrationCredentialSchemas != nil {
			if schema, ok := h.IntegrationCredentialSchemas[strings.ToLower(meta.Name)]; ok {
				meta.CredentialsSchema = schema
			}
		}
		if meta.CredentialsSchema == nil && len(sp.CredentialsSchema) > 0 {
			meta.CredentialsSchema = sp.CredentialsSchema
		}

		if sp.OAuth != nil && (sp.AuthType == keystore.AuthTypeOAuth2 || sp.AuthType == keystore.AuthTypeOIDC) {
			meta.OAuth = &openapi.IntegrationOAuthMetadata{
				AuthURL:     sp.OAuth.AuthURL,
				TokenURL:    sp.OAuth.TokenURL,
				RedirectURI: sp.OAuth.RedirectURI,
				Scopes:      append([]string{}, sp.OAuth.Scopes...),
				UsePKCE:     sp.OAuth.UsePKCE,
				AuthParams:  sp.OAuth.AuthParams,
				TokenParams: sp.OAuth.TokenParams,
			}
		}

		// Ensure provider name is the registry key if omitted.
		if meta.Name == "" {
			meta.Name = strings.ToLower(name)
		}

		result = append(result, meta)
	}

	resp := openapi.IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Schema:    keystore.Schema(),
		Providers: result,
	}

	return h.Success(ctx, resp)
}
