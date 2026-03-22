//go:build test

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/types"
)

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesSchemas() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ListIntegrationProviders"
	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", op, suite.h.ListIntegrationProviders)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{
		configTestDefinitionBuilder(configTestProviderID, "gcpscc", false),
		configTestDefinitionBuilder("def_test_other", "other", false),
	})
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationProvidersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	require.Len(t, resp.Providers, 2)

	providers := map[string]types.Definition{}
	for _, provider := range resp.Providers {
		providers[provider.Slug] = provider
	}

	gcpscc, ok := providers["gcpscc"]
	require.True(t, ok)
	assert.Equal(t, configTestProviderID, gcpscc.ID)
	assert.True(t, gcpscc.Active)
	assert.True(t, gcpscc.Visible)
	assert.Len(t, gcpscc.CredentialRegistrations, 1)
	assert.NotNil(t, gcpscc.CredentialRegistrations[0].Schema)
	assert.NotNil(t, gcpscc.UserInput)
	assert.Len(t, gcpscc.Operations, 1)
	assert.Equal(t, "health.default", gcpscc.Operations[0].Name)

	other, ok := providers["other"]
	require.True(t, ok)
	assert.Equal(t, "def_test_other", other.ID)
}
