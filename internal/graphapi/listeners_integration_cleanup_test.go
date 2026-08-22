//go:build test

package graphapi_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	slackdef "github.com/theopenlane/core/internal/integrations/definitions/slack"
)

// seedConnectedSlackInstallation creates a connected slack installation for the ctx org and
// seeds exactly one reconcile loop, returning the installation and its loop metadata fragment
func seedConnectedSlackInstallation(t *testing.T, ctx context.Context) (*ent.Integration, string) {
	t.Helper()

	clientConfig, err := json.Marshal(slackdef.UserInput{})
	assert.NilError(t, err)

	installation, err := suite.client.db.Integration.Create().
		SetName(randomName(t)).
		SetKind("slack").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		SetConfig(openapi.IntegrationConfig{ClientConfig: clientConfig}).
		Save(ctx)
	assert.NilError(t, err)

	fragment := reconcileLoopFragment(t, installation.ID, slackReconcileOperation(t))

	assert.NilError(t, suite.integrationsRT.ResetReconcileLoops(ctx, installation))

	suite.WaitForEvents()

	assert.Equal(t, 1, activeReconcileJobs(t, fragment))

	return installation, fragment
}

func TestIntegrationCleanupListenerHardDelete(t *testing.T) {
	org := suite.seedFreshMinimalOrgUsers(t, false)
	allowCtx := privacy.DecisionContext(setContext(org.owner.UserCtx, suite.client.db), privacy.Allow)

	installation, fragment := seedConnectedSlackInstallation(t, allowCtx)

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

	installation, fragment := seedConnectedSlackInstallation(t, allowCtx)

	assert.NilError(t, suite.client.db.Integration.UpdateOneID(installation.ID).SetName(randomName(t)).Exec(allowCtx))

	suite.WaitForEvents()

	assert.Equal(t, 1, activeReconcileJobs(t, fragment))

	assert.NilError(t, suite.client.db.Integration.DeleteOneID(installation.ID).Exec(allowCtx))

	suite.WaitForEvents()

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))
}
