package github

import (
	"context"
	"net/http"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const defaultGitHubGraphQLEndpoint = "https://api.github.com/graphql"

// githubHeaderTransport injects static headers into outgoing GitHub requests.
type githubHeaderTransport struct {
	// next is the wrapped transport.
	next http.RoundTripper
	// headers are the static headers to apply.
	headers map[string]string
}

// RoundTrip applies configured headers and delegates to the wrapped transport.
func (t githubHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	next := t.next
	if next == nil {
		next = http.DefaultTransport
	}

	clone := req.Clone(req.Context())
	for key, value := range t.headers {
		if clone.Header.Get(key) == "" {
			clone.Header.Set(key, value)
		}
	}

	return next.RoundTrip(clone)
}

// buildGitHubGraphQLClient returns a pooled client builder for the GitHub GraphQL API.
func buildGitHubGraphQLClient(endpoint string) types.ClientBuilderFunc {
	return func(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
		token, err := auth.OAuthTokenFromPayload(payload)
		if err != nil {
			return nil, err
		}

		return newGitHubGraphQLClient(token, endpoint), nil
	}
}

// newGitHubGraphQLClient initializes an authenticated GitHub GraphQL client.
func newGitHubGraphQLClient(token, endpoint string) *githubv4.Client {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), source)
	httpClient.Transport = githubHeaderTransport{
		next:    httpClient.Transport,
		headers: githubClientHeaders,
	}

	return githubv4.NewEnterpriseClient(endpoint, httpClient)
}

// githubGraphQLClientFromAny attempts to unwrap a GraphQL client from an arbitrary value.
func githubGraphQLClientFromAny(value any) *githubv4.Client {
	client, ok := value.(*githubv4.Client)
	if !ok {
		return nil
	}

	return client
}
