package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryPasskeys() {
	t := suite.T()

	(&WebauthnBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			expectedCount: 1,
		},
		{
			name:          "happy path, using personal access token",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			expectedCount: 1,
		},
		{
			name:   "valid user, but no passkeys",
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
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
			resp, err := tc.client.GetWebauthns(tc.ctx, nil, nil, &openlaneclient.WebauthnWhereInput{})

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.Len(t, resp.Webauthns.Edges, tc.expectedCount)
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeletePasskeys() {
	t := suite.T()

	passkey := (&WebauthnBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	secondPasskey := (&WebauthnBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
		passkeyID     string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			passkeyID:     passkey.ID,
			expectedCount: 1, // we are deleting 1
		},
		{
			// the first test case should have deleted the passkey
			name:      "happy path, but deleted passkey",
			client:    suite.client.apiWithPAT,
			ctx:       context.Background(),
			passkeyID: passkey.ID,
			errorMsg:  notFoundErrorMsg,
		},
		{
			name:          "happy path user with pat",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			passkeyID:     secondPasskey.ID,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteWebauthn(tc.ctx, tc.passkeyID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			passkeys, err := tc.client.GetWebauthns(tc.ctx, nil, nil, &openlaneclient.WebauthnWhereInput{})
			require.NoError(t, err)
			require.Len(t, passkeys.Webauthns.Edges, tc.expectedCount)
		})
	}
}
