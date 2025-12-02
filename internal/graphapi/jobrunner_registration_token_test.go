package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryJobRunnerRegistrationTokens(t *testing.T) {
	// auto cleaned up when the second job is created
	_ = (&JobRunnerRegistrationTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondJob := (&JobRunnerRegistrationTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// auto cleaned up by hook when the last job is created
	// the last job itself is deleted since we are attaching it to a runner
	_ = (&JobRunnerRegistrationTokenBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	lastJob := (&JobRunnerRegistrationTokenBuilder{client: suite.client, WithRunner: true}).MustNew(testUser2.UserCtx, t)

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
			ctx:           testUser1.UserCtx,
			expectedCount: 1,
		},
		{
			name:          "happy path, using personal access token",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			expectedCount: 1,
		},
		{
			name:          "valid test user 2",
			client:        suite.client.api,
			ctx:           testUser2.UserCtx,
			expectedCount: 0, // 0 since the fourthJob sets a runner so it should be deleted too in the hook
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
			resp, err := tc.client.GetAllJobRunnerRegistrationTokens(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.JobRunnerRegistrationTokens.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerRegistrationTokenDeleteOne]{client: suite.client.db.JobRunnerRegistrationToken, ID: secondJob.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: lastJob.JobRunnerID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationDeleteJobRunnerRegistrationToken(t *testing.T) {

	firstJob := (&JobRunnerRegistrationTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondJob := (&JobRunnerRegistrationTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	thirdJob := (&JobRunnerRegistrationTokenBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *testclient.TestClient
		ctx           context.Context
		errorMsg      string
		runnerID      string
		expectedCount int
	}{
		{
			name:     "happy path user",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			runnerID: firstJob.ID,
			// expected, we create the first job then second one.
			// Our hook clears all existing registration tokens on new one
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "happy path user with pat",
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			runnerID: secondJob.ID,
		},
		{
			name:          "happy path but cannot delete token no access to",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			runnerID:      thirdJob.ID,
			errorMsg:      notFoundErrorMsg,
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteJobRunnerRegistrationToken(tc.ctx, tc.runnerID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			runners, err := tc.client.GetAllJobRunnerRegistrationTokens(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, is.Len(runners.JobRunnerRegistrationTokens.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerRegistrationTokenDeleteOne]{client: suite.client.db.JobRunnerRegistrationToken, ID: thirdJob.ID}).MustDelete(testUser2.UserCtx, t)
}
