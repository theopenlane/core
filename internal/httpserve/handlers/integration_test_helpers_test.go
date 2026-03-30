package handlers_test

import (
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HelperTestHealthCheck is the config type for the helper test health check operation
type HelperTestHealthCheck struct{}

var (
	githubAppDefinitionID   = githubapp.DefinitionID.ID()
	githubTestCredentialRef = types.NewCredentialSlotID("github_test")
	_, _                    = providerkit.OperationSchema[HelperTestHealthCheck]()
)

// withDefinitionRuntime returns a restore function that resets IntegrationsConfig.
// All definitions and gala listeners are registered once in SetupSuite
func (suite *HandlerTestSuite) withDefinitionRuntime(_ *testing.T, _ []registry.Builder) func() {
	originalConfig := suite.h.IntegrationsConfig

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// withGitHubAppIntegrationRuntime sets the handler's GitHubApp config for the test
// and returns a restore function that resets it
func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config) func() {
	t.Helper()

	originalConfig := suite.h.IntegrationsConfig
	suite.h.IntegrationsConfig.GitHubApp = cfg

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
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
