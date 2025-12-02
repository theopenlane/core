package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryJobRunnerTokens(t *testing.T) {
	newUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(newUser, t)

	anotherUser := suite.userBuilder(context.Background(), t)

	token1 := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)
	token2 := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)
	token3 := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(anotherUser.UserCtx, t)

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
			ctx:           newUser.UserCtx,
			expectedCount: 4, // we created 2 runners, by default, each runner has it's own token, then we make 2 more tokens
		},
		{
			name:          "happy path, using personal access token",
			client:        patClient,
			ctx:           context.Background(),
			expectedCount: 4,
		},
		{
			name:          "valid test user 2",
			client:        suite.client.api,
			ctx:           anotherUser.UserCtx,
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

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.JobRunnerTokens.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, IDs: []string{token1.ID, token2.ID}}).MustDelete(newUser.UserCtx, t)
	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, IDs: []string{token3.ID}}).MustDelete(anotherUser.UserCtx, t)
}

func TestMutationDeleteJobRunnerToken(t *testing.T) {
	firstToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		userID   string
		client   *testclient.TestClient
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

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			_, err = tc.client.GetJobRunnerTokenByID(tc.ctx, tc.tokenID)
			assert.Assert(t, err != nil)
		})
	}
}
