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

// TestIntegrationLifecycleSweep drives the sweep through the real runtime: expired
// never-connected installations are reaped with their credentials, every other state survives
func TestIntegrationLifecycleSweep(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)

	internalCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)
	ownerCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	expiredPending, _ := newHarnessInstallation(t, internalCtx, testint.ModeRecurring)
	expiredCredentialIDs := installationCredentialIDs(t, internalCtx, expiredPending.ID)
	require.NotEmpty(t, expiredCredentialIDs)
	require.NoError(t, suite.Client.DB.Integration.UpdateOneID(expiredPending.ID).
		SetStatus(enums.IntegrationStatusPending).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Exec(internalCtx))

	freshPending, err := suite.Client.DB.Integration.Create().
		SetName(th.RandomName(t)).
		SetKind("testintegration").
		SetDefinitionID(testint.DefinitionID.ID()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		Save(internalCtx)
	require.NoError(t, err)

	connected, connectedFragment := seedHarnessLoop(t, internalCtx)

	errored, _ := newHarnessInstallation(t, internalCtx, testint.ModeRecurring)
	require.NoError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(internalCtx, errored, "credentials revoked"))

	erroredStray, _ := newHarnessInstallation(t, internalCtx, testint.ModeRecurring)
	strayCredentialIDs := installationCredentialIDs(t, internalCtx, erroredStray.ID)
	require.NotEmpty(t, strayCredentialIDs)
	require.NoError(t, suite.Client.DB.Integration.UpdateOneID(erroredStray.ID).
		SetStatus(enums.IntegrationStatusPending).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Exec(internalCtx))
	require.NoError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(internalCtx, reloadIntegration(t, internalCtx, erroredStray.ID), "connection check failed"))

	flippedStray := reloadIntegration(t, internalCtx, erroredStray.ID)
	require.Equal(t, enums.IntegrationStatusErrored, flippedStray.Status)
	require.NotNil(t, flippedStray.ExpiresAt)

	waitForEvents()

	t.Run("dry run dispatches nothing", func(t *testing.T) {
		processed := runLifecycleSweep(t, internalCtx, json.RawMessage(`{"dryRun":true}`))
		require.GreaterOrEqual(t, processed, 2)

		require.True(t, integrationVisible(t, internalCtx, expiredPending.ID))
		require.True(t, integrationVisible(t, internalCtx, erroredStray.ID))
	})

	t.Run("sweep reaps expired never-connected only", func(t *testing.T) {
		processed := runLifecycleSweep(t, internalCtx, nil)
		require.GreaterOrEqual(t, processed, 2)

		waitForEvents()

		require.False(t, integrationVisible(t, internalCtx, expiredPending.ID))
		require.False(t, integrationVisible(t, internalCtx, erroredStray.ID))

		credentialCount, err := suite.Client.DB.Hush.Query().
			Where(hush.IDIn(append(expiredCredentialIDs, strayCredentialIDs...)...)).
			Count(internalCtx)
		require.NoError(t, err)
		require.Zero(t, credentialCount)

		require.True(t, integrationVisible(t, internalCtx, freshPending.ID))
		require.Equal(t, enums.IntegrationStatusPending, reloadIntegration(t, internalCtx, freshPending.ID).Status)

		require.Equal(t, enums.IntegrationStatusConnected, reloadIntegration(t, internalCtx, connected.ID).Status)
		require.Equal(t, 1, activeReconcileJobs(t, connectedFragment))

		require.Equal(t, enums.IntegrationStatusErrored, reloadIntegration(t, internalCtx, errored.ID).Status)
		require.Equal(t, 0, integrationNotificationCount(t, ownerCtx, errored.OwnerID, integrationReconnectedObjectType))
	})
}
