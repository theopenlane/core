package githubapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/httpsling"
)

// TestQueryRepositoriesUsesConfiguredGraphQLEndpoint verifies that the client uses the configured API URL for GraphQL queries
func TestQueryRepositoriesUsesConfiguredGraphQLEndpoint(t *testing.T) {
	t.Parallel()

	var requestPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath = req.URL.Path
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		_, _ = w.Write([]byte(`{"data":{"viewer":{"repositories":{"nodes":[{"nameWithOwner":"acme/demo","isPrivate":false,"updatedAt":"2030-01-01T00:00:00Z","url":"https://github.example/acme/demo"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`))
	}))
	defer server.Close()

	clientValue, err := Client{APIURL: server.URL}.Build(context.Background(), types.ClientBuildRequest{
		Credential: types.CredentialSet{OAuthAccessToken: "token"},
	})
	require.NoError(t, err)

	client, err := GitHubClient.Cast(clientValue)
	require.NoError(t, err)

	repositories, err := queryRepositories(context.Background(), client, 1)
	require.NoError(t, err)
	require.Len(t, repositories, 1)
	require.Equal(t, "/api/graphql", requestPath)
}
