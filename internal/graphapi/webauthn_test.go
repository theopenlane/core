package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryPasskeys(t *testing.T) {
	w := (&th.WebauthnBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			expectedCount: 1,
		},
		{
			name:          "happy path, using personal access token",
			client:        suite.Client.APIWithPAT,
			ctx:           context.Background(),
			expectedCount: 1,
		},
		{
			name:   "valid user, but no passkeys",
			client: suite.Client.API,
			ctx:    th.SharedTestUser2.UserCtx,
		},
		{
			name:     "no auth",
			client:   suite.Client.API,
			ctx:      context.Background(),
			errorMsg: "could not identify authenticated user",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllWebauthns(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Webauthns.Edges, tc.expectedCount))
		})
	}

	(&th.Cleanup[*generated.WebauthnDeleteOne]{Client: suite.Client.DB.Webauthn, ID: w.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeletePasskeys(t *testing.T) {
	passkey := (&th.WebauthnBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	secondPasskey := (&th.WebauthnBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *testclient.TestClient
		ctx           context.Context
		errorMsg      string
		passkeyID     string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			passkeyID:     passkey.ID,
			expectedCount: 1, // we are deleting 1
		},
		{
			// the first test case should have deleted the passkey
			name:      "happy path, but deleted passkey",
			client:    suite.Client.APIWithPAT,
			ctx:       context.Background(),
			passkeyID: passkey.ID,
			errorMsg:  th.NotFoundErrorMsg,
		},
		{
			name:          "happy path user with pat",
			client:        suite.Client.APIWithPAT,
			ctx:           context.Background(),
			passkeyID:     secondPasskey.ID,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteWebauthn(tc.ctx, tc.passkeyID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			passkeys, err := tc.client.GetAllWebauthns(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, is.Len(passkeys.Webauthns.Edges, tc.expectedCount))
		})
	}
}
