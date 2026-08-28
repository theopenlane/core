//go:build test

package eventstest_test

import (
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
)

func TestOrganizationCleanupListenerCascadeWithIntegrations(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	orgID := org.Owner.OrganizationID
	ownerCtx := org.Owner.UserCtx
	allowCtx := privacy.DecisionContext(th.SetContext(ownerCtx, suite.Client.DB), privacy.Allow)

	waitForEvents()

	// the suite mocks no stripe subscription cancel call, so keep the entitlements_deleted
	// listener on its skip path to avoid parking a retrying job on the shared runtime
	assert.NilError(t, suite.Client.DB.Organization.UpdateOneID(orgID).ClearStripeCustomerID().Exec(allowCtx))

	task1 := (&th.TaskBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	contact1 := (&th.ContactBuilder{Client: suite.Client}).MustNew(ownerCtx, t)

	installation, fragment := seedHarnessLoop(t, allowCtx)
	assert.Equal(t, orgID, installation.OwnerID)

	resp, err := suite.Client.API.DeleteOrganization(ownerCtx, orgID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(orgID, resp.DeleteOrganization.DeletedID))

	waitForEvents()

	purgedCtx := entx.SkipSoftDelete(allowCtx)

	waitForCondition(t, func() bool {
		exists, err := suite.Client.DB.Organization.Query().Where(organization.ID(orgID)).Exist(purgedCtx)
		return err == nil && !exists
	}, "organization row should be hard deleted by the cascade")

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))

	taskExists, err := suite.Client.DB.Task.Query().Where(task.ID(task1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !taskExists)

	contactExists, err := suite.Client.DB.Contact.Query().Where(contact.ID(contact1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !contactExists)

	groupExists, err := suite.Client.DB.Group.Query().Where(group.ID(org.Owner.GroupID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !groupExists)

	installationExists, err := suite.Client.DB.Integration.Query().Where(integration.ID(installation.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !installationExists)
}
