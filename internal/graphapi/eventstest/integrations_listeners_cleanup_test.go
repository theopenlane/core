//go:build test

package eventstest_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
)

func TestIntegrationCleanupListenerHardDelete(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ownerCtx := org.UserCtx

	installation, fragment := seedHarnessLoop(t, ownerCtx)

	hardDeleteCtx := entx.SkipSoftDelete(ownerCtx)

	assert.NilError(t, suite.Client.DB.Integration.DeleteOneID(installation.ID).Exec(hardDeleteCtx))

	waitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))

	exists, err := suite.Client.DB.Integration.Query().Where(integration.ID(installation.ID)).Exist(hardDeleteCtx)
	assert.NilError(t, err)
	assert.Check(t, !exists)
}

func TestIntegrationCleanupListenerNonStatusUpdateKeepsLoops(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ownerCtx := org.UserCtx

	installation, fragment := seedHarnessLoop(t, ownerCtx)

	assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).SetName(th.RandomName(t)).Exec(ownerCtx))

	waitForEvents()

	assert.Equal(t, 1, activeReconcileJobs(t, fragment))

	assert.NilError(t, suite.Client.DB.Integration.DeleteOneID(installation.ID).Exec(ownerCtx))

	waitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))
}
