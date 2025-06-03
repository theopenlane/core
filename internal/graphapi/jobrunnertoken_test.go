package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryJobRunnerTokens(t *testing.T) {
	token1 := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	token2 := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	token3 := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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
			expectedCount: 4, // we created 2 runners, by default, each runner has it's own token, then we make 2 more tokens
		},
		{
			name:          "happy path, using personal access token",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			expectedCount: 4,
		},
		{
			name:          "valid test user 2",
			client:        suite.client.api,
			ctx:           testUser2.UserCtx,
			expectedCount: 2,
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
			resp, err := tc.client.GetAllJobRunnerTokens(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.JobRunnerTokens.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, IDs: []string{token1.ID, token2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, IDs: []string{token3.ID}}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationDeleteJobRunnerToken(t *testing.T) {
	firstToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		userID   string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
		tokenID  string
	}{
		{
			name:     "not enough permissions",
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			tokenID:  firstToken.ID,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:    "happy path user",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			tokenID: firstToken.ID,
		},
		{
			// the first test case should have deleted the token
			name:     "happy path, but deleted job runner token",
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			tokenID:  firstToken.ID,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, not in the correct org",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			tokenID:  secondToken.ID,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "happy path user with pat",
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			tokenID: secondToken.ID,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteJobRunnerToken(tc.ctx, tc.tokenID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			_, err = tc.client.GetJobRunnerTokenByID(tc.ctx, tc.tokenID)
			assert.Assert(t, err != nil)
		})
	}
}
