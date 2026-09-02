//go:build test

package eventstest_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

func TestIntegrationCleanupListenerHardDelete(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(th.SetContext(org.UserCtx, suite.Client.DB), privacy.Allow)

	installation, fragment := seedHarnessLoop(t, allowCtx)

	hardDeleteCtx := entx.SkipSoftDelete(allowCtx)

	assert.NilError(t, suite.Client.DB.Integration.DeleteOneID(installation.ID).Exec(hardDeleteCtx))

	waitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))

	exists, err := suite.Client.DB.Integration.Query().Where(integration.ID(installation.ID)).Exist(hardDeleteCtx)
	assert.NilError(t, err)
	assert.Check(t, !exists)
}

func TestIntegrationCleanupListenerNonStatusUpdateKeepsLoops(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(th.SetContext(org.UserCtx, suite.Client.DB), privacy.Allow)

	installation, fragment := seedHarnessLoop(t, allowCtx)

	assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).SetName(th.RandomName(t)).Exec(allowCtx))

	waitForEvents()

	assert.Equal(t, 1, activeReconcileJobs(t, fragment))

	assert.NilError(t, suite.Client.DB.Integration.DeleteOneID(installation.ID).Exec(allowCtx))

	waitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))
}
