//go:build test

package graphapi_test

import (
	"testing"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated/contact"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
)

func TestOrganizationCleanupListenerCascadeWithIntegrations(t *testing.T) {
	org := suite.seedFreshMinimalOrgUsers(t, false)
	orgID := org.owner.OrganizationID
	ownerCtx := org.owner.UserCtx
	allowCtx := privacy.DecisionContext(setContext(ownerCtx, suite.client.db), privacy.Allow)

	suite.WaitForEvents()

	// the suite mocks no stripe subscription cancel call, so keep the entitlements_deleted
	// listener on its skip path to avoid parking a retrying job on the shared runtime
	assert.NilError(t, suite.client.db.Organization.UpdateOneID(orgID).ClearStripeCustomerID().Exec(allowCtx))

	task1 := (&TaskBuilder{client: suite.client}).MustNew(ownerCtx, t)
	contact1 := (&ContactBuilder{client: suite.client}).MustNew(ownerCtx, t)

	installation, fragment := seedHarnessLoop(t, allowCtx)
	assert.Equal(t, orgID, installation.OwnerID)

	resp, err := suite.client.api.DeleteOrganization(ownerCtx, orgID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(orgID, resp.DeleteOrganization.DeletedID))

	suite.WaitForEvents()

	purgedCtx := entx.SkipSoftDelete(allowCtx)

	waitForCondition(t, func() bool {
		exists, err := suite.client.db.Organization.Query().Where(organization.ID(orgID)).Exist(purgedCtx)
		return err == nil && !exists
	}, "organization row should be hard deleted by the cascade")

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))

	taskExists, err := suite.client.db.Task.Query().Where(task.ID(task1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !taskExists)

	contactExists, err := suite.client.db.Contact.Query().Where(contact.ID(contact1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !contactExists)

	groupExists, err := suite.client.db.Group.Query().Where(group.ID(org.owner.GroupID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !groupExists)

	installationExists, err := suite.client.db.Integration.Query().Where(integration.ID(installation.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !installationExists)
}
