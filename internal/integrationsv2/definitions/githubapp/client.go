package githubapp

import (
	"context"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v83/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrationsv2/types"
)

const (
	githubAPIVersion = "2022-11-28"
	githubAPIBaseURL = "https://api.github.com"
)

var githubClientHeaders = map[string]string{
	"Accept":               "application/vnd.github+json",
	"X-GitHub-Api-Version": githubAPIVersion,
}

// githubHeaderTransport injects static headers into outgoing GitHub requests
type githubHeaderTransport struct {
	next    http.RoundTripper
	headers map[string]string
}

// RoundTrip applies configured headers and delegates to the wrapped transport
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

// buildRESTClient builds the GitHub REST API client for one installation
func (d *def) buildRESTClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	token, err := tokenFromCredential(req.Credential)
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(d.cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = githubAPIBaseURL
	}

	return newGitHubAPIClient(ctx, token, baseURL)
}

// buildGraphQLClient builds the GitHub GraphQL API client for one installation
func (d *def) buildGraphQLClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	token, err := tokenFromCredential(req.Credential)
	if err != nil {
		return nil, err
	}

	return newGitHubGraphQLClient(ctx, token), nil
}

// tokenFromCredential extracts the OAuth access token from a credential set
func tokenFromCredential(credential types.CredentialSet) (string, error) {
	if credential.OAuthAccessToken == "" {
		return "", ErrOAuthTokenMissing
	}

	return credential.OAuthAccessToken, nil
}

// newGitHubAPIClient initializes an authenticated GitHub REST client
func newGitHubAPIClient(ctx context.Context, token, baseURL string) (*gh.Client, error) {
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

// newGitHubGraphQLClient initializes an authenticated GitHub GraphQL client
func newGitHubGraphQLClient(ctx context.Context, token string) *githubv4.Client {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, source)
	httpClient.Transport = githubHeaderTransport{
		next:    httpClient.Transport,
		headers: githubClientHeaders,
	}

	return githubv4.NewClient(httpClient)
}

// restClientFromAny casts client to *gh.Client
func restClientFromAny(client any) (*gh.Client, error) {
	c, ok := client.(*gh.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}

// graphQLClientFromAny casts client to *githubv4.Client
func graphQLClientFromAny(client any) (*githubv4.Client, error) {
	c, ok := client.(*githubv4.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
