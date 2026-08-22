package fossa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestBaseURLOrDefault verifies base URL normalization and the SaaS fallback
func TestBaseURLOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty falls back to saas", input: "", expected: defaultBaseURL},
		{name: "whitespace falls back to saas", input: "   ", expected: defaultBaseURL},
		{name: "trailing slash trimmed", input: "https://fossa.internal/", expected: "https://fossa.internal"},
		{name: "on premise host preserved", input: "https://fossa.internal", expected: "https://fossa.internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, baseURLOrDefault(tt.input))
		})
	}
}

// TestResolveCredential verifies credential resolution from bindings
func TestResolveCredential(t *testing.T) {
	t.Run("decodes into credential schema", func(t *testing.T) {
		raw, err := jsonx.ToRawMessage(CredentialSchema{APIToken: "token-123", BaseURL: "https://fossa.internal"})
		require.NoError(t, err)

		bindings := types.CredentialBindings{
			{Ref: fossaCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		}

		cred, err := resolveCredential(bindings)
		require.NoError(t, err)

		assert.Equal(t, "token-123", cred.APIToken)
		assert.Equal(t, "https://fossa.internal", cred.BaseURL)
	})

	t.Run("returns decode error for invalid provider data", func(t *testing.T) {
		bindings := types.CredentialBindings{
			{Ref: fossaCredential.ID(), Credential: types.CredentialSet{Data: []byte(`{`)}},
		}

		_, err := resolveCredential(bindings)
		assert.ErrorIs(t, err, ErrCredentialDecode)
	})

	t.Run("returns decode error when unbound", func(t *testing.T) {
		_, err := resolveCredential(types.CredentialBindings{})
		assert.ErrorIs(t, err, ErrCredentialDecode)
	})
}

// TestClientBuilderRequiresToken verifies a bound credential with no token is rejected up front
func TestClientBuilderRequiresToken(t *testing.T) {
	raw, err := jsonx.ToRawMessage(CredentialSchema{})
	require.NoError(t, err)

	bindings := types.CredentialBindings{
		{Ref: fossaCredential.ID(), Credential: types.CredentialSet{Data: raw}},
	}

	_, err = ClientBuilder{}.Build(context.Background(), types.ClientBuildRequest{Credentials: bindings})
	assert.ErrorIs(t, err, ErrAPITokenMissing)
}

// TestClientSendsBearerTokenAndEncodesQuery verifies the auth header is applied and that bracketed
// query keys such as scope[type] are percent-encoded rather than passed through literally
func TestClientSendsBearerTokenAndEncodesQuery(t *testing.T) {
	var (
		gotAuth     string
		gotRawQuery string
		gotPath     string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotRawQuery = r.URL.RawQuery
		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, "token-123")
	require.NoError(t, err)

	issues, err := client.listIssues(context.Background(), categoryVulnerability, statusActive, 1)
	require.NoError(t, err)
	assert.Empty(t, issues)

	assert.Equal(t, "Bearer token-123", gotAuth)
	assert.Equal(t, pathIssues, gotPath)
	assert.Contains(t, gotRawQuery, "scope%5Btype%5D=global")
	assert.NotContains(t, gotRawQuery, "scope[type]")
}

// TestClientStatusErrors verifies non-success responses map to sentinels rather than decode errors.
// httpsling does not treat a non-2xx status as an error and still unmarshals the body, so this
// guards the ordering of the status check against the decode result.
func TestClientStatusErrors(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		expected error
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{"error":"bad token"}`, expected: ErrUnauthorized},
		{name: "forbidden", status: http.StatusForbidden, body: `{"error":"push only token"}`, expected: ErrUnauthorized},
		{name: "rate limited", status: http.StatusTooManyRequests, body: `slow down`, expected: ErrRateLimited},
		{name: "server error", status: http.StatusInternalServerError, body: `boom`, expected: ErrAPIRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client, err := NewAPIClient(server.URL, "token-123")
			require.NoError(t, err)

			_, err = client.issueCategories(context.Background())
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}

// TestOrganizationIdentifier verifies the numeric organization ID is rendered as a stable string
func TestOrganizationIdentifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"organizationId":63774,"subscription":"Free","usesSAML":false}`))
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, "token-123")
	require.NoError(t, err)

	org, err := client.organization(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "63774", org.identifier())
	assert.Equal(t, "Free", org.Subscription)

	assert.Empty(t, organizationResponse{}.identifier(), "a missing organization ID must not render as 0")
}
