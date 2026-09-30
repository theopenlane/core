//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// TestIntegrationDegradedLifecycle verifies per-operation degradation, recovery, and escalation
func TestIntegrationDegradedLifecycle(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)

	internalCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)
	ownerCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	installation, fragment := seedHarnessLoop(t, internalCtx)

	recurringOp := testint.RecurringOp.Name()
	validatedOp := testint.ValidatedOp.Name()
	repoSyncOp := testint.RepoSyncOp.Name()

	t.Run("degrading one operation cancels only its loop", func(t *testing.T) {
		require.NoError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, installation, recurringOp, "missing directory permission"))

		reloaded := reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusDegraded, reloaded.Status)
		require.Len(t, reloaded.Health.UnhealthyOperations, 1)
		require.Contains(t, reloaded.Health.UnhealthyOperations, recurringOp)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationOperationDegradedObjectType))

		waitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))
	})

	t.Run("degrading is idempotent", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, recurringOp, "missing directory permission again"))

		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationOperationDegradedObjectType))
	})

	t.Run("clearing the only degraded operation reconnects and reseeds", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.ClearOperationUnhealthy(internalCtx, reloaded, recurringOp))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusConnected, reloaded.Status)
		require.Empty(t, reloaded.Health.UnhealthyOperations)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconnectedObjectType))

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("degrading a non-reconcile operation keeps other loops running", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, validatedOp, "required config missing upstream"))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusDegraded, reloaded.Status)
		require.Equal(t, 2, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationOperationDegradedObjectType))

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("clearing one of several failing operations stays degraded and reseeds its loop", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, recurringOp, "missing directory permission"))

		waitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.NoError(t, suite.IntegrationsRT.ClearOperationUnhealthy(internalCtx, reloaded, recurringOp))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusDegraded, reloaded.Status)
		require.Len(t, reloaded.Health.UnhealthyOperations, 1)
		require.Contains(t, reloaded.Health.UnhealthyOperations, validatedOp)

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("the last healthy workload operation failing escalates to errored", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, repoSyncOp, "repository access revoked"))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusDegraded, reloaded.Status)

		require.NoError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, recurringOp, "missing directory permission"))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.NotEmpty(t, reloaded.Health.UnhealthyReason)
		require.Len(t, reloaded.Health.UnhealthyOperations, 3)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		waitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))
	})

	t.Run("full recovery wipes every recorded failure", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.ClearIntegrationUnhealthy(internalCtx, reloaded))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusConnected, reloaded.Status)
		require.Empty(t, reloaded.Health.UnhealthyReason)
		require.Empty(t, reloaded.Health.UnhealthyOperations)
		require.Equal(t, 2, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconnectedObjectType))

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})
}
