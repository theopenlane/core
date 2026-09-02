//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hush"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	systemdef "github.com/theopenlane/core/v2/internal/integrations/definitions/system"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// runLifecycleSweep executes the lifecycle sweep operation inline through the real runtime
// and returns the processed count from the cycle result
func runLifecycleSweep(t *testing.T, ctx context.Context, config json.RawMessage) int {
	t.Helper()

	raw, err := suite.IntegrationsRT.ExecuteRuntimeOperation(ctx, systemdef.DefinitionID.ID(), systemdef.IntegrationLifecycleOp.Name(), config)
	require.NoError(t, err)

	var result integrationtypes.ScheduledCycleResult
	require.NoError(t, json.Unmarshal(raw, &result))

	return result.Processed
}

// integrationVisible reports whether the installation row is still readable
func integrationVisible(t *testing.T, ctx context.Context, id string) bool {
	t.Helper()

	_, err := suite.Client.DB.Integration.Get(ctx, id)
	if ent.IsNotFound(err) {
		return false
	}

	require.NoError(t, err)

	return true
}

// installationCredentialIDs lists the hush credential row IDs attached to the installation
func installationCredentialIDs(t *testing.T, ctx context.Context, integrationID string) []string {
	t.Helper()

	installation, err := suite.Client.DB.Integration.Get(ctx, integrationID)
	require.NoError(t, err)

	ids, err := installation.QuerySecrets().IDs(ctx)
	require.NoError(t, err)

	return ids
}

// TestIntegrationLifecycleSweep drives the sweep through the real runtime: expired pending
// are reaped with their credentials, every other state survives
func TestIntegrationLifecycleSweep(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)

	allowCtx := privacy.DecisionContext(th.SetContext(org.UserCtx, suite.Client.DB), privacy.Allow)
	ownerCtx := th.SetContext(org.UserCtx, suite.Client.DB)

	expiredPending, _ := newHarnessInstallation(t, allowCtx, testint.ModeRecurring)
	expiredCredentialIDs := installationCredentialIDs(t, allowCtx, expiredPending.ID)
	require.NotEmpty(t, expiredCredentialIDs)
	require.NoError(t, suite.Client.DB.Integration.UpdateOneID(expiredPending.ID).
		SetStatus(enums.IntegrationStatusPending).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Exec(allowCtx))

	freshPending, err := suite.Client.DB.Integration.Create().
		SetName(th.RandomName(t)).
		SetKind("testintegration").
		SetDefinitionID(testint.DefinitionID.ID()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		Save(allowCtx)
	require.NoError(t, err)

	connected, connectedFragment := seedHarnessLoop(t, allowCtx)

	errored, _ := newHarnessInstallation(t, allowCtx, testint.ModeRecurring)
	require.NoError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(allowCtx, errored, "credentials revoked"))

	waitForEvents()

	t.Run("dry run dispatches nothing", func(t *testing.T) {
		processed := runLifecycleSweep(t, allowCtx, json.RawMessage(`{"dryRun":true}`))
		require.GreaterOrEqual(t, processed, 1)

		require.True(t, integrationVisible(t, allowCtx, expiredPending.ID))
	})

	t.Run("sweep reaps expired pending only", func(t *testing.T) {
		processed := runLifecycleSweep(t, allowCtx, nil)
		require.GreaterOrEqual(t, processed, 1)

		waitForEvents()

		require.False(t, integrationVisible(t, allowCtx, expiredPending.ID))

		credentialCount, err := suite.Client.DB.Hush.Query().
			Where(hush.IDIn(expiredCredentialIDs...)).
			Count(allowCtx)
		require.NoError(t, err)
		require.Zero(t, credentialCount)

		require.True(t, integrationVisible(t, allowCtx, freshPending.ID))
		require.Equal(t, enums.IntegrationStatusPending, reloadIntegration(t, allowCtx, freshPending.ID).Status)

		require.Equal(t, enums.IntegrationStatusConnected, reloadIntegration(t, allowCtx, connected.ID).Status)
		require.Equal(t, 1, activeReconcileJobs(t, connectedFragment))

		require.Equal(t, enums.IntegrationStatusErrored, reloadIntegration(t, allowCtx, errored.ID).Status)
		require.Equal(t, 0, integrationNotificationCount(t, ownerCtx, errored.OwnerID, integrationReconnectedObjectType))
	})
}
