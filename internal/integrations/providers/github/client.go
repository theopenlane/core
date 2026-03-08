package github

import (
	"context"
	"encoding/json"
	"strings"

	gh "github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientGitHubAPI identifies the GitHub REST API client.
	ClientGitHubAPI types.ClientName = "api"
	// ClientGitHubGraphQL identifies the GitHub GraphQL API client.
	ClientGitHubGraphQL types.ClientName = "graphql"
)

var githubClientHeaders = map[string]string{
	"Accept":               "application/vnd.github+json",
	"X-GitHub-Api-Version": githubAPIVersion,
}

// githubClientDescriptors returns the client descriptors for the GitHub provider.
func githubClientDescriptors(provider types.ProviderType) []types.ClientDescriptor {
	return githubClientDescriptorsWithBaseURL(provider, githubAPIBaseURL)
}

// githubClientDescriptorsWithBaseURL returns client descriptors for a provider and optional custom API base URL.
func githubClientDescriptorsWithBaseURL(provider types.ProviderType, baseURL string) []types.ClientDescriptor {
	descriptors := providerkit.DefaultClientDescriptors(provider, ClientGitHubAPI, "GitHub REST API client", buildGitHubAPIClient(baseURL))
	descriptors = append(descriptors, providerkit.DefaultClientDescriptor(
		provider,
		ClientGitHubGraphQL,
		"GitHub GraphQL API client",
		buildGitHubGraphQLClient(),
	))

	return descriptors
}

// buildGitHubAPIClient returns a pooled client builder for the GitHub REST API.
func buildGitHubAPIClient(baseURL string) types.ClientBuilderFunc {
	return func(ctx context.Context, payload models.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
		token, err := auth.OAuthTokenFromPayload(payload)
		if err != nil {
			return types.EmptyClientInstance(), err
		}

		client, err := newGitHubAPIClient(ctx, token, baseURL)
		if err != nil {
			return types.EmptyClientInstance(), err
		}

		return types.NewClientInstance(client), nil
	}
}

// newGitHubAPIClient initializes an authenticated GitHub REST client.
func newGitHubAPIClient(ctx context.Context, token string, baseURL string) (*gh.Client, error) {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, source)
	httpClient.Transport = githubHeaderTransport{
		next:    httpClient.Transport,
		headers: githubClientHeaders,
	}

	client := gh.NewClient(httpClient)
	normalizedBaseURL := strings.TrimRight(baseURL, "/")
	if normalizedBaseURL == "" || normalizedBaseURL == githubAPIBaseURL {
		return client, nil
	}

	uploadURL := strings.TrimSuffix(normalizedBaseURL, "/api/v3")
	if uploadURL == "" {
		uploadURL = normalizedBaseURL
	}

	return client.WithEnterpriseURLs(normalizedBaseURL, uploadURL)
}

// githubAPIClientFromClient attempts to unwrap a REST client from a wrapped client value.
func githubAPIClientFromClient(value types.ClientInstance) *gh.Client {
	client, ok := types.ClientInstanceAs[*gh.Client](value)
	if !ok {
		return nil
	}

	return client
}

// githubRESTClientForOperation returns a pooled REST client or a token-derived fallback.
func githubRESTClientForOperation(ctx context.Context, input types.OperationInput) (*gh.Client, error) {
	return githubRESTClientForOperationWithBaseURL(ctx, input, githubAPIBaseURL)
}

// githubRESTClientForOperationWithBaseURL returns a pooled REST client or a token-derived fallback using the given base URL.
func githubRESTClientForOperationWithBaseURL(ctx context.Context, input types.OperationInput, baseURL string) (*gh.Client, error) {
	client := githubAPIClientFromClient(input.Client)
	if client != nil {
		return client, nil
	}

	token, err := auth.OAuthTokenFromPayload(input.Credential)
	if err != nil {
		return nil, err
	}

	return newGitHubAPIClient(ctx, token, baseURL)
}
