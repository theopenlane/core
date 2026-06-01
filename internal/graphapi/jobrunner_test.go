package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryJobRunners(t *testing.T) {
	systemJob := (&JobRunnerBuilder{client: suite.client}).MustNew(sharedSystemAdminUser.UserCtx, t)
	firstJob := (&JobRunnerBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	secondJob := (&JobRunnerBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	thirdJob := (&JobRunnerBuilder{client: suite.client}).MustNew(sharedTestUser2.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *testclient.TestClient
		ctx           context.Context
		errorMsg      string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.client.api,
			ctx:           sharedTestUser1.UserCtx,
			expectedCount: 2,
		},
		{
			name:          "happy path, using personal access token",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			expectedCount: 2,
		},
		{
			name:          "valid test user 2",
			client:        suite.client.api,
			ctx:           sharedTestUser2.UserCtx,
			expectedCount: 1,
		},
		{
			name:     "no auth",
			client:   suite.client.api,
			ctx:      context.Background(),
			errorMsg: "could not identify authenticated user",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllJobRunners(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.JobRunners.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: systemJob.ID}).MustDelete(sharedSystemAdminUser.UserCtx, t)
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: thirdJob.ID}).MustDelete(sharedTestUser2.UserCtx, t)
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, IDs: []string{firstJob.ID, secondJob.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestMutationDeleteJobRunner(t *testing.T) {
	t.Parallel()
	localTestUser := suite.seedOrgOwner(t)

	systemJob := (&JobRunnerBuilder{client: suite.client}).MustNew(sharedSystemAdminUser.UserCtx, t)
	firstJob := (&JobRunnerBuilder{client: suite.client}).MustNew(localTestUser.owner.UserCtx, t)
	secondJob := (&JobRunnerBuilder{client: suite.client}).MustNew(localTestUser.owner.UserCtx, t)

	testCases := []struct {
		name     string
		userID   string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
		runnerID string
	}{
		{
			name:     "happy path user",
			client:   suite.client.api,
			ctx:      localTestUser.owner.UserCtx,
			runnerID: firstJob.ID,
		},
		{
			// the first test case should have deleted the runner
			name:     "job runner already deleted",
			client:   localTestUser.patClient,
			ctx:      context.Background(),
			runnerID: firstJob.ID,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "happy path user with pat",
			client:   localTestUser.patClient,
			ctx:      context.Background(),
			runnerID: secondJob.ID,
		},
		{
			name:     "happy path but cannot delete system runner",
			client:   localTestUser.apiClient,
			ctx:      context.Background(),
			runnerID: systemJob.ID,
			errorMsg: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteJobRunner(tc.ctx, tc.runnerID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}

	cleanupOrganizationDataWithContext(localTestUser.owner.UserCtx, t)
}
