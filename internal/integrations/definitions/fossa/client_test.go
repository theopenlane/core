package fossa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *APIClient {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewAPIClient(context.Background(), server.URL, "token-123", true)
	assert.NilError(t, err)

	return client
}

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
			assert.Check(t, is.Equal(tt.expected, baseURLOrDefault(tt.input)))
		})
	}
}

func TestClientBuilderRequiresToken(t *testing.T) {
	_, err := buildClient(context.Background(), types.ConnectionRequest[CredentialSchema]{}, true)
	assert.Check(t, is.ErrorIs(err, ErrAPITokenMissing))
}

func TestClientRejectsNonPublicBaseURL(t *testing.T) {
	req := types.ConnectionRequest[CredentialSchema]{Credential: CredentialSchema{APIToken: "token-123", BaseURL: "http://127.0.0.1:8080"}}

	_, err := buildClient(context.Background(), req, false)
	assert.Check(t, err != nil)

	_, err = buildClient(context.Background(), req, true)
	assert.Check(t, is.Nil(err))
}

func TestClientSendsBearerTokenAndEncodesQuery(t *testing.T) {
	var (
		gotAuth     string
		gotRawQuery string
		gotPath     string
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotRawQuery = r.URL.RawQuery
		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[]}`))
	})

	issues, err := client.listIssues(context.Background(), categoryVulnerability, 1)
	assert.NilError(t, err)
	assert.Check(t, is.Len(issues, 0))

	assert.Check(t, is.Equal("Bearer token-123", gotAuth))
	assert.Check(t, is.Equal(pathIssues, gotPath))
	assert.Check(t, is.Contains(gotRawQuery, "scope%5Btype%5D=global"))
}

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
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})

			_, err := client.issueCategories(context.Background())
			assert.Check(t, is.ErrorIs(err, tt.expected))
		})
	}
}

func TestOrganizationIdentifier(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"organizationId":63774,"subscription":"Free","usesSAML":false}`))
	})

	org, err := client.organization(context.Background())
	assert.NilError(t, err)

	assert.Check(t, is.Equal("63774", org.identifier()))
	assert.Check(t, is.Equal("Free", org.Subscription))

	assert.Check(t, is.Equal("", organizationResponse{}.identifier()), "a missing organization ID must not render as 0")
}

func TestCheckAccess(t *testing.T) {
	tests := []struct {
		name          string
		organization  string
		categoryCode  int
		expected      InstallationMetadata
		expectedError error
	}{
		{
			name:         "full token",
			organization: `{"organizationId":63774,"subscription":"Free"}`,
			categoryCode: http.StatusOK,
			expected:     InstallationMetadata{OrganizationID: "63774", Subscription: "Free"},
		},
		{
			name:          "push only token cannot read issues",
			organization:  `{"organizationId":63774,"subscription":"Free"}`,
			categoryCode:  http.StatusForbidden,
			expectedError: ErrUnauthorized,
		},
		{
			name:          "organization without an id",
			organization:  `{}`,
			categoryCode:  http.StatusOK,
			expectedError: ErrOrganizationIDMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.URL.Path == pathOrganization {
					_, _ = w.Write([]byte(tt.organization))

					return
				}

				w.WriteHeader(tt.categoryCode)
				_, _ = w.Write([]byte(`{"vulnerability":1}`))
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
