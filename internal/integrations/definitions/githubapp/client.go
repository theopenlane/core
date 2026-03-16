package githubapp

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

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
	if c.APIURL != "" {
		return githubv4.NewEnterpriseClient(c.APIURL+"/api/graphql", httpClient), nil
	}

	return githubv4.NewClient(httpClient), nil
}

// FromAny casts a registered client instance to the GitHub GraphQL client type
func (Client) FromAny(value any) (*githubv4.Client, error) {
	client, ok := value.(*githubv4.Client)
	if !ok {
		return nil, ErrClientType
	}

	return client, nil
}

// tokenFromCredential extracts the OAuth access token from a credential set
func tokenFromCredential(credential types.CredentialSet) (string, error) {
	if credential.OAuthAccessToken == "" {
		return "", ErrOAuthTokenMissing
	}

	return credential.OAuthAccessToken, nil
}
