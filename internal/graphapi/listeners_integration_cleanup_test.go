//go:build test

package graphapi_test

import (
	"testing"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func TestIntegrationCleanupListenerHardDelete(t *testing.T) {
	org := suite.seedFreshMinimalOrgUsers(t, false)
	allowCtx := privacy.DecisionContext(setContext(org.owner.UserCtx, suite.client.db), privacy.Allow)

	installation, fragment := seedHarnessLoop(t, allowCtx)

	hardDeleteCtx := entx.SkipSoftDelete(allowCtx)

	assert.NilError(t, suite.client.db.Integration.DeleteOneID(installation.ID).Exec(hardDeleteCtx))

	suite.WaitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))

	exists, err := suite.client.db.Integration.Query().Where(integration.ID(installation.ID)).Exist(hardDeleteCtx)
	assert.NilError(t, err)
	assert.Check(t, !exists)
}

func TestIntegrationCleanupListenerNonStatusUpdateKeepsLoops(t *testing.T) {
	org := suite.seedFreshMinimalOrgUsers(t, false)
	allowCtx := privacy.DecisionContext(setContext(org.owner.UserCtx, suite.client.db), privacy.Allow)

	installation, fragment := seedHarnessLoop(t, allowCtx)

	assert.NilError(t, suite.client.db.Integration.UpdateOneID(installation.ID).SetName(randomName(t)).Exec(allowCtx))

	suite.WaitForEvents()

	assert.Equal(t, 1, activeReconcileJobs(t, fragment))

	assert.NilError(t, suite.client.db.Integration.DeleteOneID(installation.ID).Exec(allowCtx))

	suite.WaitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))
}
