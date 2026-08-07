package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/ent/historygenerated/filehistory"
	"github.com/theopenlane/core/internal/ent/historygenerated/taskhistory"
)

// TestMutationDeleteOrganizationCascade covers what happens to the records an organization owns once
// it is deleted, they must be hard deleted, their history rows purged and the objects backing any
// files removed from object storage rather than left orphaned
func TestMutationDeleteOrganizationCascade(t *testing.T) {
	orgUser := suite.seedFreshMinimalOrgUsers(t, false)

	reqCtx := orgUser.owner.UserCtx
	orgID := orgUser.owner.OrganizationID

	// a task gives us an org owned record that tracks history
	task1 := (&TaskBuilder{client: suite.client}).MustNew(reqCtx, t)

	// the storage path is what the file hook hands to the object storage provider on delete
	storageKey := "organizations/" + orgID + "/cascade-test-object"

	allowCtx := setContext(reqCtx, suite.client.db)

	file1, err := suite.client.db.File.Create().
		SetProvidedFileName("cascade-test.txt").
		SetProvidedFileExtension("txt").
		SetDetectedContentType("text/plain").
		SetStoragePath(storageKey).
		SetURI("file:///tmp/cascade-test.txt").
		AddOrganizationIDs(orgID).
		Save(allowCtx)
	assert.NilError(t, err)

	// the history rows have to exist up front, otherwise asserting they are gone proves nothing
	assertHistoryExists(t, allowCtx, task1.ID, file1.ID, true)

	resp, err := suite.client.api.DeleteOrganization(reqCtx, orgID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(orgID, resp.DeleteOrganization.DeletedID))

	// the cascade runs on a gala worker, so wait for it to drain before asserting
	suite.WaitForEvents()

	// skipping soft delete makes soft deleted rows visible, so these assertions fail if the cascade
	// only marked the records deleted instead of removing them
	purgedCtx := entx.SkipSoftDelete(allowCtx)

	orgExists, err := suite.client.db.Organization.Query().Where(organization.ID(orgID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !orgExists, "organization should be hard deleted, not left soft deleted")

	taskExists, err := suite.client.db.Task.Query().Where(task.ID(task1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !taskExists, "task should be hard deleted with the organization, not soft deleted")

	fileExists, err := suite.client.db.File.Query().Where(file.ID(file1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !fileExists, "file should be hard deleted with the organization, not soft deleted")

	assertHistoryExists(t, allowCtx, task1.ID, file1.ID, false)

	// the file hook only reaches object storage on a hard delete, a soft delete leaves the object
	assert.Check(t, suite.client.deletedStorageKeys.Has(storageKey),
		"the object backing the deleted file should have been removed from object storage")
}

// assertHistoryExists checks whether the history rows for the given task and file are present
func assertHistoryExists(t *testing.T, ctx context.Context, taskID, fileID string, want bool) {
	t.Helper()

	historyCtx := history.WithContext(ctx)

	taskHistory, err := suite.client.db.HistoryClient.TaskHistory.Query().
		Where(taskhistory.Ref(taskID)).Exist(historyCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(want, taskHistory), "task history presence mismatch")

	fileHistory, err := suite.client.db.HistoryClient.FileHistory.Query().
		Where(filehistory.Ref(fileID)).Exist(historyCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(want, fileHistory), "file history presence mismatch")
}
