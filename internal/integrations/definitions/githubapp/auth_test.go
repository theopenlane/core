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

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/utils/jsonx"
)

// testPrivateKey generates a PEM-encoded RSA private key for testing
func testPrivateKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NilError(t, err)

	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}

	return string(pem.EncodeToMemory(block))
}

// TestFlowStartProducesInstallURL verifies Start returns a valid GitHub App install URL and state
func TestFlowStartProducesInstallURL(t *testing.T) {
	t.Parallel()

	result, err := startAppInstall(Config{AppSlug: "my-app"})
	assert.NilError(t, err)
	assert.Assert(t, is.Contains(result.URL, "https://github.com/apps/my-app/installations/new"))
	assert.Assert(t, is.Contains(result.URL, "state="))
	assert.Assert(t, len(result.State) > 0)

	var state statePayload
	assert.NilError(t, json.Unmarshal(result.State, &state))
	assert.Assert(t, len(state.Token) > 0)
}

// TestFlowStartMissingSlug verifies Start returns an error when AppSlug is empty
func TestFlowStartMissingSlug(t *testing.T) {
	t.Parallel()

	_, err := startAppInstall(Config{})
	assert.ErrorIs(t, err, ErrAppSlugMissing)
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
	assert.NilError(t, err)

	result, err := completeAppInstall(context.Background(), cfg, stateJSON, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "installation_id", Values: []string{"99"}},
		},
	})
	assert.NilError(t, err)

	var cred githubAppCredential
	assert.NilError(t, json.Unmarshal(result.Credential.Data, &cred))
	assert.Assert(t, is.Equal(int64(12345), cred.AppID))
	assert.Assert(t, is.Equal(int64(99), cred.InstallationID))
	assert.Assert(t, is.Equal("ghs_test123", cred.AccessToken))
	assert.Assert(t, cred.Expiry != nil)
}

// TestFlowCompleteMissingInstallationID verifies Complete fails when installation_id is zero
func TestFlowCompleteMissingInstallationID(t *testing.T) {
	t.Parallel()

	stateJSON, err := jsonx.ToRawMessage(statePayload{Token: "tok"})
	assert.NilError(t, err)

	_, err = completeAppInstall(context.Background(), Config{AppSlug: "test-app"}, stateJSON, types.AuthCallbackInput{})
	assert.ErrorIs(t, err, ErrInstallationIDMissing)
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
	assert.NilError(t, err)

	_, err = completeAppInstall(context.Background(), cfg, stateJSON, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "installation_id", Values: []string{"55"}},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, strings.HasPrefix(requestPath, "/api/v3/app/installations/55/access_tokens"))
}
