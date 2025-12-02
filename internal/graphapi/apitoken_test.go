package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/hooks"
	openlaneclient "github.com/theopenlane/go-client"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/testutils"
)

func TestQueryApiToken(t *testing.T) {
	apiToken := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: apiToken.ID,
			ctx:     testUser1.UserCtx,
		},
		{
			name:     "not found, no access",
			queryID:  apiToken.ID,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     notFoundErrorMsg,
			queryID:  "notfound",
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetAPITokenByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(redacted, resp.APIToken.Token))
			assert.Check(t, is.Equal(testUser1.OrganizationID, resp.APIToken.Owner.ID))
		})
	}

	(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, ID: apiToken.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryAPITokens(t *testing.T) {
	token1 := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	token2 := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		errorMsg string
	}{
		{
			name: "happy path, all api tokens",
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetAllAPITokens(testUser1.UserCtx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// this is three because we create two tokens in the test
			// and there is one created in the suite setup
			assert.Check(t, is.Len(resp.APITokens.Edges, 3))
		})
	}

	(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, ID: token1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, ID: token2.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateAPIToken(t *testing.T) {
	tokenDescription := gofakeit.Sentence()
	expiration30Days := time.Now().Add(time.Hour * 24 * 30)

	testCases := []struct {
		name     string
		input    testclient.CreateAPITokenInput
		errorMsg string
	}{
		{
			name: "happy path",
			input: testclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				Scopes:      []string{"read", "write"},
			},
		},
		{
			name: "bad path, set expire to the past",
			input: testclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   lo.ToPtr(time.Now().Add(-time.Hour)),
			},
			errorMsg: hooks.ErrPastTimeNotAllowed.Error(),
		},
		{
			name: "happy path, set expire",
			input: testclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   &expiration30Days,
			},
		},
		{
			name: "happy path, set org",
			input: testclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   &expiration30Days,
			},
		},
		{
			name: "happy path, name only",
			input: testclient.CreateAPITokenInput{
				Name: "forthethingz",
			},
		},
		{
			name: "empty name",
			input: testclient.CreateAPITokenInput{
				Description: &tokenDescription,
			},
			errorMsg: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.CreateAPIToken(testUser1.UserCtx, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.input.Name, resp.CreateAPIToken.APIToken.Name))
			assert.Check(t, is.DeepEqual(tc.input.Description, resp.CreateAPIToken.APIToken.Description))
			assert.Check(t, is.DeepEqual(tc.input.Scopes, resp.CreateAPIToken.APIToken.Scopes))

			// check expiration if set
			if tc.input.ExpiresAt == nil {
				assert.Equal(t, resp.CreateAPIToken.APIToken.ExpiresAt, (*time.Time)(nil))
			} else {
				assert.Check(t, tc.input.ExpiresAt.Equal(*resp.CreateAPIToken.APIToken.ExpiresAt))
			}

			// ensure the owner is the org set in the request
			assert.Check(t, is.Equal(testUser1.OrganizationID, *resp.CreateAPIToken.APIToken.OwnerID))

			// token should not be redacted on create
			assert.Check(t, redacted != resp.CreateAPIToken.APIToken.Token)

			// ensure the token is prefixed
			assert.Check(t, is.Contains(resp.CreateAPIToken.APIToken.Token, "tola_"))

			(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, ID: resp.CreateAPIToken.APIToken.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateAPIToken(t *testing.T) {
	token := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	tokenDescription := gofakeit.Sentence()
	tokenName := gofakeit.Word()

	testCases := []struct {
		name     string
		tokenID  string
		input    testclient.UpdateAPITokenInput
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path, update name ",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Name: &tokenName,
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "happy path, update expiration",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Name:      &tokenName,
				ExpiresAt: lo.ToPtr(time.Now().Add(time.Hour)),
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "update name, no access",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Name: &tokenName,
			},
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "happy path, update description",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Description: &tokenDescription,
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "happy path, add scope",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Scopes: []string{"write"},
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "invalid token id",
			tokenID: "notvalidtoken",
			input: testclient.UpdateAPITokenInput{
				Description: &tokenDescription,
			},
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.UpdateAPIToken(tc.ctx, tc.tokenID, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.input.Name != nil {
				assert.Check(t, is.Equal(resp.UpdateAPIToken.APIToken.Name, *tc.input.Name))
			}

			if tc.input.Description != nil {
				assert.Check(t, is.DeepEqual(resp.UpdateAPIToken.APIToken.Description, tc.input.Description))
			}

			// Ensure its added
			if tc.input.Scopes != nil {
				assert.Check(t, is.Len(resp.UpdateAPIToken.APIToken.Scopes, 1))
			}

			assert.Check(t, is.Equal(testUser1.OrganizationID, *resp.UpdateAPIToken.APIToken.OwnerID))

			// token should be redacted on update
			assert.Check(t, is.Equal(redacted, resp.UpdateAPIToken.APIToken.Token))
		})
	}

	(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, ID: token.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteAPIToken(t *testing.T) {
	// create user to make tokens
	user := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	user2 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	orgID := user.Edges.Setting.Edges.DefaultOrg.ID
	orgID2 := user2.Edges.Setting.Edges.DefaultOrg.ID

	reqCtx := auth.NewTestContextWithOrgID(user.ID, orgID)

	token := (&APITokenBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx2 := auth.NewTestContextWithOrgID(user2.ID, orgID2)

	token2 := (&APITokenBuilder{client: suite.client}).MustNew(reqCtx2, t)

	testCases := []struct {
		name     string
		tokenID  string
		errorMsg string
	}{
		{
			name:    "happy path, delete token",
			tokenID: token.ID,
		},
		{
			name:     "delete someone else's token, no go",
			tokenID:  token2.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.DeleteAPIToken(reqCtx, tc.tokenID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.tokenID, resp.DeleteAPIToken.DeletedID))
		})
	}
}

func TestLastUsedAPIToken(t *testing.T) {
	// create new API token
	token := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// check that the last used is empty
	res, err := suite.client.api.GetAPITokenByID(testUser1.UserCtx, token.ID)
	assert.NilError(t, err)
	assert.Check(t, res.APIToken.LastUsedAt == nil)

	// setup graph client using the API token
	authHeader := openlaneclient.Authorization{
		BearerToken: token.Token,
	}

	graphClient, err := testutils.TestClientWithAuth(suite.client.db, suite.client.objectStore,
		openlaneclient.WithCredentials(authHeader),
	)
	assert.NilError(t, err)

	// get the token to make sure the last used is updated using the token
	out, err := graphClient.GetAPITokenByID(context.Background(), token.ID)
	assert.NilError(t, err)
	assert.Check(t, !out.APIToken.LastUsedAt.IsZero())
}
