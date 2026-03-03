package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationconfig "github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/activation"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
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

	originalConfig := suite.h.IntegrationGitHubApp
	suite.h.IntegrationGitHubApp = handlers.IntegrationGitHubAppConfig{
		Enabled:            true,
		AppID:              "123",
		AppSlug:            "openlane",
		PrivateKey:         "private-key",
		WebhookSecret:      "secret",
		SuccessRedirectURL: "https://console.openlane.io/integrations",
	}
	defer func() {
		suite.h.IntegrationGitHubApp = originalConfig
	}()

	originalRegistry := suite.h.IntegrationRegistry
	suite.h.IntegrationRegistry = noHealthIntegrationRegistry{base: originalRegistry}
	defer func() {
		suite.h.IntegrationRegistry = originalRegistry
	}()

	originalOps := suite.h.IntegrationOperations
	ops, err := keystore.NewOperationManager(suite.h.IntegrationBroker, nil)
	require.NoError(t, err)
	suite.h.IntegrationOperations = ops
	defer func() {
		suite.h.IntegrationOperations = originalOps
	}()

	originalActivation := suite.h.IntegrationActivation
	store := keystore.NewStore(suite.db)
	sessions := keymaker.NewMemorySessionStore()
	svc, err := keymaker.NewService(suite.h.IntegrationRegistry, store, sessions, keymaker.ServiceOptions{})
	require.NoError(t, err)
	activationSvc, err := activation.NewService(svc, store, &mockOperationRunner{}, &mockPayloadMinter{})
	require.NoError(t, err)
	suite.h.IntegrationActivation = activationSvc
	defer func() {
		suite.h.IntegrationActivation = originalActivation
	}()

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

	callbackReq.AddCookie(cookies["githubapp_state"])

	callbackRec := httptest.NewRecorder()
	suite.e.ServeHTTP(callbackRec, callbackReq.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusFound, callbackRec.Code)
	location := callbackRec.Header().Get("Location")
	assert.Contains(t, location, "provider=githubapp")
	assert.Contains(t, location, "status=success")
}

// noHealthIntegrationRegistry suppresses health operation descriptors for GitHub App in tests.
type noHealthIntegrationRegistry struct {
	base handlers.ProviderRegistry
}

// Provider returns the provider implementation for the given type.
func (r noHealthIntegrationRegistry) Provider(provider types.ProviderType) (types.Provider, bool) {
	if r.base == nil {
		return nil, false
	}
	return r.base.Provider(provider)
}

// Config returns the provider config for the given type.
func (r noHealthIntegrationRegistry) Config(provider types.ProviderType) (integrationconfig.ProviderSpec, bool) {
	if r.base == nil {
		return integrationconfig.ProviderSpec{}, false
	}
	return r.base.Config(provider)
}

// ProviderMetadataCatalog returns the provider metadata catalog.
func (r noHealthIntegrationRegistry) ProviderMetadataCatalog() map[types.ProviderType]types.ProviderConfig {
	if r.base == nil {
		return nil
	}
	return r.base.ProviderMetadataCatalog()
}

// OperationDescriptors returns operation descriptors, skipping GitHub App health checks.
func (r noHealthIntegrationRegistry) OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor {
	if provider == github.TypeGitHubApp {
		return nil
	}
	if r.base == nil {
		return nil
	}
	return r.base.OperationDescriptors(provider)
}
