//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// TestIntegrationDegradedLifecycle drives one installation through per-operation degradation,
// idempotent marking, single-operation recovery, escalation to errored when the last healthy
// workload operation fails, and full recovery; subtests share the installation and run in order
func TestIntegrationDegradedLifecycle(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)

	internalCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)
	ownerCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	installation, fragment := seedHarnessLoop(t, internalCtx)

	recurringOp := testint.RecurringOp.Name()
	validatedOp := testint.ValidatedOp.Name()
	repoSyncOp := testint.RepoSyncOp.Name()

	t.Run("degrading one operation cancels only its loop", func(t *testing.T) {
		assert.NilError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, installation, recurringOp, "missing directory permission"))

		reloaded := reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusDegraded, reloaded.Status))
		assert.Assert(t, is.Len(reloaded.Health.UnhealthyOperations, 1))
		assert.Assert(t, is.Contains(reloaded.Health.UnhealthyOperations, recurringOp))
		assert.Assert(t, is.Equal(1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationOperationDegradedObjectType)))

		// the status-change listener reseed must skip the unhealthy operation
		waitForEvents()

		assert.Assert(t, is.Equal(0, activeReconcileJobs(t, fragment)))
	})

	t.Run("degrading is idempotent", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		assert.NilError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, recurringOp, "missing directory permission again"))

		assert.Assert(t, is.Equal(1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationOperationDegradedObjectType)))
	})

	t.Run("clearing the only degraded operation reconnects and reseeds", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		assert.NilError(t, suite.IntegrationsRT.ClearOperationUnhealthy(internalCtx, reloaded, recurringOp))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusConnected, reloaded.Status))
		assert.Assert(t, is.Len(reloaded.Health.UnhealthyOperations, 0))
		assert.Assert(t, is.Equal(1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconnectedObjectType)))

		waitForEvents()

		assert.Assert(t, is.Equal(1, activeReconcileJobs(t, fragment)))
	})

	t.Run("degrading a non-reconcile operation keeps other loops running", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		assert.NilError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, validatedOp, "required config missing upstream"))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusDegraded, reloaded.Status))
		assert.Assert(t, is.Equal(2, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationOperationDegradedObjectType)))

		waitForEvents()

		assert.Assert(t, is.Equal(1, activeReconcileJobs(t, fragment)))
	})

	t.Run("clearing one of several failing operations stays degraded and reseeds its loop", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		assert.NilError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, recurringOp, "missing directory permission"))

		waitForEvents()

		assert.Assert(t, is.Equal(0, activeReconcileJobs(t, fragment)))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.NilError(t, suite.IntegrationsRT.ClearOperationUnhealthy(internalCtx, reloaded, recurringOp))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusDegraded, reloaded.Status))
		assert.Assert(t, is.Len(reloaded.Health.UnhealthyOperations, 1))
		assert.Assert(t, is.Contains(reloaded.Health.UnhealthyOperations, validatedOp))

		waitForEvents()

		assert.Assert(t, is.Equal(1, activeReconcileJobs(t, fragment)))
	})

	t.Run("the last healthy workload operation failing escalates to errored", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		assert.NilError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, repoSyncOp, "repository access revoked"))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusDegraded, reloaded.Status))

		assert.NilError(t, suite.IntegrationsRT.MarkOperationUnhealthy(internalCtx, reloaded, recurringOp, "missing directory permission"))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusErrored, reloaded.Status))
		assert.Assert(t, reloaded.Health.UnhealthyReason != "")
		// escalation keeps the per-operation reasons for recovery surfaces
		assert.Assert(t, is.Len(reloaded.Health.UnhealthyOperations, 3))
		assert.Assert(t, is.Equal(1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType)))

		waitForEvents()

		assert.Assert(t, is.Equal(0, activeReconcileJobs(t, fragment)))
	})

	t.Run("full recovery wipes every recorded failure", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		assert.NilError(t, suite.IntegrationsRT.ClearIntegrationUnhealthy(internalCtx, reloaded))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		assert.Assert(t, is.Equal(enums.IntegrationStatusConnected, reloaded.Status))
		assert.Assert(t, is.Equal(reloaded.Health.UnhealthyReason, ""))
		assert.Assert(t, is.Len(reloaded.Health.UnhealthyOperations, 0))
		assert.Assert(t, is.Equal(2, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconnectedObjectType)))

		waitForEvents()

		assert.Assert(t, is.Equal(1, activeReconcileJobs(t, fragment)))
	})
}
