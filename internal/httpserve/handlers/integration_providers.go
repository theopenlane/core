package handlers

import (
	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party integration definitions
func (h *Handler) ListIntegrationProviders(ctx echo.Context, _ *OpenAPIContext) error {
	reg := h.IntegrationsRuntime.Registry()
	specs := reg.Catalog()
	entries := make([]DefinitionCatalogEntry, 0, len(specs))

	for _, spec := range specs {
		def, ok := reg.Definition(spec.ID)
		if !ok {
			continue
		}

		entries = append(entries, buildDefinitionCatalogEntry(def))
	}

	return h.Success(ctx, IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Providers: entries,
	})
}

// buildDefinitionCatalogEntry constructs one catalog entry from an integration definition
func buildDefinitionCatalogEntry(def types.Definition) DefinitionCatalogEntry {
	spec := def.DefinitionSpec
	entry := DefinitionCatalogEntry{
		ID:          spec.ID,
		Slug:        spec.Slug,
		Version:     spec.Version,
		Family:      spec.Family,
		DisplayName: spec.DisplayName,
		Description: spec.Description,
		Category:    spec.Category,
		DocsURL:     spec.DocsURL,
		LogoURL:     spec.LogoURL,
		Tags:        spec.Tags,
		Labels:      spec.Labels,
		Active:      spec.Active,
		Visible:     spec.Visible,
		HasAuth:     def.Auth != nil,
	}

	if def.Auth != nil && (def.Auth.StartPath != "" || def.Auth.CallbackPath != "" || def.Auth.OAuth != nil) {
		entry.Auth = def.Auth
	}

	if def.Credentials != nil {
		entry.CredentialSchema = jsonx.CloneRawMessage(def.Credentials.Schema)
	}

	if def.OperatorConfig != nil {
		entry.OperatorConfig = jsonx.CloneRawMessage(def.OperatorConfig.Schema)
	}

	if def.UserInput != nil {
		entry.UserInputSchema = jsonx.CloneRawMessage(def.UserInput.Schema)
	}

	if len(def.Operations) > 0 {
		entry.Operations = lo.Map(def.Operations, func(op types.OperationRegistration, _ int) DefinitionOperationEntry {
			return DefinitionOperationEntry{
				Name:         op.Name,
				Description:  op.Description,
				ConfigSchema: jsonx.CloneRawMessage(op.ConfigSchema),
			}
		})
	}

	return entry
}
