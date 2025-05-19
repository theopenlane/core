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
	firstToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	thirdToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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

	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, ID: firstToken.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, ID: secondToken.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.JobRunnerTokenDeleteOne]{client: suite.client.db.JobRunnerToken, ID: thirdToken.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationDeleteJobRunnerToken(t *testing.T) {
	firstToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondToken := (&JobRunnerTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
		tokenID       string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			tokenID:       firstToken.ID,
			expectedCount: 3,
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
			name:          "happy path user with pat",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			tokenID:       secondToken.ID,
			expectedCount: 2,
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

			tokens, err := tc.client.GetAllJobRunnerTokens(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, is.Len(tokens.JobRunnerTokens.Edges, tc.expectedCount))
		})
	}
}
