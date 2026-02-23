package github

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
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
	descriptors := auth.DefaultClientDescriptors(provider, ClientGitHubAPI, "GitHub REST API client", auth.OAuthClientBuilder(githubClientHeaders))
	descriptors = append(descriptors, auth.DefaultClientDescriptor(
		provider,
		ClientGitHubGraphQL,
		"GitHub GraphQL API client",
		buildGitHubGraphQLClient(defaultGitHubGraphQLEndpoint),
	))

	return descriptors
}
