package handlers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	v2definition "github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	IntegrationsRuntime "github.com/theopenlane/core/internal/integrations/runtime"
	v2types "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
)

const githubAppSlug = githubapp.Slug

var githubAppDefinitionID = githubapp.DefinitionID.ID()

// withDefinitionRuntime swaps the suite handler's IntegrationsRuntime for a new one
// built from the given definition builders. Returns a restore function.
func (suite *HandlerTestSuite) withDefinitionRuntime(t *testing.T, builders []v2definition.Builder, successRedirectURL string) func() {
	t.Helper()

	original := suite.h.IntegrationsRuntime

	galaInstance, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
		Enabled:      true,
	})
	assert.NoError(t, err)

	credStore, err := keystore.NewStore(suite.db)
	assert.NoError(t, err)

	rt, err := IntegrationsRuntime.New(IntegrationsRuntime.Config{
		DB:                    suite.db,
		Gala:                  galaInstance,
		Keystore:              credStore,
		AuthStateStore:        keymaker.NewInMemoryAuthStateStore(),
		DefinitionBuilders:    builders,
		SkipExecutorListeners: true,
		SuccessRedirectURL:    successRedirectURL,
	})
	assert.NoError(t, err)

	suite.h.IntegrationsRuntime = rt

	return func() {
		suite.h.IntegrationsRuntime = original
	}
}

// withGitHubAppIntegrationRuntime swaps the suite handler's IntegrationsRuntime for one
// configured with the GitHub App definition built from cfg. Returns a restore function.
func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config, successRedirectURL string) func() {
	t.Helper()
	return suite.withDefinitionRuntime(t, []v2definition.Builder{githubapp.Builder(cfg)}, successRedirectURL)
}

// gcpSCCTestDefinitionBuilder returns a test definition for GCP SCC-style credential config tests.
// The definition uses definitionID as both Spec.ID and Spec.Slug so registry lookups and DB
// queries using the same string work correctly.
func gcpSCCTestDefinitionBuilder(definitionID string) v2definition.Builder {
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
			"label": map[string]any{
				"type":  "string",
				"title": "Installation Label",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return v2definition.Builder(func(_ context.Context) (v2types.Definition, error) {
		return v2types.Definition{
			DefinitionSpec: v2types.DefinitionSpec{
				ID:          definitionID,
				Slug:        definitionID,
				Version:     "v1",
				DisplayName: "Google Cloud SCC",
				Description: "Google Cloud Security Command Center integration",
				Category:    "cloud",
				Active:      true,
				Visible:     true,
				Tags:        []string{"cloud", "google"},
			},
			UserInput: &v2types.UserInputRegistration{
				Schema: json.RawMessage(userInputSchema),
			},
			Credentials: &v2types.CredentialRegistration{
				Schema: json.RawMessage(schema),
			},
			Operations: []v2types.OperationRegistration{
				{
					Name:        "health.default",
					Description: "Validate the test credential payload",
					Topic:       gala.TopicName("integration." + definitionID + ".health.default"),
					Handle: func(context.Context, *ent.Integration, v2types.CredentialSet, any, json.RawMessage) (json.RawMessage, error) {
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
func githubTestDefinitionBuilder(definitionID string) v2definition.Builder {
	return v2definition.Builder(func(_ context.Context) (v2types.Definition, error) {
		return v2types.Definition{
			DefinitionSpec: v2types.DefinitionSpec{
				ID:          definitionID,
				Slug:        definitionID,
				Version:     "v1",
				DisplayName: "GitHub",
				Active:      true,
				Visible:     true,
			},
		}, nil
	})
}
