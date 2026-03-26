package handlers_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
)

// HelperTestHealthCheck is the config type for the helper test health check operation
type HelperTestHealthCheck struct{}

var (
	githubAppDefinitionID   = githubapp.DefinitionID.ID()
	githubTestCredentialRef = types.NewCredentialSlotID("github_test")
	_, _                    = providerkit.OperationSchema[HelperTestHealthCheck]()
)

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
		DB:                 suite.db,
		Gala:               galaInstance,
		Keystore:           credStore,
		DefinitionBuilders: builders,
	})
	assert.NoError(t, err)

	suite.h.IntegrationsRuntime = rt

	return func() {
		_ = galaInstance.Close()
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

// withDurableGitHubAppIntegrationRuntime swaps the suite handler's IntegrationsRuntime for one
// backed by a durable Gala instance with River workers, matching production dispatch behavior
func (suite *HandlerTestSuite) withDurableGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config) func() {
	t.Helper()

	original := suite.h.IntegrationsRuntime
	originalConfig := suite.h.IntegrationsConfig

	galaInstance, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode:      gala.DispatchModeDurable,
		ConnectionURI:     suite.tf.URI,
		QueueName:         "github_webhook_test",
		WorkerCount:       5,
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: 10 * time.Millisecond,
	})
	assert.NoError(t, err)

	do.ProvideValue(galaInstance.Injector(), suite.db)

	credStore, err := keystore.NewStore(suite.db)
	assert.NoError(t, err)

	rt, err := runtime.New(runtime.Config{
		DB:                 suite.db,
		Gala:               galaInstance,
		Keystore:           credStore,
		DefinitionBuilders: []registry.Builder{githubapp.Builder(cfg)},
	})
	assert.NoError(t, err)

	assert.NoError(t, galaInstance.StartWorkers(context.Background()))

	suite.h.IntegrationsRuntime = rt
	suite.h.IntegrationsConfig.GitHubApp = cfg

	return func() {
		_ = galaInstance.StopWorkers(context.Background())
		_ = galaInstance.Close()
		suite.h.IntegrationsRuntime = original
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
