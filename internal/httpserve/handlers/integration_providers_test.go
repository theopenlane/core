//go:build test

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestListIntegrationProvidersIncludesSchemas(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.Check(t, is.Equal(http.StatusOK, rec.Code))

	var resp struct {
		Success   bool                           `json:"success"`
		Providers []handlers.IntegrationProvider `json:"providers"`
	}
	assert.Check(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Check(t, resp.Success)
	assert.Check(t, len(resp.Providers) >= 2)

	providers := map[string]handlers.IntegrationProvider{}
	for _, provider := range resp.Providers {
		providers[provider.Spec.ID] = provider
	}

	provider, ok := providers[configTestProviderID]
	assert.Check(t, ok)
	assert.Check(t, provider.Spec.Active)
	assert.Check(t, provider.Spec.Visible)
	assert.Check(t, is.Len(provider.CredentialRegistrations, 1))
	assert.Check(t, is.DeepEqual(configTestCredentialRef, provider.CredentialRegistrations[0].Ref))
	assert.Check(t, len(provider.CredentialRegistrations[0].Schema) > 0)
	assert.Check(t, is.Len(provider.Connections, 1))
	assert.Check(t, is.DeepEqual(configTestCredentialRef, provider.Connections[0].CredentialRef))
	assert.Check(t, is.DeepEqual([]string{configTestCredentialRef}, provider.Connections[0].CredentialRefs))
	assert.Check(t, is.Nil(provider.Connections[0].Auth))
	assert.Check(t, provider.Connections[0].Disconnect != nil)
	assert.Check(t, provider.UserInput != nil)
	assert.Check(t, is.Len(provider.Operations, 0))

	_, ok = providers["def_01K0TESTOTH00000000000001"]
	assert.Check(t, ok)

	_, ok = providers["def_01K0TESTHID00000000000001"]
	assert.Check(t, !ok)
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
