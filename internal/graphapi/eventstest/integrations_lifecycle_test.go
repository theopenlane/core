//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// harnessReconcileOperation returns the reconcile operation name for one harness mode
func TestIntegrationLifecycle(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)

	internalCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)
	ownerCtx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	installation, fragment := newHarnessInstallation(t, internalCtx, testint.ModeRecurring)
	require.Equal(t, org.OrganizationID, installation.OwnerID)

	opName := harnessReconcileOperation(t, testint.ModeRecurring)

	t.Run("seeding creates exactly one loop", func(t *testing.T) {
		require.NoError(t, suite.IntegrationsRT.ResetReconcileLoops(internalCtx, installation))

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("unhealthy errors the installation and cancels the loop", func(t *testing.T) {
		require.NoError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(internalCtx, installation, "credentials revoked"))

		reloaded := reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		waitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))
	})

	t.Run("unhealthy is idempotent", func(t *testing.T) {
		require.NoError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(internalCtx, installation, "credentials revoked again"))

		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))
	})

	t.Run("recovery reconnects and reseeds a single loop", func(t *testing.T) {
		reloaded := reloadIntegration(t, internalCtx, installation.ID)

		require.NoError(t, suite.IntegrationsRT.ClearIntegrationUnhealthy(internalCtx, reloaded))

		reloaded = reloadIntegration(t, internalCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusConnected, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconnectedObjectType))

		// the direct seed and the async status-change listener reseed must collapse
		// to exactly one loop
		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("listener alone cancels and reseeds on direct status flips", func(t *testing.T) {
		require.NoError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetStatus(enums.IntegrationStatusErrored).
			Exec(internalCtx))

		waitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))

		require.NoError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetStatus(enums.IntegrationStatusConnected).
			Exec(internalCtx))

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("duplicate loops collapse to one on reset", func(t *testing.T) {
		// mirror resetReconcileLoop's emit but skip the unique key so a second live loop lands
		oc := integrationtypes.NewOperationContext(installation.OwnerID, opName, integrationtypes.IntegrationSource{
			IntegrationID: installation.ID,
			DefinitionID:  installation.DefinitionID,
			RunType:       enums.IntegrationRunTypeReconcile,
		})

		emitCtx, headers := intobvs.EmitContext(internalCtx, oc)
		headers.SkipUniqueKey = true

		_, err := suite.GalaRuntime.EmitWithHeaders(emitCtx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, headers)
		require.NoError(t, err)

		waitForEvents()

		require.Equal(t, 2, activeReconcileJobs(t, fragment))

		require.NoError(t, suite.IntegrationsRT.ResetReconcileLoops(internalCtx, reloadIntegration(t, internalCtx, installation.ID)))

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("soft delete cancels the loop", func(t *testing.T) {
		require.NoError(t, suite.Client.DB.Integration.DeleteOneID(installation.ID).Exec(internalCtx))

		waitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))
	})
}
