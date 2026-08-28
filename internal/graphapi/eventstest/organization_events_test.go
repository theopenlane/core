package eventstest_test

import (
	"context"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/v2/internal/ent/historygenerated/filehistory"
	"github.com/theopenlane/core/v2/internal/ent/historygenerated/taskhistory"
)

func TestMutationOrganizationCascadeDelete(t *testing.T) {
	suite.EnableGalaForTestSuite(t)

	orgUser := suite.UserBuilder(context.Background(), t)

	org := (&th.OrganizationBuilder{Client: suite.Client}).MustNew(orgUser.UserCtx, t)

	reqCtx := auth.NewTestContextWithOrgID(orgUser.ID, org.ID)
	group1 := (&th.GroupBuilder{Client: suite.Client}).MustNew(reqCtx, t)
	customDomain := (&th.CustomDomainBuilder{Client: suite.Client}).MustNew(reqCtx, t)

	// add child org
	childOrg := (&th.OrganizationBuilder{Client: suite.Client, ParentOrgID: org.ID}).MustNew(reqCtx, t)

	// a task gives us an org owned record that tracks history
	task1 := (&th.TaskBuilder{Client: suite.Client}).MustNew(reqCtx, t)

	allowCtx := th.SetContext(reqCtx, suite.Client.DB)

	// the trust center is org owned, but the setting created alongside it is not, it only points at
	// the trust center. The cascade has to recurse through the trust center to reach it
	trustCenter := (&th.TrustCenterBuilder{Client: suite.Client}).MustNew(reqCtx, t)

	trustCenterSetting, err := suite.Client.DB.TrustCenterSetting.Query().
		Where(trustcentersetting.TrustCenterID(trustCenter.ID)).First(allowCtx)
	assert.NilError(t, err)

	// the storage path is what the file hook hands to the object storage provider on delete
	storageKey := "organizations/" + org.ID + "/cascade-test-object"

	file1, err := suite.Client.DB.File.Create().
		SetProvidedFileName("cascade-test.txt").
		SetProvidedFileExtension("txt").
		SetDetectedContentType("text/plain").
		SetStoragePath(storageKey).
		SetURI("file:///tmp/cascade-test.txt").
		AddOrganizationIDs(org.ID).
		Save(allowCtx)
	assert.NilError(t, err)

	// the history rows have to exist up front, otherwise asserting they are gone proves nothing
	assertHistoryExists(t, allowCtx, task1.ID, file1.ID, true)

	// delete org
	resp, err := suite.Client.API.DeleteOrganization(reqCtx, org.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.DeleteOrganization.DeletedID != "")

	// make sure the deletedID matches the ID we wanted to delete
	assert.Check(t, is.Equal(org.ID, resp.DeleteOrganization.DeletedID))

	_, err = suite.Client.API.GetOrganizationByID(reqCtx, org.ID)

	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	waitForCondition(t, func() bool {
		_, err := suite.Client.API.GetOrganizationByID(reqCtx, childOrg.ID)
		return err != nil && strings.Contains(err.Error(), th.NotFoundErrorMsg)
	}, "child org should be deleted by async edge cleanup")

	waitForCondition(t, func() bool {
		_, err := suite.Client.API.GetGroupByID(reqCtx, group1.ID)
		return err != nil && strings.Contains(err.Error(), th.NotFoundErrorMsg)
	}, "group should be deleted by async edge cleanup")

	waitForCondition(t, func() bool {
		// make sure the custom domain(s) no longer exists
		ctx := privacy.DecisionContext(reqCtx, privacy.Allow)
		_, err := suite.Client.DB.CustomDomain.Get(ctx, customDomain.ID)
		return generated.IsNotFound(err)
	}, "custom domain should be deleted by async edge cleanup")

	// skipping soft delete makes soft deleted rows visible, so the assertions below fail if the
	// cascade only marked the records deleted instead of removing them
	purgedCtx := entx.SkipSoftDelete(privacy.DecisionContext(reqCtx, privacy.Allow))

	// the organization row is removed last, once everything it owned is gone, so its absence is
	// what tells us the whole cascade finished. Waiting for the queue to go idle is not enough on
	// its own, a handler that errors parks the job in a retryable state the idle check ignores
	waitForCondition(t, func() bool {
		exists, err := suite.Client.DB.Organization.Query().Where(organization.ID(org.ID)).Exist(purgedCtx)

		return err == nil && !exists
	}, "organization should be hard deleted by async edge cleanup")

	taskExists, err := suite.Client.DB.Task.Query().Where(task.ID(task1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !taskExists, "task should be hard deleted with the organization, not soft deleted")

	fileExists, err := suite.Client.DB.File.Query().Where(file.ID(file1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !fileExists, "file should be hard deleted with the organization, not soft deleted")

	trustCenterExists, err := suite.Client.DB.TrustCenter.Query().Where(trustcenter.ID(trustCenter.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !trustCenterExists, "trust center should be hard deleted with the organization")

	// the setting carries no organization id of its own, it is only reachable by recursing through
	// the trust center, so this is what proves the nested cleanup runs
	settingExists, err := suite.Client.DB.TrustCenterSetting.Query().
		Where(trustcentersetting.ID(trustCenterSetting.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !settingExists, "trust center setting should be hard deleted with the organization")

	assertHistoryExists(t, purgedCtx, task1.ID, file1.ID, false)

	// the file hook only reaches object storage on a hard delete, a soft delete leaves the object
	assert.Check(t, suite.Client.DeletedStorageKeys.Has(storageKey),
		"the object backing the deleted file should have been removed from object storage")

	// the cascade runs as an internal caller, which the delete permissions hook skips by default,
	// so without the explicit opt in every cascaded record leaves its relationships behind
	groupTuples, err := suite.Client.FGA.GetTuplesForObject(context.Background(), "group:"+group1.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(groupTuples, 0), "group relationship tuples should be cleaned out of FGA")

	orgTuples, err := suite.Client.FGA.GetTuplesForObject(context.Background(), "organization:"+org.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(orgTuples, 0), "organization relationship tuples should be cleaned out of FGA")

	// ensure all tuples, like feature tuples are cleaned up
	allTuples, err := suite.Client.FGA.GetAllTuples(context.Background())
	assert.NilError(t, err)

	for _, tup := range allTuples {
		assert.Check(t, tup.Key.User != "organization:"+org.ID,
			"tuple %s#%s@%s should have been cleaned out of FGA", tup.Key.Object, tup.Key.Relation, tup.Key.User)
	}
}

// assertHistoryExists checks whether the history rows for the given task and file are present

func assertHistoryExists(t *testing.T, ctx context.Context, taskID, fileID string, want bool) {
	t.Helper()

	historyCtx := history.WithContext(ctx)

	taskHistory, err := suite.Client.DB.HistoryClient.TaskHistory.Query().
		Where(taskhistory.Ref(taskID)).Exist(historyCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(want, taskHistory), "task history presence mismatch")

	fileHistory, err := suite.Client.DB.HistoryClient.FileHistory.Query().
		Where(filehistory.Ref(fileID)).Exist(historyCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(want, fileHistory), "file history presence mismatch")
}
