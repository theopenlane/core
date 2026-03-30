//go:build test

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesSchemas() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProviders"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{
		configTestDefinitionBuilder(configTestProviderID, false),
		configTestDefinitionBuilder("def_01K0TESTOTH00000000000001", false),
	})
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationProvidersResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Len(t, resp.Providers, 2)

	providers := map[string]types.Definition{}
	for _, provider := range resp.Providers {
		providers[provider.ID] = provider
	}

	provider, ok := providers[configTestProviderID]
	assert.True(t, ok)
	assert.True(t, provider.Active)
	assert.True(t, provider.Visible)
	assert.Len(t, provider.CredentialRegistrations, 1)
	assert.NotNil(t, provider.CredentialRegistrations[0].Schema)
	assert.NotNil(t, provider.UserInput)
	assert.Len(t, provider.Operations, 1)
	assert.Equal(t, configHealthCheckOperation.Name(), provider.Operations[0].Name)

	_, ok = providers["def_01K0TESTOTH00000000000001"]
	assert.True(t, ok)
}
