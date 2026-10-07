//go:build test

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func (suite *HandlerTestSuite) TestListIntegrationProvidersIncludesSchemas() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodGet, "/v1/integrations/providers", suite.h.ListIntegrationProviders)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{
		configTestDefinitionBuilder(configTestProviderID, false),
		configTestDefinitionBuilder("def_01K0TESTOTH00000000000001", false),
		hiddenDefinitionBuilder("def_01K0TESTHID00000000000001"),
	})
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/providers", nil)
	req = req.WithContext(echocontext.NewTestEchoContext().Request().Context())
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Success   bool                           `json:"success"`
		Providers []handlers.IntegrationProvider `json:"providers"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.GreaterOrEqual(t, len(resp.Providers), 2)

	providers := map[string]handlers.IntegrationProvider{}
	for _, provider := range resp.Providers {
		providers[provider.Spec.ID] = provider
	}

	provider, ok := providers[configTestProviderID]
	assert.True(t, ok)
	assert.True(t, provider.Spec.Active)
	assert.True(t, provider.Spec.Visible)
	assert.Len(t, provider.CredentialRegistrations, 1)
	assert.Equal(t, configTestCredentialRef, provider.CredentialRegistrations[0].Ref)
	assert.NotEmpty(t, provider.CredentialRegistrations[0].Schema)
	assert.Len(t, provider.Connections, 1)
	assert.Equal(t, configTestCredentialRef, provider.Connections[0].CredentialRef)
	assert.Equal(t, []string{configTestCredentialRef}, provider.Connections[0].CredentialRefs)
	assert.Nil(t, provider.Connections[0].Auth)
	assert.NotNil(t, provider.Connections[0].Disconnect)
	assert.NotNil(t, provider.UserInput)
	assert.Empty(t, provider.Operations)

	_, ok = providers["def_01K0TESTOTH00000000000001"]
	assert.True(t, ok)

	_, ok = providers["def_01K0TESTHID00000000000001"]
	assert.False(t, ok)
}

// hiddenDefinitionBuilder returns a builder for a definition that is active but not visible in catalog surfaces
func hiddenDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def, err := configTestDefinitionBuilder(definitionID, false)()
		if err != nil {
			return types.Definition{}, err
		}

		def.Visible = false

		return def, nil
	})
}
