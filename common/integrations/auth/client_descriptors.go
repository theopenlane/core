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

// OAuthClientBuilder returns a ClientBuilderFunc that extracts an OAuth token and creates an AuthenticatedClient.
func OAuthClientBuilder(headers map[string]string) types.ClientBuilderFunc {
	return func(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
		token, err := OAuthTokenFromPayload(payload)
		if err != nil {
			return nil, err
		}

		return NewAuthenticatedClient(token, headers), nil
	}
}

// APITokenClientBuilder returns a ClientBuilderFunc that extracts an API token and creates an AuthenticatedClient.
func APITokenClientBuilder(headers map[string]string) types.ClientBuilderFunc {
	return func(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
		token, err := APITokenFromPayload(payload)
		if err != nil {
			return nil, err
		}

		return NewAuthenticatedClient(token, headers), nil
	}
}
