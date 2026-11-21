package aikido

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	testClientID     = "client-123"
	testClientSecret = "secret-456"
	testAccessToken  = "token-789"
)

func tokenHandler(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	assert.Check(t, is.Equal(http.MethodPost, r.Method))
	assert.Check(t, is.Equal("Basic "+base64.StdEncoding.EncodeToString([]byte(testClientID+":"+testClientSecret)), r.Header.Get("Authorization")))
	assert.Check(t, r.ParseForm())
	assert.Check(t, is.Equal("client_credentials", r.PostForm.Get("grant_type")))

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"` + testAccessToken + `","token_type":"bearer","expires_in":3600}`))
}

func newTestClient(t *testing.T, api http.HandlerFunc) *APIClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathToken {
			tokenHandler(t, w, r)

			return
		}

		if !strings.HasPrefix(r.URL.Path, pathAPI+"/") {
			http.NotFound(w, r)

			return
		}

		api(w, r)
	}))
	t.Cleanup(server.Close)

	client, err := NewAPIClient(context.Background(), server.URL, testClientID, testClientSecret)
	assert.NilError(t, err)

	client.limiter = rate.NewLimiter(rate.Inf, 1)

	return client
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

func TestAppURLForRegion(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		expected    string
		expectedErr error
	}{
		{name: "empty defaults to eu", region: "", expected: "https://app.aikido.dev"},
		{name: "us", region: "us", expected: "https://app.us.aikido.dev"},
		{name: "case and whitespace", region: " AU ", expected: "https://app.au.aikido.dev"},
		{name: "unknown", region: "mars", expectedErr: ErrRegionInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appURLForRegion(tt.region)
			if tt.expectedErr != nil {
				assert.Check(t, is.ErrorIs(err, tt.expectedErr))

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected, got))
		})
	}
}

func TestBuildClientRequiresCredentials(t *testing.T) {
	_, err := buildClient(context.Background(), types.ConnectionRequest[CredentialSchema]{Credential: CredentialSchema{ClientID: testClientID}})
	assert.Check(t, is.ErrorIs(err, ErrClientCredentialsMissing))

	_, err = buildClient(context.Background(), types.ConnectionRequest[CredentialSchema]{Credential: CredentialSchema{ClientID: testClientID, ClientSecret: testClientSecret, Region: "mars"}})
	assert.Check(t, is.ErrorIs(err, ErrRegionInvalid))
}

func TestClientExchangesTokenAndSendsBearer(t *testing.T) {
	var gotAuth string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		assert.Check(t, is.Equal(pathWorkspace, r.URL.Path))
		writeJSON(w, `{"id":1,"name":"Pied Piper","linked_provider":"github"}`)
	})

	workspace, err := client.workspace(context.Background())
	assert.NilError(t, err)

	assert.Check(t, is.Equal("Bearer "+testAccessToken, gotAuth))
	assert.Check(t, is.Equal("1", workspace.identifier()))
	assert.Check(t, is.Equal("Pied Piper", workspace.Name))
	assert.Check(t, is.Equal("", workspaceResponse{}.identifier()))
}

func TestClientRejectedToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"The provided client credentials are invalid"}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewAPIClient(context.Background(), server.URL, testClientID, testClientSecret)
	assert.NilError(t, err)

	_, err = client.workspace(context.Background())
	assert.Check(t, is.ErrorIs(err, ErrUnauthorized))
}

func TestClientStatusErrors(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected error
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, expected: ErrUnauthorized},
		{name: "forbidden", status: http.StatusForbidden, expected: ErrUnauthorized},
		{name: "server error", status: http.StatusInternalServerError, expected: ErrAPIRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"error":"nope"}`))
			})

			_, err := client.workspace(context.Background())
			assert.Check(t, is.ErrorIs(err, tt.expected))
		})
	}
}

func TestClientRetriesRateLimit(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		writeJSON(w, `{"id":1,"name":"Pied Piper"}`)
	})

	_, err := client.workspace(context.Background())
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int32(2), requests.Load()))
}

func TestClientGivesUpAfterRateLimitRetries(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := client.workspace(context.Background())
	assert.Check(t, is.ErrorIs(err, ErrRateLimited))
	assert.Check(t, is.Equal(int32(maxRateLimitRetries+1), requests.Load()))
}

func TestRetryAfterDelay(t *testing.T) {
	assert.Check(t, is.Equal(maxRetryAfter, retryAfterDelay("")))
	assert.Check(t, is.Equal(maxRetryAfter, retryAfterDelay("soon")))
	assert.Check(t, is.Equal(maxRetryAfter, retryAfterDelay("600")))
	assert.Check(t, is.Equal(int64(5), int64(retryAfterDelay("5").Seconds())))
}

func TestCheckAccess(t *testing.T) {
	tests := []struct {
		name          string
		workspace     string
		issuesStatus  int
		expected      InstallationMetadata
		expectedError error
	}{
		{
			name:         "full access",
			workspace:    `{"id":1,"name":"Pied Piper","linked_provider":"github"}`,
			issuesStatus: http.StatusOK,
			expected:     InstallationMetadata{WorkspaceID: "1", WorkspaceName: "Pied Piper", LinkedProvider: "github"},
		},
		{
			name:          "missing issues scope",
			workspace:     `{"id":1,"name":"Pied Piper"}`,
			issuesStatus:  http.StatusForbidden,
			expectedError: ErrUnauthorized,
		},
		{
			name:          "workspace without an id",
			workspace:     `{}`,
			issuesStatus:  http.StatusOK,
			expectedError: ErrWorkspaceIDMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == pathWorkspace {
					writeJSON(w, tt.workspace)

					return
				}

				assert.Check(t, is.Equal(pathIssuesExport, r.URL.Path))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.issuesStatus)
				_, _ = w.Write([]byte(`[]`))
			})

			metadata, err := client.CheckAccess(context.Background())
			if tt.expectedError != nil {
				assert.Check(t, is.ErrorIs(err, tt.expectedError))

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, metadata))
		})
	}
}

func TestBuildOAuthClientRequiresAccessToken(t *testing.T) {
	_, err := buildOAuthClient(context.Background(), Config{}, oauthAppURL, aikidoOAuthCred{})
	assert.Check(t, is.ErrorIs(err, ErrOAuthTokenMissing))
}

func TestOAuthClientUsesStoredToken(t *testing.T) {
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"id":1,"name":"Pied Piper"}`)
	}))
	t.Cleanup(server.Close)

	expiry := time.Now().Add(time.Hour)

	client, err := buildOAuthClient(context.Background(), Config{ClientID: testClientID, ClientSecret: testClientSecret}, server.URL, aikidoOAuthCred{AccessToken: "stored-token", RefreshToken: "refresh-1", Expiry: &expiry})
	assert.NilError(t, err)

	_, err = client.workspace(context.Background())
	assert.NilError(t, err)
	assert.Check(t, is.Equal("Bearer stored-token", gotAuth))
}

func TestOAuthClientRefreshesExpiredToken(t *testing.T) {
	var (
		gotGrant   string
		gotRefresh string
		gotAuth    string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathToken {
			assert.Check(t, r.ParseForm())
			gotGrant = r.PostForm.Get("grant_type")
			gotRefresh = r.PostForm.Get("refresh_token")

			writeJSON(w, `{"access_token":"`+testAccessToken+`","token_type":"bearer","expires_in":3600}`)

			return
		}

		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"id":1,"name":"Pied Piper"}`)
	}))
	t.Cleanup(server.Close)

	expired := time.Now().Add(-time.Hour)

	client, err := buildOAuthClient(context.Background(), Config{ClientID: testClientID, ClientSecret: testClientSecret}, server.URL, aikidoOAuthCred{AccessToken: "stale-token", RefreshToken: "refresh-1", Expiry: &expired})
	assert.NilError(t, err)

	_, err = client.workspace(context.Background())
	assert.NilError(t, err)
	assert.Check(t, is.Equal("refresh_token", gotGrant))
	assert.Check(t, is.Equal("refresh-1", gotRefresh))
	assert.Check(t, is.Equal("Bearer "+testAccessToken, gotAuth))
}

func TestOAuthMaterial(t *testing.T) {
	expiry := time.Now().UTC()

	cred, err := oauthMaterial(auth.OAuthMaterial{AccessToken: "a", RefreshToken: "r", Expiry: &expiry})
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(aikidoOAuthCred{AccessToken: "a", RefreshToken: "r", Expiry: &expiry}, cred))
}
