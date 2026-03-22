package githubapp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/httpsling"
)

// testPrivateKey generates a PEM-encoded RSA private key for testing
func testPrivateKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}

	return string(pem.EncodeToMemory(block))
}

// TestFlowStartProducesInstallURL verifies Start returns a valid GitHub App install URL and state
func TestFlowStartProducesInstallURL(t *testing.T) {
	t.Parallel()

	result, err := startAppInstall(Config{AppSlug: "my-app"})
	require.NoError(t, err)
	require.Contains(t, result.URL, "https://github.com/apps/my-app/installations/new")
	require.Contains(t, result.URL, "state=")
	require.NotEmpty(t, result.State)

	var state statePayload
	require.NoError(t, json.Unmarshal(result.State, &state))
	require.NotEmpty(t, state.Token)
}

// TestFlowStartMissingSlug verifies Start returns an error when AppSlug is empty
func TestFlowStartMissingSlug(t *testing.T) {
	t.Parallel()

	_, err := startAppInstall(Config{})
	require.ErrorIs(t, err, ErrAppSlugMissing)
}

// TestFlowCompleteDecodesCallbackAndProducesCredential verifies Complete mints a credential from callback input
func TestFlowCompleteDecodesCallbackAndProducesCredential(t *testing.T) {
	t.Parallel()

	tokenExpiry := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		_, _ = w.Write([]byte(`{"token":"ghs_test123","expires_at":"` + tokenExpiry.Format(time.RFC3339) + `"}`))
	}))
	defer server.Close()

	pk := testPrivateKey(t)

	cfg := Config{
		AppID:      "12345",
		PrivateKey: pk,
		AppSlug:    "test-app",
		APIURL:     server.URL,
	}

	stateJSON, err := jsonx.ToRawMessage(statePayload{Token: "csrf-token"})
	require.NoError(t, err)

	result, err := completeAppInstall(context.Background(), cfg, stateJSON, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "installation_id", Values: []string{"99"}},
		},
	})
	require.NoError(t, err)

	var cred githubAppCredential
	require.NoError(t, json.Unmarshal(result.Credential.Data, &cred))
	require.Equal(t, int64(12345), cred.AppID)
	require.Equal(t, int64(99), cred.InstallationID)
	require.Equal(t, "ghs_test123", cred.AccessToken)
	require.NotNil(t, cred.Expiry)
}

// TestFlowCompleteSetsInstallationInput verifies Complete populates InstallationInput with metadata
func TestFlowCompleteSetsInstallationInput(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		_, _ = w.Write([]byte(`{"token":"ghs_abc","expires_at":"2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	pk := testPrivateKey(t)

	cfg := Config{
		AppID:      "42",
		PrivateKey: pk,
		AppSlug:    "test-app",
		APIURL:     server.URL,
	}

	stateJSON, err := jsonx.ToRawMessage(statePayload{Token: "tok"})
	require.NoError(t, err)

	result, err := completeAppInstall(context.Background(), cfg, stateJSON, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "installation_id", Values: []string{"77"}},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.InstallationInput)

	var meta InstallationMetadata
	require.NoError(t, json.Unmarshal(result.InstallationInput, &meta))
	require.Equal(t, "77", meta.InstallationID)
}

// TestFlowCompleteMissingInstallationID verifies Complete fails when installation_id is zero
func TestFlowCompleteMissingInstallationID(t *testing.T) {
	t.Parallel()

	stateJSON, err := jsonx.ToRawMessage(statePayload{Token: "tok"})
	require.NoError(t, err)

	_, err = completeAppInstall(context.Background(), Config{AppSlug: "test-app"}, stateJSON, types.AuthCallbackInput{})
	require.ErrorIs(t, err, ErrInstallationIDMissing)
}

// TestFlowTokenViewDecodesCredential verifies TokenView returns the token and expiry from credential data
func TestFlowTokenViewDecodesCredential(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC)
	cred := githubAppCredential{
		AppID:          1,
		InstallationID: 2,
		AccessToken:    "ghs_viewtest",
		Expiry:         &expiry,
	}

	data, err := jsonx.ToRawMessage(cred)
	require.NoError(t, err)

	reg := appInstallAuthRegistration(Config{})
	view, err := reg.TokenView(context.Background(), types.CredentialSet{Data: data})
	require.NoError(t, err)
	require.Equal(t, "ghs_viewtest", view.AccessToken)
	require.NotNil(t, view.ExpiresAt)
	require.Equal(t, expiry, *view.ExpiresAt)
}

// TestFlowTokenViewEmptyData verifies TokenView returns empty values when credential data is nil
func TestFlowTokenViewEmptyData(t *testing.T) {
	t.Parallel()

	reg := appInstallAuthRegistration(Config{})
	view, err := reg.TokenView(context.Background(), types.CredentialSet{})
	require.NoError(t, err)
	require.Empty(t, view.AccessToken)
	require.Nil(t, view.ExpiresAt)
}

// TestFlowRefreshMissingInstallationID verifies Refresh fails when credential has no installation ID
func TestFlowRefreshMissingInstallationID(t *testing.T) {
	t.Parallel()

	cred := githubAppCredential{AppID: 1, AccessToken: "old"}
	data, err := jsonx.ToRawMessage(cred)
	require.NoError(t, err)

	_, err = refreshAppInstall(context.Background(), Config{AppID: "1", PrivateKey: testPrivateKey(t)}, types.CredentialSet{Data: data})
	require.ErrorIs(t, err, ErrInstallationIDMissing)
}

// TestFlowCompleteURLPath verifies the installation token request hits the expected API path
func TestFlowCompleteURLPath(t *testing.T) {
	t.Parallel()

	var requestPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		_, _ = w.Write([]byte(`{"token":"ghs_path","expires_at":"2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	pk := testPrivateKey(t)

	cfg := Config{
		AppID:      "10",
		PrivateKey: pk,
		AppSlug:    "test-app",
		APIURL:     server.URL,
	}

	stateJSON, err := jsonx.ToRawMessage(statePayload{Token: "t"})
	require.NoError(t, err)

	_, err = completeAppInstall(context.Background(), cfg, stateJSON, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "installation_id", Values: []string{"55"}},
		},
	})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(requestPath, "/api/v3/app/installations/55/access_tokens"))
}
