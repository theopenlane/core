package githubapp

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

// GraphQLClient is the subset of the GitHub GraphQL client used by this definition
type GraphQLClient interface {
	// Query executes a GraphQL query against the GitHub API
	Query(ctx context.Context, q any, variables map[string]any) error
}

// graphQLClient wraps graphql.Client to satisfy the GraphQLClient interface
type graphQLClient struct {
	// client is the underlying shurcooL GraphQL client
	client *graphql.Client
}

// Query executes a GitHub GraphQL query using the underlying client
func (c *graphQLClient) Query(ctx context.Context, q any, variables map[string]any) error {
	return c.client.Query(ctx, q, variables)
}

// Client builds installation-scoped GitHub GraphQL clients
type Client struct {
	// AppConfig holds the operator-owned GitHub App settings used for token refresh.
	AppConfig Config
}

// Build constructs the GitHub GraphQL client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	credential, err := credentialFromBindings(req.Credentials)
	if err != nil {
		return nil, err
	}

	tokenSource := oauth2.ReuseTokenSource(
		tokenFromCredential(credential),
		installationTokenSource{
			ctx:            context.WithoutCancel(ctx),
			cfg:            tokenRefreshConfig(c.AppConfig, credential),
			installationID: credential.InstallationID,
		},
	)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	return newGraphQLClient(httpClient, c.AppConfig.APIURL), nil
}

// newGraphQLClient constructs a GitHub GraphQL client targeting the given API URL
func newGraphQLClient(httpClient *http.Client, apiURL string) GraphQLClient {
	endpoint := "https://api.github.com/graphql"
	if apiURL != "" {
		endpoint = strings.TrimRight(apiURL, "/") + "/api/graphql"
	}

	return &graphQLClient{client: graphql.NewClient(endpoint, httpClient)}
}

// installationTokenSource re-mints GitHub App installation tokens when the cached token expires.
type installationTokenSource struct {
	ctx            context.Context
	cfg            Config
	installationID int64
}

// Token returns a fresh installation token for the configured GitHub App installation.
func (s installationTokenSource) Token() (*oauth2.Token, error) {
	jwtToken, err := appJWT(s.cfg)
	if err != nil {
		return nil, err
	}

	return installationToken(s.ctx, s.cfg, s.installationID, jwtToken)
}

// credentialFromBindings extracts the GitHub App credential payload from credential bindings.
func credentialFromBindings(bindings types.CredentialBindings) (githubAppCredential, error) {
	cred, _, err := gitHubAppCredential.Resolve(bindings)
	if err != nil {
		return githubAppCredential{}, ErrCredentialDecode
	}

	if cred.InstallationID == 0 {
		return githubAppCredential{}, ErrInstallationIDMissing
	}

	if cred.AccessToken == "" || cred.Expiry == nil {
		return githubAppCredential{}, ErrAccessTokenMissing
	}

	return cred, nil
}

// tokenRefreshConfig fills refresh-only config from persisted credential data when possible.
func tokenRefreshConfig(cfg Config, credential githubAppCredential) Config {
	if cfg.AppID == "" && credential.AppID != 0 {
		cfg.AppID = strconv.FormatInt(credential.AppID, 10)
	}

	return cfg
}

// tokenFromCredential converts a persisted GitHub App credential into an oauth token seed.
func tokenFromCredential(credential githubAppCredential) *oauth2.Token {
	if credential.AccessToken == "" {
		return nil
	}

	token := &oauth2.Token{
		AccessToken: credential.AccessToken,
		TokenType:   "Bearer",
	}

	if credential.Expiry != nil {
		token.Expiry = credential.Expiry.UTC()
	}

	return token
}
