package githubapp

import (
	"context"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

const githubAPIBaseURL = "https://api.github.com"

// Client builds installation-scoped GitHub GraphQL clients
type Client struct {
	// Config holds the operator-supplied GitHub App settings
	Config Config
}

// Build constructs the GitHub GraphQL client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	token, err := tokenFromCredential(req.Credential)
	if err != nil {
		return nil, err
	}

	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, source)

	endpoint := c.enterpriseGraphQLEndpoint()
	if endpoint == "" {
		return githubv4.NewClient(httpClient), nil
	}

	return githubv4.NewEnterpriseClient(endpoint, httpClient), nil
}

// FromAny casts a registered client instance to the GitHub GraphQL client type
func (Client) FromAny(value any) (*githubv4.Client, error) {
	client, ok := value.(*githubv4.Client)
	if !ok {
		return nil, ErrClientType
	}

	return client, nil
}

// enterpriseGraphQLEndpoint derives the GraphQL endpoint for GitHub Enterprise installations
func (c Client) enterpriseGraphQLEndpoint() string {
	baseURL := strings.TrimRight(c.Config.BaseURL, "/")
	if baseURL == "" || baseURL == githubAPIBaseURL {
		return ""
	}

	enterpriseBaseURL := strings.TrimSuffix(baseURL, "/api/v3")
	if enterpriseBaseURL == "" {
		enterpriseBaseURL = baseURL
	}

	return enterpriseBaseURL + "/api/graphql"
}

// tokenFromCredential extracts the OAuth access token from a credential set
func tokenFromCredential(credential types.CredentialSet) (string, error) {
	if credential.OAuthAccessToken == "" {
		return "", ErrOAuthTokenMissing
	}

	return credential.OAuthAccessToken, nil
}
