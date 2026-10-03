package handlers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// githubTestCredential is the credential type stored by the GitHub disconnect test definition
type githubTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

var (
	githubAppDefinitionID   = githubapp.DefinitionID.ID()
	githubTestCredentialRef = types.CredentialRefOf[githubTestCredential]()
)

// withDefinitionRuntime returns a restore function that resets IntegrationsConfig
func (suite *HandlerTestSuite) withDefinitionRuntime(_ *testing.T, _ []registry.Builder) func() {
	originalConfig := suite.h.IntegrationsConfig

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// withGitHubAppIntegrationRuntime sets the handler's GitHubApp config for the test and returns a restore function that resets it
func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config) func() {
	t.Helper()

	originalConfig := suite.h.IntegrationsConfig
	suite.h.IntegrationsConfig.GitHubApp = cfg

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// githubTestDefinitionBuilder returns a minimal test definition used for disconnect tests
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
				githubTestCredentialRef.Registration(types.CredentialRegistration{
					Name:        "GitHub Test Credential",
					Description: "Credential slot used by the GitHub disconnect test definition.",
				}),
			},
			HealthCheck: types.CredentialHealthCheck(func(context.Context, types.OperationRequest) (json.RawMessage, error) {
				return json.RawMessage(`{"ok":true}`), nil
			}),
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  githubTestCredentialRef.ID(),
					Name:           "GitHub Test Connection",
					Description:    "Test connection used for handler disconnect flows.",
					CredentialRefs: []types.CredentialSlotID{githubTestCredentialRef.ID()},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: githubTestCredentialRef.ID(),
						Description:   "Remove the persisted GitHub test credential and disconnect this installation.",
					},
				},
			},
		}, nil
	})
}
