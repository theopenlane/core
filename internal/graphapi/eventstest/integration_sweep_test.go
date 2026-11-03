//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
	assert.NilError(t, err)

	var result integrationtypes.ScheduledCycleResult
	assert.NilError(t, json.Unmarshal(raw, &result))

	return result.Processed
}

// integrationVisible reports whether the installation row is still readable
func integrationVisible(t *testing.T, ctx context.Context, id string) bool {
	t.Helper()

	_, err := suite.Client.DB.Integration.Get(ctx, id)
	if ent.IsNotFound(err) {
		return false
	}

	assert.NilError(t, err)

	return true
}

// installationCredentialIDs lists the hush credential row IDs attached to the installation
func installationCredentialIDs(t *testing.T, ctx context.Context, integrationID string) []string {
	t.Helper()

	installation, err := suite.Client.DB.Integration.Get(ctx, integrationID)
	assert.NilError(t, err)

	ids, err := installation.QuerySecrets().IDs(ctx)
	assert.NilError(t, err)

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
	assert.Assert(t, len(expiredCredentialIDs) > 0)
	assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(expiredPending.ID).
		SetStatus(enums.IntegrationStatusPending).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Exec(internalCtx))

	freshPending, err := suite.Client.DB.Integration.Create().
		SetName(th.RandomName(t)).
		SetKind("testintegration").
		SetDefinitionID(testint.DefinitionID.ID()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		Save(internalCtx)
	assert.NilError(t, err)

	connected, connectedFragment := seedHarnessLoop(t, internalCtx)

	errored, _ := newHarnessInstallation(t, internalCtx, testint.ModeRecurring)
	assert.NilError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(internalCtx, errored, "credentials revoked"))

	erroredStray, _ := newHarnessInstallation(t, internalCtx, testint.ModeRecurring)
	strayCredentialIDs := installationCredentialIDs(t, internalCtx, erroredStray.ID)
	assert.Assert(t, len(strayCredentialIDs) > 0)
	assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(erroredStray.ID).
		SetStatus(enums.IntegrationStatusPending).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Exec(internalCtx))
	assert.NilError(t, suite.IntegrationsRT.MarkIntegrationUnhealthy(internalCtx, reloadIntegration(t, internalCtx, erroredStray.ID), "connection check failed"))

	flippedStray := reloadIntegration(t, internalCtx, erroredStray.ID)
	assert.Assert(t, is.Equal(enums.IntegrationStatusErrored, flippedStray.Status))
	assert.Assert(t, flippedStray.ExpiresAt != nil)

	waitForEvents()

	t.Run("dry run dispatches nothing", func(t *testing.T) {
		processed := runLifecycleSweep(t, internalCtx, json.RawMessage(`{"dryRun":true}`))
		assert.Assert(t, processed >= 2)

		assert.Assert(t, integrationVisible(t, internalCtx, expiredPending.ID))
		assert.Assert(t, integrationVisible(t, internalCtx, erroredStray.ID))
	})

	t.Run("sweep reaps expired never-connected only", func(t *testing.T) {
		processed := runLifecycleSweep(t, internalCtx, nil)
		assert.Assert(t, processed >= 2)

		waitForEvents()

		assert.Assert(t, !integrationVisible(t, internalCtx, expiredPending.ID))
		assert.Assert(t, !integrationVisible(t, internalCtx, erroredStray.ID))

		credentialCount, err := suite.Client.DB.Hush.Query().
			Where(hush.IDIn(append(expiredCredentialIDs, strayCredentialIDs...)...)).
			Count(internalCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Equal(credentialCount, 0))

		assert.Assert(t, integrationVisible(t, internalCtx, freshPending.ID))
		assert.Assert(t, is.Equal(enums.IntegrationStatusPending, reloadIntegration(t, internalCtx, freshPending.ID).Status))

		assert.Assert(t, is.Equal(enums.IntegrationStatusConnected, reloadIntegration(t, internalCtx, connected.ID).Status))
		assert.Assert(t, is.Equal(1, activeReconcileJobs(t, connectedFragment)))

		assert.Assert(t, is.Equal(enums.IntegrationStatusErrored, reloadIntegration(t, internalCtx, errored.ID).Status))
		assert.Assert(t, is.Equal(0, integrationNotificationCount(t, ownerCtx, errored.OwnerID, integrationReconnectedObjectType)))
	})
}
