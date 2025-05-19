package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryJobRunners(t *testing.T) {
	systemJob := (&JobRunnerBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	firstJob := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondJob := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	thirdJob := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
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
			ctx:           testUser2.UserCtx,
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
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.JobRunners.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: systemJob.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: thirdJob.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, IDs: []string{firstJob.ID, secondJob.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteJobRunner(t *testing.T) {
	systemJob := (&JobRunnerBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	firstJob := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondJob := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
		runnerID      string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			runnerID:      firstJob.ID,
			expectedCount: 1,
		},
		{
			// the first test case should have deleted the runner
			name:     "happy path, but deleted job runner",
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			runnerID: firstJob.ID,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:          "happy path user with pat",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			runnerID:      secondJob.ID,
			expectedCount: 0,
		},
		{
			name:     "happy path but cannot delete system runner",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			runnerID: systemJob.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {

			resp, err := tc.client.DeleteJobRunner(tc.ctx, tc.runnerID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			runners, err := tc.client.GetAllJobRunners(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, is.Len(runners.JobRunners.Edges, tc.expectedCount))
		})
	}
}
