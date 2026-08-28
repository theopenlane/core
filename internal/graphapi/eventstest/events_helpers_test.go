//go:build test

package eventstest_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// notification object types mirrored from internal/integrations/runtime/health.go
const (
	integrationReconfigurationRequiredObjectType = "INTEGRATION_RECONFIGURATION_REQUIRED"
	integrationReconnectedObjectType             = "INTEGRATION_RECONNECTED"
)

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

	installation, err := suite.Client.DB.Integration.Create().
		SetName(th.RandomName(t)).
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

	require.NoError(t, suite.IntegrationsRT.Reconcile(ctx, installation, testint.ModeInput(mode), credentialRef, &credential, nil))

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

	count, err := suite.GalaRuntime.CountActiveJobsWithMetadata(context.Background(), fragment)
	require.NoError(t, err)

	return count
}

// integrationNotificationCount counts the org's notifications with the given object type
func integrationNotificationCount(t *testing.T, ctx context.Context, orgID, objectType string) int {
	t.Helper()

	count, err := suite.Client.DB.Notification.Query().
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

	installation, err := suite.Client.DB.Integration.Get(ctx, id)
	require.NoError(t, err)

	return installation
}

// TestIntegrationLifecycle drives one slack installation through seeding, unhealthy,
// recovery, listener-driven cancel/reseed, duplicate collapse, and soft delete; subtests
// share the installation and run in order

// waitForCondition polls condition until it holds or the deadline passes
func waitForCondition(t *testing.T, condition func() bool, msg string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for condition: %s", msg)
}
