package handlers_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
)

// GitHub App install callback test path.
const (
	githubAppInstallPath  = "/v1/integrations/github/app/install"
	githubAppCallbackPath = "/v1/integrations/github/app/callback"
)

// TestGitHubAppInstallCallback_RedirectsWhenConfigured verifies the GitHub App callback redirects when configured.
func (suite *HandlerTestSuite) TestGitHubAppInstallCallback_RedirectsWhenConfigured() {
	t := suite.T()

	installOp := suite.createImpersonationOperation("StartGitHubAppInstall", "Start GitHub App install flow")
	suite.registerRouteOnce(http.MethodPost, githubAppInstallPath, installOp, suite.h.StartGitHubAppInstallation)

	callbackOp := suite.createImpersonationOperation("HandleGitHubAppInstallCallback", "Handle GitHub App install callback")
	suite.registerRouteOnce(http.MethodGet, githubAppCallbackPath, callbackOp, suite.h.GitHubAppInstallCallback)

	mockGitHubAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := strings.TrimPrefix(req.URL.Path, "/api/v3")
		switch {
		case req.Method == http.MethodPost && path == "/app/installations/12345678/access_tokens":
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token":"ghs_test_installation_token","expires_at":"2030-01-01T00:00:00Z"}`))
		case req.Method == http.MethodPost && req.URL.Path == "/api/graphql":
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			_, _ = w.Write([]byte(`{"data":{"viewer":{"repositories":{"nodes":[{"nameWithOwner":"acme/demo","isPrivate":false,"updatedAt":"2030-01-01T00:00:00Z","url":"https://github.example/acme/demo"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer mockGitHubAPI.Close()

	privateKey := testRSAPrivateKeyPEM(t)

	cfg := githubapp.Config{
		APIURL:        mockGitHubAPI.URL,
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    privateKey,
		WebhookSecret: "secret",
	}

	restore := suite.withGitHubAppIntegrationRuntime(t, cfg, "https://console.openlane.io/integrations")
	defer restore()

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installReq := httptest.NewRequest(http.MethodPost, githubAppInstallPath, bytes.NewReader([]byte(`{}`)))
	installReq.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	installRec := httptest.NewRecorder()
	suite.e.ServeHTTP(installRec, installReq.WithContext(user.UserCtx))
	require.Equal(t, http.StatusOK, installRec.Code)

	var installResp openapi.GitHubAppInstallResponse
	require.NoError(t, json.Unmarshal(installRec.Body.Bytes(), &installResp))
	require.NotEmpty(t, installResp.State)

	cookies := cookieMap(installRec.Result().Cookies())
	require.Contains(t, cookies, "githubapp_state")

	callbackReq := httptest.NewRequest(http.MethodGet, githubAppCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("installation_id", "12345678")
	query.Set("state", installResp.State)
	callbackReq.URL.RawQuery = query.Encode()
	for _, name := range []string{"githubapp_state", "githubapp_org_id", "githubapp_user_id"} {
		callbackReq.AddCookie(cookies[name])
	}

	callbackRec := httptest.NewRecorder()
	suite.e.ServeHTTP(callbackRec, callbackReq.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusFound, callbackRec.Code)
	location := callbackRec.Header().Get("Location")
	assert.Contains(t, location, "provider=github_app")
	assert.Contains(t, location, "status=success")
}

// TestGitHubAppInstallCallback_VerifiesInstallationAgainstGitHubAPI verifies callback success requires a valid installation token + health call.
func (suite *HandlerTestSuite) TestGitHubAppInstallCallback_VerifiesInstallationAgainstGitHubAPI() {
	t := suite.T()

	installOp := suite.createImpersonationOperation("StartGitHubAppInstallWithHealth", "Start GitHub App install flow")
	suite.registerRouteOnce(http.MethodPost, githubAppInstallPath, installOp, suite.h.StartGitHubAppInstallation)

	callbackOp := suite.createImpersonationOperation("HandleGitHubAppInstallCallbackWithHealth", "Handle GitHub App install callback")
	suite.registerRouteOnce(http.MethodGet, githubAppCallbackPath, callbackOp, suite.h.GitHubAppInstallCallback)

	var (
		accessTokenCalls atomic.Int32
		repoLookupCalls  atomic.Int32
	)

	mockGitHubAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := strings.TrimPrefix(req.URL.Path, "/api/v3")
		switch {
		case req.Method == http.MethodPost && path == "/app/installations/12345678/access_tokens":
			accessTokenCalls.Add(1)
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token":"ghs_test_installation_token","expires_at":"2030-01-01T00:00:00Z"}`))
		case req.Method == http.MethodPost && req.URL.Path == "/api/graphql":
			repoLookupCalls.Add(1)
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			_, _ = w.Write([]byte(`{"data":{"viewer":{"repositories":{"nodes":[{"nameWithOwner":"acme/demo","isPrivate":false,"updatedAt":"2030-01-01T00:00:00Z","url":"https://github.example/acme/demo"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer mockGitHubAPI.Close()

	privateKey := testRSAPrivateKeyPEM(t)
	cfg := githubapp.Config{
		APIURL:        mockGitHubAPI.URL,
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    privateKey,
		WebhookSecret: "secret",
	}

	restoreRuntime := suite.withGitHubAppIntegrationRuntime(t, cfg, "https://console.openlane.io/integrations")
	defer restoreRuntime()

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installReq := httptest.NewRequest(http.MethodPost, githubAppInstallPath, bytes.NewReader([]byte(`{}`)))
	installReq.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	installRec := httptest.NewRecorder()
	suite.e.ServeHTTP(installRec, installReq.WithContext(user.UserCtx))
	require.Equal(t, http.StatusOK, installRec.Code)

	var installResp openapi.GitHubAppInstallResponse
	require.NoError(t, json.Unmarshal(installRec.Body.Bytes(), &installResp))
	require.NotEmpty(t, installResp.State)

	cookies := cookieMap(installRec.Result().Cookies())
	require.Contains(t, cookies, "githubapp_state")

	callbackReq := httptest.NewRequest(http.MethodGet, githubAppCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("installation_id", "12345678")
	query.Set("state", installResp.State)
	callbackReq.URL.RawQuery = query.Encode()
	for _, name := range []string{"githubapp_state", "githubapp_org_id", "githubapp_user_id"} {
		callbackReq.AddCookie(cookies[name])
	}

	callbackRec := httptest.NewRecorder()
	suite.e.ServeHTTP(callbackRec, callbackReq.WithContext(user.UserCtx))

	require.Equal(t, http.StatusFound, callbackRec.Code)
	location := callbackRec.Header().Get("Location")
	require.Contains(t, location, "provider=github_app")
	require.Contains(t, location, "status=success")
	require.Equal(t, int32(1), accessTokenCalls.Load())
	require.Equal(t, int32(1), repoLookupCalls.Load())

	integrationRecord, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(user.OrganizationID),
			integration.DefinitionIDEQ(githubAppDefinitionID),
		).
		Only(user.UserCtx)
	require.NoError(t, err)

	providerState, err := jsonx.ToMap(integrationRecord.ProviderState.ProviderData(githubAppSlug))
	require.NoError(t, err)
	require.Equal(t, "123", providerState["appId"])
	require.Equal(t, "12345678", providerState["installationId"])
}

func testRSAPrivateKeyPEM(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	encoded := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	require.NotEmpty(t, encoded)

	return string(encoded)
}
