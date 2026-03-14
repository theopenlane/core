package githuboauth

import (
	"context"
	"net/http"

	gh "github.com/google/go-github/v83/github"
	githubv4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	githubAPIVersion = "2022-11-28"
	githubAPIBaseURL = "https://api.github.com"
)

var githubClientHeaders = map[string]string{
	"Accept":               "application/vnd.github+json",
	"X-GitHub-Api-Version": githubAPIVersion,
}

// githubHeaderTransport injects required GitHub API headers on every request
type githubHeaderTransport struct {
	next    http.RoundTripper
	headers map[string]string
}

func (t githubHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	for k, v := range t.headers {
		r.Header.Set(k, v)
	}
	return t.next.RoundTrip(r)
}

// buildRESTClient builds the GitHub REST client using the OAuth access token
func buildRESTClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	client, err := newGitHubRESTClient(ctx, token)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// buildGraphQLClient builds the GitHub GraphQL client using the OAuth access token
func buildGraphQLClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	return githubv4.NewClient(httpClient), nil
}

// newGitHubRESTClient initializes an authenticated GitHub REST client
func newGitHubRESTClient(ctx context.Context, token string) (*gh.Client, error) {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, src)
	httpClient.Transport = githubHeaderTransport{
		next:    httpClient.Transport,
		headers: githubClientHeaders,
	}

	return gh.NewClient(httpClient), nil
}

// restClientFromAny asserts the client as a GitHub REST client
func restClientFromAny(client any) (*gh.Client, error) {
	c, ok := client.(*gh.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
