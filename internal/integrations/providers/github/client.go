package github

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientGitHubAPI identifies the GitHub REST API client.
	ClientGitHubAPI types.ClientName = "api"
)

// githubClientDescriptors returns the client descriptors for the GitHub provider.
func githubClientDescriptors(provider types.ProviderType) []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(provider, ClientGitHubAPI, "GitHub REST API client", buildGitHubClient(provider))
}

// buildGitHubClient constructs an authenticated GitHub REST API client.
func buildGitHubClient(provider types.ProviderType) types.ClientBuilderFunc {
	return func(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
		token, err := auth.OAuthTokenFromPayload(payload, string(provider))
		if err != nil {
			return nil, err
		}

		headers := map[string]string{
			"Accept":               "application/vnd.github+json",
			"X-GitHub-Api-Version": githubAPIVersion,
		}

		return auth.NewAuthenticatedClient(token, headers), nil
	}
}
