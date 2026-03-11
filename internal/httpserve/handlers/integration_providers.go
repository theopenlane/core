package handlers

import (
	"encoding/json"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party providers
func (h *Handler) ListIntegrationProviders(ctx echo.Context, _ *OpenAPIContext) error {
	reg := h.IntegrationRuntime.Registry()
	catalog := reg.ProviderMetadataCatalog()
	result := make([]types.IntegrationProviderMetadata, 0, len(catalog))

	for providerType, meta := range catalog {
		providerSpec, ok := reg.Config(providerType)
		if !ok {
			continue
		}

		result = append(result, buildIntegrationProviderMetadata(providerType, providerSpec, meta, reg))
	}

	resp := IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Schema:    keystore.Schema(),
		Providers: result,
	}

	return h.Success(ctx, resp)
}

// environmentCredentials returns the provider-specific operator config as raw JSON for API responses.
// Returns nil when the spec carries no provider config.
func environmentCredentials(providerSpec spec.ProviderSpec) json.RawMessage {
	if len(providerSpec.ProviderConfig) == 0 {
		return nil
	}

	return jsonx.CloneRawMessage(providerSpec.ProviderConfig)
}

// buildIntegrationProviderMetadata constructs provider metadata for API responses.
// It starts from the pre-built catalog metadata and augments it with environment credentials
// and operation descriptors derived from the live registry.
func buildIntegrationProviderMetadata(providerType types.ProviderType, providerSpec spec.ProviderSpec, meta types.IntegrationProviderMetadata, reg *registry.Registry) types.IntegrationProviderMetadata {
	entry := meta
	entry.EnvironmentCredentials = environmentCredentials(providerSpec)

	if reg != nil {
		if descriptors := reg.OperationDescriptors(providerType); len(descriptors) > 0 {
			entry.Operations = make([]types.OperationMetadata, 0, len(descriptors))
			for _, descriptor := range descriptors {
				entry.Operations = append(entry.Operations, types.OperationMetadata{
					Name:         string(descriptor.Name),
					Kind:         string(descriptor.Kind),
					Description:  descriptor.Description,
					Client:       string(descriptor.Client),
					ConfigSchema: jsonx.CloneRawMessage(descriptor.ConfigSchema),
				})
			}
		}
	}

	return entry
}
