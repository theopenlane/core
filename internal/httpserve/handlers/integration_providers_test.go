package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	models "github.com/theopenlane/core/common/openapi"
)

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesSchemas() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProviders"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	}
	restore := suite.withIntegrationRegistry(t, specs)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.IntegrationProvidersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	providersByName := make(map[string]models.IntegrationProviderMetadata)
	for _, provider := range resp.Providers {
		providersByName[provider.Name] = provider
	}

	for name, spec := range specs {
		provider, ok := providersByName[string(name)]
		require.True(t, ok, "expected provider %s", name)
		assert.Equal(t, spec.DisplayName, provider.DisplayName)
		assert.Equal(t, spec.Description, provider.Description)
		assert.Equal(t, lo.FromPtr(spec.Visible), provider.Visible)
		assert.ElementsMatch(t, spec.Tags, provider.Tags)
		assert.NotNil(t, provider.CredentialsSchema, "expected schema for %s", provider.Name)
	}
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersMultipleProviders() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersMultiple"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
		types.ProviderType("github"): {
			Name:             "github",
			DisplayName:      "GitHub",
			Category:         "code",
			Description:      "GitHub integration",
			AuthType:         types.AuthKindOAuth2,
			AuthStartPath:    "/v1/integrations/oauth/start",
			AuthCallbackPath: "/v1/integrations/oauth/callback",
			Active:           lo.ToPtr(true),
			Visible:          lo.ToPtr(true),
			Tags:             []string{"code", "github"},
			OAuth: &config.OAuthSpec{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				AuthURL:      "https://github.com/login/oauth/authorize",
				TokenURL:     "https://github.com/login/oauth/access_token",
				Scopes:       []string{"repo", "user"},
				RedirectURI:  "https://example.com/callback",
			},
		},
	}
	restore := suite.withIntegrationRegistry(t, specs)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.IntegrationProvidersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	providerNames := make(map[string]bool)
	providersByName := make(map[string]models.IntegrationProviderMetadata)
	for _, provider := range resp.Providers {
		providerNames[provider.Name] = true
		providersByName[provider.Name] = provider
	}
	assert.True(t, providerNames["gcpscc"])
	assert.True(t, providerNames["github"])
	assert.Equal(t, "/v1/integrations/oauth/start", providersByName["github"].AuthStartPath)
	assert.Equal(t, "/v1/integrations/oauth/callback", providersByName["github"].AuthCallbackPath)
	assert.Equal(t, "GitHub integration", providersByName["github"].Description)
	assert.Equal(t, []string{"code", "github"}, providersByName["github"].Tags)
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersSingleProvider() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersSingle"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	}
	restore := suite.withIntegrationRegistry(t, specs)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.IntegrationProvidersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	found := false
	for _, provider := range resp.Providers {
		if provider.Name == "gcpscc" {
			found = true
			break
		}
	}
	require.True(t, found, "expected gcpscc provider")
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesActiveStatus() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersActiveStatus"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	activeSpec := gcpSCCSpec()
	activeSpec.Active = lo.ToPtr(true)

	inactiveSpec := gcpSCCSpec()
	inactiveSpec.Name = "inactive_provider"
	inactiveSpec.DisplayName = "Inactive Provider"
	inactiveSpec.Active = lo.ToPtr(false)
	inactiveSpec.Visible = lo.ToPtr(true)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcpscc"):            activeSpec,
		types.ProviderType("inactive_provider"): inactiveSpec,
	}
	restore := suite.withIntegrationRegistry(t, specs)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.IntegrationProvidersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	providersByName := make(map[string]models.IntegrationProviderMetadata)
	for _, provider := range resp.Providers {
		providersByName[provider.Name] = provider
	}

	require.Contains(t, providersByName, "gcpscc")
	require.Contains(t, providersByName, "inactive_provider")

	assert.True(t, providersByName["gcpscc"].Active)
	assert.True(t, providersByName["gcpscc"].Visible)
	assert.False(t, providersByName["inactive_provider"].Active)
	assert.True(t, providersByName["inactive_provider"].Visible)
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesGitHubAppInstallMetadata() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersGitHubAppMetadata"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("githubapp"): {
			Name:             "githubapp",
			DisplayName:      "GitHub App",
			Category:         "code",
			Description:      "GitHub App integration",
			AuthType:         types.AuthKindGitHubApp,
			AuthStartPath:    "/v1/integrations/github/app/install",
			AuthCallbackPath: "/v1/integrations/github/app/callback",
			Active:           lo.ToPtr(true),
			Visible:          lo.ToPtr(true),
			GitHubApp: &config.GitHubAppSpec{
				BaseURL: "https://api.github.com",
				AppSlug: "openlane-test-app",
			},
		},
	}
	restore := suite.withIntegrationRegistry(t, specs)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.IntegrationProvidersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	var githubAppProvider *models.IntegrationProviderMetadata
	for i := range resp.Providers {
		if resp.Providers[i].Name == "githubapp" {
			githubAppProvider = &resp.Providers[i]
			break
		}
	}
	require.NotNil(t, githubAppProvider, "expected githubapp provider in response")
	require.NotNil(t, githubAppProvider.GitHubApp, "expected githubApp metadata")
	assert.Equal(t, "openlane-test-app", githubAppProvider.GitHubApp.AppSlug)
	assert.Equal(t, "/v1/integrations/github/app/install", githubAppProvider.AuthStartPath)
	assert.Equal(t, "/v1/integrations/github/app/callback", githubAppProvider.AuthCallbackPath)
	assert.Equal(t, "GitHub App integration", githubAppProvider.Description)
}
