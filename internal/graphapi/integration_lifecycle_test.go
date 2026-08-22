//go:build test

package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/operations"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	testint "github.com/theopenlane/core/internal/testutils/integrations"
)

// notification object types mirrored from internal/integrations/runtime/health.go
const (
	integrationReconfigurationRequiredObjectType = "INTEGRATION_RECONFIGURATION_REQUIRED"
	integrationReconnectedObjectType             = "INTEGRATION_RECONNECTED"
)

// harnessReconcileOperation returns the reconcile operation name for one harness mode
func harnessReconcileOperation(t *testing.T, mode string) string {
	t.Helper()

	switch mode {
	case testint.ModeRecurring:
		return testint.RecurringOp.Name()
	case testint.ModeExhausting:
		return testint.ExhaustingOp.Name()
	case testint.ModeUnresolvable:
		return testint.UnresolvableOp.Name()
	}

	t.Fatalf("unknown harness mode %q", mode)

	return ""
}

// newHarnessInstallation installs the test integration in the given mode through the prod
// connect flow; the unresolvable mode stores a non-token credential so the client cannot build
func newHarnessInstallation(t *testing.T, ctx context.Context, mode string) (*ent.Integration, string) {
	t.Helper()

	installation, err := suite.client.db.Integration.Create().
		SetName(randomName(t)).
		SetKind("testintegration").
		SetDefinitionID(testint.DefinitionID.ID()).
		Save(ctx)
	require.NoError(t, err)

	credentialRef := testint.TokenCredential.ID()
	credential := testint.TokenCredentialSet("test-token")

	if mode == testint.ModeUnresolvable {
		credentialRef = testint.ServiceAccountCredential.ID()
		credential = testint.ServiceAccountCredentialSet("test-project", "svc@example.com")
	}

	require.NoError(t, suite.integrationsRT.Reconcile(ctx, installation, testint.ModeInput(mode), credentialRef, &credential, nil))

	fragment := reconcileLoopFragment(t, installation.ID, harnessReconcileOperation(t, mode))

	return reloadIntegration(t, ctx, installation.ID), fragment
}

// seedHarnessLoop installs the test integration in recurring mode and asserts the connect flow
// seeded exactly one loop
func seedHarnessLoop(t *testing.T, ctx context.Context) (*ent.Integration, string) {
	t.Helper()

	installation, fragment := newHarnessInstallation(t, ctx, testint.ModeRecurring)

	suite.WaitForEvents()

	require.Equal(t, 1, activeReconcileJobs(t, fragment))

	return installation, fragment
}

// reconcileLoopFragment builds the metadata containment fragment identifying the recurring
// loop jobs for one installation and operation, matching the keys ResetReconcileLoops uses
func reconcileLoopFragment(t *testing.T, integrationID, operation string) string {
	t.Helper()

	fragment, err := integrationtypes.PropertiesFragment(map[string]string{
		"entityId":  integrationID,
		"operation": operation,
		"runType":   enums.IntegrationRunTypeReconcile.String(),
	})
	require.NoError(t, err)

	return fragment
}

// activeReconcileJobs counts the active River jobs matching the loop fragment
func activeReconcileJobs(t *testing.T, fragment string) int {
	t.Helper()

	count, err := suite.galaRuntime.CountActiveJobsWithMetadata(context.Background(), fragment)
	require.NoError(t, err)

	return count
}

// integrationNotificationCount counts the org's notifications with the given object type
func integrationNotificationCount(t *testing.T, ctx context.Context, orgID, objectType string) int {
	t.Helper()

	count, err := suite.client.db.Notification.Query().
		Where(
			notification.OwnerID(orgID),
			notification.ObjectType(objectType),
		).
		Count(ctx)
	require.NoError(t, err)

	return count
}

// reloadIntegration fetches the current installation row with privacy allowed
func reloadIntegration(t *testing.T, ctx context.Context, id string) *ent.Integration {
	t.Helper()

	installation, err := suite.client.db.Integration.Get(ctx, id)
	require.NoError(t, err)

	return installation
}

// TestIntegrationLifecycle drives one slack installation through seeding, unhealthy,
// recovery, listener-driven cancel/reseed, duplicate collapse, and soft delete; subtests
// share the installation and run in order
func TestIntegrationLifecycle(t *testing.T) {
	org := suite.seedFreshMinimalOrgUsers(t, false)

	allowCtx := privacy.DecisionContext(setContext(org.owner.UserCtx, suite.client.db), privacy.Allow)
	ownerCtx := setContext(org.owner.UserCtx, suite.client.db)

	installation, fragment := newHarnessInstallation(t, allowCtx, testint.ModeRecurring)
	require.Equal(t, org.owner.OrganizationID, installation.OwnerID)

	opName := harnessReconcileOperation(t, testint.ModeRecurring)

	t.Run("seeding creates exactly one loop", func(t *testing.T) {
		require.NoError(t, suite.integrationsRT.ResetReconcileLoops(allowCtx, installation))

		suite.WaitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("unhealthy errors the installation and cancels the loop", func(t *testing.T) {
		require.NoError(t, suite.integrationsRT.MarkIntegrationUnhealthy(allowCtx, installation, "credentials revoked"))

		reloaded := reloadIntegration(t, allowCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		suite.WaitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))
	})

	t.Run("unhealthy is idempotent", func(t *testing.T) {
		require.NoError(t, suite.integrationsRT.MarkIntegrationUnhealthy(allowCtx, installation, "credentials revoked again"))

		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))
	})

	t.Run("recovery reconnects and reseeds a single loop", func(t *testing.T) {
		reloaded := reloadIntegration(t, allowCtx, installation.ID)

		require.NoError(t, suite.integrationsRT.ClearIntegrationUnhealthy(allowCtx, reloaded))

		reloaded = reloadIntegration(t, allowCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusConnected, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconnectedObjectType))

		// the direct seed and the async status-change listener reseed must collapse
		// to exactly one loop
		suite.WaitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("listener alone cancels and reseeds on direct status flips", func(t *testing.T) {
		require.NoError(t, suite.client.db.Integration.UpdateOneID(installation.ID).
			SetStatus(enums.IntegrationStatusErrored).
			Exec(allowCtx))

		suite.WaitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))

		require.NoError(t, suite.client.db.Integration.UpdateOneID(installation.ID).
			SetStatus(enums.IntegrationStatusConnected).
			Exec(allowCtx))

		suite.WaitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("duplicate loops collapse to one on reset", func(t *testing.T) {
		// mirror emitReconcileLoop but bypass the topic's insert-time dedup so a
		// second live loop for the same operation context actually lands
		oc := integrationtypes.NewOperationContext(installation.OwnerID, opName, integrationtypes.IntegrationSource{
			IntegrationID: installation.ID,
			DefinitionID:  installation.DefinitionID,
			RunType:       enums.IntegrationRunTypeReconcile,
		})

		emitCtx, headers := intobvs.EmitContext(allowCtx, oc)
		headers.SkipUniqueKey = true

		_, err := suite.galaRuntime.EmitWithHeaders(emitCtx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, headers)
		require.NoError(t, err)

		suite.WaitForEvents()

		require.Equal(t, 2, activeReconcileJobs(t, fragment))

		require.NoError(t, suite.integrationsRT.ResetReconcileLoops(allowCtx, reloadIntegration(t, allowCtx, installation.ID)))

		suite.WaitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, fragment))
	})

	t.Run("soft delete cancels the loop", func(t *testing.T) {
		require.NoError(t, suite.client.db.Integration.DeleteOneID(installation.ID).Exec(allowCtx))

		suite.WaitForEvents()

		require.Equal(t, 0, activeReconcileJobs(t, fragment))
	})
}
