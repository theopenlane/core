package githubapp

import (
	"context"
	"net/http"
	"strings"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

// GraphQLClient is the subset of the GitHub GraphQL client used by this definition.
type GraphQLClient interface {
	// Query executes a GraphQL query against the GitHub API
	Query(ctx context.Context, q any, variables map[string]any) error
}

type graphQLClient struct {
	client *graphql.Client
}

// Query executes a GitHub GraphQL query using the underlying client
func (c *graphQLClient) Query(ctx context.Context, q any, variables map[string]any) error {
	return c.client.Query(ctx, q, variables)
}

// Client builds installation-scoped GitHub GraphQL clients
type Client struct {
	// APIURL overrides the GitHub API host for local tests.
	APIURL string
}

// Build constructs the GitHub GraphQL client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	token, err := tokenFromCredential(req.Credential)
	if err != nil {
		return nil, err
	}

	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, source)

	return newGraphQLClient(httpClient, c.APIURL), nil
}

// FromAny casts a registered client instance to the GitHub GraphQL client type
func (Client) FromAny(value any) (GraphQLClient, error) {
	client, ok := value.(GraphQLClient)
	if !ok {
		return nil, ErrClientType
	}

	return client, nil
}

// newGraphQLClient constructs a GitHub GraphQL client targeting the given API URL
func newGraphQLClient(httpClient *http.Client, apiURL string) GraphQLClient {
	endpoint := "https://api.github.com/graphql"
	if apiURL != "" {
		endpoint = strings.TrimRight(apiURL, "/") + "/api/graphql"
	}

	return &graphQLClient{client: graphql.NewClient(endpoint, httpClient)}
}

// tokenFromCredential extracts the OAuth access token from a credential set
func tokenFromCredential(credential types.CredentialSet) (string, error) {
	if credential.OAuthAccessToken == "" {
		return "", ErrOAuthTokenMissing
	}

	return credential.OAuthAccessToken, nil
}
