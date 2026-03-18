package auth

import (
	"context"

	"github.com/theopenlane/core/common/integrations/types"
)

// DefaultClientDescriptor returns a descriptor with a default object config schema.
func DefaultClientDescriptor(provider types.ProviderType, name types.ClientName, description string, build types.ClientBuilderFunc) types.ClientDescriptor {
	return types.ClientDescriptor{
		Provider:     provider,
		Name:         name,
		Description:  description,
		Build:        build,
		ConfigSchema: map[string]any{"type": "object"},
	}
}

// DefaultClientDescriptors returns a single-descriptor slice with a default object config schema.
func DefaultClientDescriptors(provider types.ProviderType, name types.ClientName, description string, build types.ClientBuilderFunc) []types.ClientDescriptor {
	return []types.ClientDescriptor{
		DefaultClientDescriptor(provider, name, description, build),
	}
}

// TokenClientBuilder returns a ClientBuilderFunc that extracts a token and creates an AuthenticatedClient.
func TokenClientBuilder(extract TokenExtractor, headers map[string]string) types.ClientBuilderFunc {
	return func(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
		token, err := extract(payload)

		return NewAuthenticatedClient(token, headers), err
	}
}
