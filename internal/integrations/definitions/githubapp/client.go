package githubapp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
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

// appClient builds the installation-scoped GitHub GraphQL client from the decoded credential
func appClient(cfg Config) func(context.Context, types.ConnectionRequest[githubAppCredential]) (GraphQLClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[githubAppCredential]) (GraphQLClient, error) {
		credential := req.Credential

		if credential.InstallationID == 0 {
			return nil, ErrInstallationIDMissing
		}

		if credential.AccessToken == "" || credential.Expiry == nil {
			return nil, ErrAccessTokenMissing
		}

		tokenSource := oauth2.ReuseTokenSource(
			tokenFromCredential(credential),
			installationTokenSource{
				ctx:            context.WithoutCancel(ctx),
				cfg:            tokenRefreshConfig(cfg, credential),
				installationID: credential.InstallationID,
			},
		)
		httpClient := oauth2.NewClient(ctx, tokenSource)

		return newGraphQLClient(httpClient, cfg.APIURL)
	}
}

// enterpriseAPIPath is the REST API path under a GitHub Enterprise Server host
const enterpriseAPIPath = "api/v3/"

// enterpriseUploadPath is the upload API path under a GitHub Enterprise Server host
const enterpriseUploadPath = "api/uploads/"

// enterpriseGraphQLPath is the GraphQL API path under a GitHub Enterprise Server host
const enterpriseGraphQLPath = "api/graphql"

// defaultGraphQLEndpoint is the GraphQL endpoint for github.com
const defaultGraphQLEndpoint = "https://api.github.com/graphql"

// newGraphQLClient constructs a GitHub GraphQL client targeting the given API URL
func newGraphQLClient(httpClient *http.Client, apiURL string) (GraphQLClient, error) {
	endpoint := defaultGraphQLEndpoint

	if apiURL != "" {
		parsed, err := urlx.ParseAbsolute(apiURL)
		if err != nil {
			return nil, err
		}

		endpoint = parsed.JoinPath(enterpriseGraphQLPath).String()
	}

	return &graphQLClient{client: graphql.NewClient(endpoint, httpClient)}, nil
}

// installationTokenSource re-mints GitHub App installation tokens when the cached token expires
type installationTokenSource struct {
	ctx            context.Context
	cfg            Config
	installationID int64
}

// Token returns a fresh installation token for the configured GitHub App installation
func (s installationTokenSource) Token() (*oauth2.Token, error) {
	jwtToken, err := appJWT(s.cfg)
	if err != nil {
		return nil, err
	}

	return installationToken(s.ctx, s.cfg, s.installationID, jwtToken)
}

// tokenRefreshConfig fills refresh-only config from persisted credential data when possible
func tokenRefreshConfig(cfg Config, credential githubAppCredential) Config {
	if cfg.AppID == "" && credential.AppID != 0 {
		cfg.AppID = strconv.FormatInt(credential.AppID, 10)
	}

	return cfg
}

// tokenFromCredential converts a persisted GitHub App credential into an oauth token seed
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
