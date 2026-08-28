//go:build test

package eventstest_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// waitForInstallationErrored polls until the installation is marked unhealthy; the exhausting
// loop reschedules through River's scheduler across several cycles, so it needs a longer window
// than the shared waitForCondition helper allows
func waitForInstallationErrored(t *testing.T, ctx context.Context, id string) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		inst, err := suite.Client.DB.Integration.Get(ctx, id)
		if err == nil && inst.Status == enums.IntegrationStatusErrored {
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatal("timed out waiting for exhausting loop to mark the installation unhealthy")
}

// TestReconcileLoopExhaustsToUnhealthy drives a loop whose every cycle fails and asserts the
// runtime stops rescheduling after the error budget and marks the installation unhealthy
func TestReconcileLoopExhaustsToUnhealthy(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	allowCtx := privacy.DecisionContext(th.SetContext(org.Owner.UserCtx, suite.Client.DB), privacy.Allow)
	ownerCtx := th.SetContext(org.Owner.UserCtx, suite.Client.DB)

	installation, fragment := newHarnessInstallation(t, allowCtx, testint.ModeExhausting)

	require.NoError(t, suite.IntegrationsRT.ResetReconcileLoops(allowCtx, installation))

	waitForInstallationErrored(t, allowCtx, installation.ID)

	waitForEvents()

	require.Equal(t, 0, activeReconcileJobs(t, fragment))
	require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))
}

// TestReconcileLoopUnresolvableClientMarksUnhealthy asserts a loop whose client cannot be built
// is never seeded and the installation is marked unhealthy at seed time
func TestReconcileLoopUnresolvableClientMarksUnhealthy(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	allowCtx := privacy.DecisionContext(th.SetContext(org.Owner.UserCtx, suite.Client.DB), privacy.Allow)
	ownerCtx := th.SetContext(org.Owner.UserCtx, suite.Client.DB)

	installation, fragment := newHarnessInstallation(t, allowCtx, testint.ModeUnresolvable)

	require.NoError(t, suite.IntegrationsRT.ResetReconcileLoops(allowCtx, installation))

	waitForEvents()

	require.Equal(t, 0, activeReconcileJobs(t, fragment))

	reloaded := reloadIntegration(t, allowCtx, installation.ID)
	require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
	require.Equal(t, 1, integrationNotificationCount(t, ownerCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))
}
