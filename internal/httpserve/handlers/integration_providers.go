package handlers

import (
	"maps"
	"sort"
	"strings"

	"github.com/samber/lo"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party providers
func (h *Handler) ListIntegrationProviders(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if h.IntegrationRegistry == nil {
		return h.InternalServerError(ctx, errIntegrationRegistryNotConfigured, openapiCtx)
	}

	catalog := h.IntegrationRegistry.ProviderMetadataCatalog()
	result := make([]openapi.IntegrationProviderMetadata, 0, len(catalog))

	for providerType, meta := range catalog {
		spec, ok := h.IntegrationRegistry.Config(providerType)
		if !ok {
			continue
		}

		entry := buildIntegrationProviderMetadata(providerType, spec, meta, h.IntegrationRegistry)
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

func providerTags(spec config.ProviderSpec) []string {
	if len(spec.Tags) > 0 {
		return append([]string{}, spec.Tags...)
	}

	keys := lo.Keys(spec.Labels)
	sort.Strings(keys)

	tags := lo.FilterMap(keys, func(key string, _ int) (string, bool) {
		value := strings.TrimSpace(spec.Labels[key])
		return value, value != ""
	})
	if category := strings.TrimSpace(spec.Category); category != "" {
		tags = append([]string{category}, tags...)
	}
	tags = lo.Uniq(tags)

	if len(tags) == 0 {
		return nil
	}

	return tags
}

// buildIntegrationProviderMetadata constructs provider metadata for API responses
func buildIntegrationProviderMetadata(providerType types.ProviderType, spec config.ProviderSpec, meta types.ProviderConfig, registry ProviderRegistry) openapi.IntegrationProviderMetadata {
	entry := openapi.IntegrationProviderMetadata{
		Name:                   defaultProviderName(providerType, spec.Name),
		DisplayName:            meta.DisplayName,
		Category:               meta.Category,
		Description:            meta.Description,
		AuthType:               keystore.AuthType(meta.Auth),
		AuthStartPath:          spec.AuthStartPath,
		AuthCallbackPath:       spec.AuthCallbackPath,
		Active:                 spec.Active,
		Visible:                spec.Visible,
		Tags:                   providerTags(spec),
		LogoURL:                meta.LogoURL,
		DocsURL:                meta.DocsURL,
		Persistence:            spec.Persistence,
		GoogleWorkloadIdentity: spec.GoogleWorkloadIdentity,
		GitHubApp:              spec.GitHubApp,
		Labels:                 spec.Labels,
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

	if registry != nil {
		if descriptors := registry.OperationDescriptors(providerType); len(descriptors) > 0 {
			entry.Operations = make([]openapi.IntegrationOperationMetadata, 0, len(descriptors))
			for _, descriptor := range descriptors {
				entry.Operations = append(entry.Operations, openapi.IntegrationOperationMetadata{
					Name:         string(descriptor.Name),
					Kind:         string(descriptor.Kind),
					Description:  descriptor.Description,
					Client:       string(descriptor.Client),
					ConfigSchema: maps.Clone(descriptor.ConfigSchema),
				})
			}
		}
	}

	return entry
}
