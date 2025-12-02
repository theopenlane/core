package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/shared/integrations/config"
	"github.com/theopenlane/shared/integrations/types"
	models "github.com/theopenlane/shared/openapi"
)

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesSchemas() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProviders"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
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
		assert.NotNil(t, provider.CredentialsSchema, "expected schema for %s", provider.Name)
	}
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersMultipleProviders() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersMultiple"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
		types.ProviderType("github"): {
			Name:        "github",
			DisplayName: "GitHub",
			Category:    "code",
			AuthType:    types.AuthKindOAuth2,
			Active:      true,
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
	for _, provider := range resp.Providers {
		providerNames[provider.Name] = true
	}
	assert.True(t, providerNames["gcp_scc"])
	assert.True(t, providerNames["github"])
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersSingleProvider() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersSingle"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
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
		if provider.Name == "gcp_scc" {
			found = true
			break
		}
	}
	require.True(t, found, "expected gcp_scc provider")
}

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesActiveStatus() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProvidersActiveStatus"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	activeSpec := gcpSCCSpec()
	activeSpec.Active = true

	inactiveSpec := gcpSCCSpec()
	inactiveSpec.Name = "inactive_provider"
	inactiveSpec.DisplayName = "Inactive Provider"
	inactiveSpec.Active = false

	specs := map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"):           activeSpec,
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

	require.Contains(t, providersByName, "gcp_scc")
	require.NotContains(t, providersByName, "inactive_provider")

	assert.True(t, providersByName["gcp_scc"].Active)
}
