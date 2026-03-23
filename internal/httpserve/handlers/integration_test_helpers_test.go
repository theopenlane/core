package handlers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
)

var githubAppDefinitionID = githubapp.DefinitionID.ID()
var githubTestCredentialRef = types.NewCredentialSlotID("github_test")

// withDefinitionRuntime swaps the suite handler's IntegrationsRuntime for a new one
// built from the given definition builders. Returns a restore function.
func (suite *HandlerTestSuite) withDefinitionRuntime(t *testing.T, builders []registry.Builder) func() {
	t.Helper()

	original := suite.h.IntegrationsRuntime
	originalConfig := suite.h.IntegrationsConfig

	galaInstance, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
		Enabled:      true,
	})
	assert.NoError(t, err)

	credStore, err := keystore.NewStore(suite.db)
	assert.NoError(t, err)

	rt, err := runtime.New(runtime.Config{
		DB:                    suite.db,
		Gala:                  galaInstance,
		Keystore:              credStore,
		DefinitionBuilders:    builders,
		SkipExecutorListeners: true,
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

// gcpSCCTestDefinitionBuilder returns a test definition for GCP SCC-style credential config tests.
// The definition uses definitionID as Spec.ID so registry lookups work correctly.
func gcpSCCTestDefinitionBuilder(definitionID string) registry.Builder {
	schema, err := json.Marshal(map[string]any{
		"type": "object",
		"required": []string{
			"projectId",
			"serviceAccountEmail",
		},
		"properties": map[string]any{
			"projectId": map[string]any{
				"type":        "string",
				"title":       "Project ID",
				"description": "GCP project identifier",
			},
			"serviceAccountEmail": map[string]any{
				"type":        "string",
				"title":       "Service Account Email",
				"description": "Workload identity service account",
			},
			"organizationId": map[string]any{
				"type":        "string",
				"title":       "Organization ID",
				"description": "Optional organization scope",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	userInputSchema, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"filterExpr": map[string]any{
				"type":  "string",
				"title": "Filter Expression",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return registry.Builder(func() (types.Definition, error) {
		gcpSCCTestCredential := types.NewCredentialSlotID("gcp_scc_test")

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Google Cloud SCC",
				Description: "Google Cloud Security Command Center integration",
				Category:    "cloud",
				Active:      true,
				Visible:     true,
				Tags:        []string{"cloud", "google"},
			},
			UserInput: &types.UserInputRegistration{
				Schema: json.RawMessage(userInputSchema),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         gcpSCCTestCredential,
					Name:        "GCP SCC Test Credential",
					Description: "Credential slot used by the GCP SCC test definition.",
					Schema:      json.RawMessage(schema),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       gcpSCCTestCredential,
					Name:                "GCP SCC Test Connection",
					Description:         "Connect the GCP SCC test definition using the configured credential payload.",
					CredentialRefs:      []types.CredentialSlotID{gcpSCCTestCredential},
					ValidationOperation: "health.default",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: gcpSCCTestCredential,
						Name:          "Disconnect GCP SCC Test Connection",
						Description:   "Remove the persisted GCP SCC test credential and disconnect this installation.",
					},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Description: "Validate the test credential payload",
					Topic:       gala.TopicName("integration." + definitionID + ".health.default"),
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				},
			},
		}, nil
	})
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
						Name:          "Disconnect GitHub Test Connection",
						Description:   "Remove the persisted GitHub test credential and disconnect this installation.",
					},
				},
			},
		}, nil
	})
}
