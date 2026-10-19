package runtime

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// seedTestCredential is the credential schema behind the client-bound scheduled operation fixture
type seedTestCredential struct {
	Token string `json:"token"`
}

// seedTestMetadata is the installation metadata the fixture connection verifies to
type seedTestMetadata struct {
	Tenant string `json:"tenant"`
}

// seedTestClient is the client the fixture connection provides
type seedTestClient struct{}

func TestSeedScheduledOperationsSkipsClientBoundWithoutRuntime(t *testing.T) {
	t.Parallel()

	connection := types.ConnectionOf[seedTestCredential]().
		Provides(func(context.Context, types.ConnectionRequest[seedTestCredential]) (*seedTestClient, error) {
			return &seedTestClient{}, nil
		}).
		Verified(func(context.Context, types.ConnectionRequest[seedTestCredential], *seedTestClient) (seedTestMetadata, error) {
			return seedTestMetadata{}, nil
		})

	reg := registry.New()
	assert.NilError(t, reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "def_scheduled_client", DisplayName: "Scheduled Client", Active: true},
		Connections:    []types.Connector{connection},
		Installation:   types.InstallationOf[seedTestMetadata]().Registration(),
		Operations: []types.OperationRegistration{{
			Name:      "import",
			ClientRef: connection.Connection().Verify.ClientRef,
			Policy:    types.ExecutionPolicy{Scheduled: true},
			Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
				return nil, nil
			},
		}},
	}))

	// the test runtime carries no gala, so reaching the emit would dereference nil
	assert.NilError(t, NewForTesting(reg).SeedScheduledOperations(context.Background()))
}
