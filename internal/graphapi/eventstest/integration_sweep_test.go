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

// staleInstallationAge backdates a pending installation past the reap-abandoned-pending window
const staleInstallationAge = 169 * time.Hour

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

// TestIntegrationLifecycleSweep drives the lifecycle sweep operation through the real runtime
// against persisted rows: stale pending installs are reaped, deleted installs are finalized with
// their credentials, errored installs are probed back to connected, and untouched states survive
func TestIntegrationLifecycleSweep(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)

	allowCtx := privacy.DecisionContext(th.SetContext(org.UserCtx, suite.Client.DB), privacy.Allow)
	ownerCtx := th.SetContext(org.UserCtx, suite.Client.DB)

	// the audit mixin only stamps updated_at on update mutations, so a backdated
	// value survives creation but not any later update
	stalePending, err := suite.Client.DB.Integration.Create().
		SetName(th.RandomName(t)).
		SetKind("testintegration").
		SetDefinitionID(testint.DefinitionID.ID()).
		SetUpdatedAt(time.Now().Add(-staleInstallationAge)).
		Save(allowCtx)
	require.NoError(t, err)

	freshPending, err := suite.Client.DB.Integration.Create().
		SetName(th.RandomName(t)).
		SetKind("testintegration").
		SetDefinitionID(testint.DefinitionID.ID()).
		Save(allowCtx)
	require.NoError(t, err)

	connected, connectedFragment := seedHarnessLoop(t, allowCtx)

	errored, _ := newHarnessInstallation(t, allowCtx, testint.ModeRecurring)
	require.NoError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(allowCtx, errored, "credentials revoked"))

	deleted, _ := newHarnessInstallation(t, allowCtx, testint.ModeRecurring)
	deletedCredentialIDs := installationCredentialIDs(t, allowCtx, deleted.ID)
	require.NotEmpty(t, deletedCredentialIDs)
	require.NoError(t, suite.Client.DB.Integration.UpdateOneID(deleted.ID).
		SetStatus(enums.IntegrationStatusDeleted).
		Exec(allowCtx))

	waitForEvents()

	t.Run("dry run dispatches nothing", func(t *testing.T) {
		processed := runLifecycleSweep(t, allowCtx, json.RawMessage(`{"dryRun":true}`))
		require.GreaterOrEqual(t, processed, 3)

		require.True(t, integrationVisible(t, allowCtx, stalePending.ID))
		require.True(t, integrationVisible(t, allowCtx, deleted.ID))
		require.Equal(t, enums.IntegrationStatusErrored, reloadIntegration(t, allowCtx, errored.ID).Status)
	})

	t.Run("sweep reaps probes and finalizes", func(t *testing.T) {
		processed := runLifecycleSweep(t, allowCtx, nil)
		require.GreaterOrEqual(t, processed, 3)

		waitForEvents()

		require.False(t, integrationVisible(t, allowCtx, stalePending.ID))
		require.False(t, integrationVisible(t, allowCtx, deleted.ID))

		credentialCount, err := suite.Client.DB.Hush.Query().
			Where(hush.IDIn(deletedCredentialIDs...)).
			Count(allowCtx)
		require.NoError(t, err)
		require.Zero(t, credentialCount)

		require.True(t, integrationVisible(t, allowCtx, freshPending.ID))
		require.Equal(t, enums.IntegrationStatusPending, reloadIntegration(t, allowCtx, freshPending.ID).Status)

		require.Equal(t, enums.IntegrationStatusConnected, reloadIntegration(t, allowCtx, connected.ID).Status)
		require.Equal(t, 1, activeReconcileJobs(t, connectedFragment))

		require.Equal(t, enums.IntegrationStatusConnected, reloadIntegration(t, allowCtx, errored.ID).Status)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, errored.OwnerID, integrationReconnectedObjectType))
	})
}
