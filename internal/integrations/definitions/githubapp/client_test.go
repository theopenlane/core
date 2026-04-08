package githubapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/httpsling"
)

// TestQueryRepositoriesUsesConfiguredGraphQLEndpoint verifies that the client uses the configured API URL for GraphQL queries
func TestQueryRepositoriesUsesConfiguredGraphQLEndpoint(t *testing.T) {
	t.Parallel()

	privateKey := testPrivateKey(t)

	var requestPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath = req.URL.Path
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		_, _ = w.Write([]byte(`{"data":{"viewer":{"repositories":{"nodes":[{"nameWithOwner":"acme/demo","isPrivate":false,"updatedAt":"2030-01-01T00:00:00Z","url":"https://github.example/acme/demo"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`))
	}))
	defer server.Close()

	futureExpiry := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)

	clientValue, err := Client{AppConfig: Config{
		AppID:      "1",
		PrivateKey: privateKey,
		APIURL:     server.URL,
	}}.Build(context.Background(), types.ClientBuildRequest{
		Credentials: types.CredentialBindings{{
			Ref: gitHubAppCredential.ID(),
			Credential: types.CredentialSet{
				Data: json.RawMessage(`{"appId":1,"installationId":2,"accessToken":"token","expiry":"` + futureExpiry + `"}`),
			},
		}},
	})
	require.NoError(t, err)

	client, err := gitHubClient.Cast(clientValue)
	require.NoError(t, err)

	repositories, err := queryRepositories(context.Background(), client, 1)
	require.NoError(t, err)
	require.Len(t, repositories, 1)
	require.Equal(t, "/api/graphql", requestPath)
}

// TestCredentialFromBindings verifies credential extraction and validation
func TestCredentialFromBindings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		wantErr error
	}{
		{
			name:    "valid token and expiry",
			data:    `{"appId":1,"installationId":2,"accessToken":"tok","expiry":"2099-01-01T00:00:00Z"}`,
			wantErr: nil,
		},
		{
			name:    "missing installation ID",
			data:    `{"appId":1,"installationId":0,"accessToken":"tok","expiry":"2099-01-01T00:00:00Z"}`,
			wantErr: ErrInstallationIDMissing,
		},
		{
			name:    "missing access token",
			data:    `{"appId":1,"installationId":2,"expiry":"2099-01-01T00:00:00Z"}`,
			wantErr: ErrAccessTokenMissing,
		},
		{
			name:    "missing expiry",
			data:    `{"appId":1,"installationId":2,"accessToken":"tok"}`,
			wantErr: ErrAccessTokenMissing,
		},
		{
			name:    "missing both token and expiry",
			data:    `{"appId":1,"installationId":2}`,
			wantErr: ErrAccessTokenMissing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := credentialFromBindings(types.CredentialBindings{{
				Ref:        gitHubAppCredential.ID(),
				Credential: types.CredentialSet{Data: json.RawMessage(tc.data)},
			}})

			switch {
			case tc.wantErr != nil:
				require.ErrorIs(t, err, tc.wantErr)
			default:
				require.NoError(t, err)
			}
		})
	}
}

// TestTokenRefreshConfigFillsAppID verifies AppID is filled from credential when missing from operator config
func TestTokenRefreshConfigFillsAppID(t *testing.T) {
	t.Parallel()

	cfg := tokenRefreshConfig(Config{}, githubAppCredential{AppID: 42})
	require.Equal(t, "42", cfg.AppID)

	cfg = tokenRefreshConfig(Config{AppID: "99"}, githubAppCredential{AppID: 42})
	require.Equal(t, "99", cfg.AppID)
}

// TestBuildRefreshesExpiredInstallationToken verifies the client re-mints an installation token when the persisted one is expired.
func TestBuildRefreshesExpiredInstallationToken(t *testing.T) {
	t.Parallel()

	privateKey := testPrivateKey(t)
	expired := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)

	var graphQLAuth string
	var tokenRequestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/v3/app/installations/2/access_tokens":
			tokenRequestCount++
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			_, _ = w.Write([]byte(`{"token":"ghs_fresh","expires_at":"2099-01-01T00:00:00Z"}`))
		case "/api/graphql":
			graphQLAuth = req.Header.Get("Authorization")
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			_, _ = w.Write([]byte(`{"data":{"viewer":{"repositories":{"nodes":[{"nameWithOwner":"acme/demo","isPrivate":false,"updatedAt":"2030-01-01T00:00:00Z","url":"https://github.example/acme/demo"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	clientValue, err := Client{AppConfig: Config{
		AppID:      "1",
		PrivateKey: privateKey,
		APIURL:     server.URL,
	}}.Build(context.Background(), types.ClientBuildRequest{
		Credentials: types.CredentialBindings{{
			Ref: gitHubAppCredential.ID(),
			Credential: types.CredentialSet{
				Data: json.RawMessage(`{"appId":1,"installationId":2,"accessToken":"ghs_stale","expiry":"` + expired.Format(time.RFC3339) + `"}`),
			},
		}},
	})
	require.NoError(t, err)

	client, err := gitHubClient.Cast(clientValue)
	require.NoError(t, err)

	repositories, err := queryRepositories(context.Background(), client, 1)
	require.NoError(t, err)
	require.Len(t, repositories, 1)
	require.Equal(t, 1, tokenRequestCount)
	require.Equal(t, "Bearer ghs_fresh", graphQLAuth)
}
