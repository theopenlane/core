package handlers_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
)

// HelperTestHealthCheck is the config type for the helper test health check operation
type HelperTestHealthCheck struct{}

var (
	githubAppDefinitionID   = githubapp.DefinitionID.ID()
	githubTestCredentialRef = types.NewCredentialSlotID("github_test")
	_, _                    = providerkit.OperationSchema[HelperTestHealthCheck]()
)

// withDefinitionRuntime swaps the suite handler's IntegrationsRuntime for a new one
// built from the given definition builders, backed by the suite's durable Gala instance.
// Returns a restore function
func (suite *HandlerTestSuite) withDefinitionRuntime(t *testing.T, builders []registry.Builder) func() {
	t.Helper()

	original := suite.h.IntegrationsRuntime
	originalConfig := suite.h.IntegrationsConfig

	credStore, err := keystore.NewStore(suite.db)
	assert.NoError(t, err)

	rt, err := runtime.New(runtime.Config{
		DB:                 suite.db,
		Gala:               suite.galaRuntime,
		Keystore:           credStore,
		DefinitionBuilders: builders,
	})
	assert.NoError(t, err)

	suite.h.IntegrationsRuntime = rt

	return func() {
		suite.h.IntegrationsRuntime = original
		suite.h.IntegrationsConfig = originalConfig
	}
}

// withGitHubAppIntegrationRuntime swaps the suite handler's IntegrationsRuntime for one
// configured with the GitHub App definition built from cfg. Returns a restore function.
func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config) func() {
	t.Helper()
	restore := suite.withDefinitionRuntime(t, []registry.Builder{githubapp.Builder(cfg)})
	suite.h.IntegrationsConfig.GitHubApp = cfg

	return restore
}

// githubTestDefinitionBuilder returns a minimal test definition used for disconnect tests.
// The definition has no credentials schema or auth flow; it only needs to be present in
// the registry so the handler can resolve the provider by ID.
func githubTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "GitHub",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         githubTestCredentialRef,
					Name:        "GitHub Test Credential",
					Description: "Credential slot used by the GitHub disconnect test definition.",
					Schema:      json.RawMessage(`{"type":"object","properties":{"token":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  githubTestCredentialRef,
					Name:           "GitHub Test Connection",
					Description:    "Test connection used for handler disconnect flows.",
					CredentialRefs: []types.CredentialSlotID{githubTestCredentialRef},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: githubTestCredentialRef,
						Description:   "Remove the persisted GitHub test credential and disconnect this installation.",
					},
				},
			},
		}, nil
	})
}
