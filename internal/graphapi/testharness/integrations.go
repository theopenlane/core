//go:build test

package testharness

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
	"github.com/theopenlane/iam/auth"
)

// notification object types mirrored from internal/integrations/runtime/health.go
const (
	IntegrationReconfigurationRequiredObjectType = "INTEGRATION_RECONFIGURATION_REQUIRED"
	IntegrationReconnectedObjectType             = "INTEGRATION_RECONNECTED"
)

func HarnessReconcileOperation(t *testing.T, mode string) string {
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

// NewHarnessInstallation installs the test integration in the given mode through the prod
// connect flow; the unresolvable mode stores a non-token credential so the client cannot build
func NewHarnessInstallation(t *testing.T, ctx context.Context, mode string) (*ent.Integration, string) {
	t.Helper()

	installation, err := Suite.Client.DB.Integration.Create().
		SetName(RandomName(t)).
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

	require.NoError(t, Suite.IntegrationsRT.Reconcile(ctx, installation, testint.ModeInput(mode), credentialRef, &credential, nil))

	fragment := ReconcileLoopFragment(t, installation.ID, HarnessReconcileOperation(t, mode))

	return ReloadIntegration(t, ctx, installation.ID), fragment
}

// SeedHarnessLoop installs the test integration in recurring mode and asserts the connect flow
// seeded exactly one loop
func SeedHarnessLoop(t *testing.T, ctx context.Context) (*ent.Integration, string) {
	t.Helper()

	installation, fragment := NewHarnessInstallation(t, ctx, testint.ModeRecurring)

	Suite.WaitForEvents()

	require.Equal(t, 1, ActiveReconcileJobs(t, fragment))

	return installation, fragment
}

// ReconcileLoopFragment builds the metadata containment fragment identifying the recurring
// loop jobs for one installation and operation, matching the keys ResetReconcileLoops uses
func ReconcileLoopFragment(t *testing.T, integrationID, operation string) string {
	t.Helper()

	fragment, err := integrationtypes.PropertiesFragment(map[string]string{
		"entityId":  integrationID,
		"operation": operation,
		"runType":   enums.IntegrationRunTypeReconcile.String(),
	})
	require.NoError(t, err)

	return fragment
}

// ActiveReconcileJobs counts the active River jobs matching the loop fragment
func ActiveReconcileJobs(t *testing.T, fragment string) int {
	t.Helper()

	count, err := Suite.GalaRuntime.CountActiveJobsWithMetadata(context.Background(), fragment)
	require.NoError(t, err)

	return count
}

// IntegrationNotificationCount counts the org's notifications with the given object type
func IntegrationNotificationCount(t *testing.T, ctx context.Context, orgID, objectType string) int {
	t.Helper()

	count, err := Suite.Client.DB.Notification.Query().
		Where(
			notification.OwnerID(orgID),
			notification.ObjectType(objectType),
		).
		Count(ctx)
	require.NoError(t, err)

	return count
}

// ReloadIntegration fetches the current installation row with privacy allowed
func ReloadIntegration(t *testing.T, ctx context.Context, id string) *ent.Integration {
	t.Helper()

	installation, err := Suite.Client.DB.Integration.Get(ctx, id)
	require.NoError(t, err)

	return installation
}

// TestIntegrationLifecycle drives one slack installation through seeding, unhealthy,
// recovery, listener-driven cancel/reseed, duplicate collapse, and soft delete; subtests
// share the installation and run in order

// CleanupOrganizationDataWithContext removes the caller's organization
func CleanupOrganizationDataWithContext(ctx context.Context, t *testing.T) {
	t.Helper()

	caller, _ := auth.CallerFromContext(ctx)
	if caller == nil && caller.OrganizationID == "" {
		FailNow(t)
	}

	_, err := Suite.Client.API.DeleteOrganization(ctx, caller.OrganizationID)
	RequireNoError(t, err)
}
