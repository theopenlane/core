package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/testutils"
)

func (suite *GraphTestSuite) TestQueryApiToken() {
	t := suite.T()

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.APIToken)
			assert.Equal(t, redacted, resp.APIToken.Token)
			assert.Equal(t, testUser1.OrganizationID, resp.APIToken.Owner.ID)
		})
	}
}

func (suite *GraphTestSuite) TestQueryAPITokens() {
	t := suite.T()

	(&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// this is three because we create two tokens in the test
			// and there is one created in the suite setup
			assert.Len(t, resp.APITokens.Edges, 3)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateAPIToken() {
	t := suite.T()

	tokenDescription := gofakeit.Sentence(5)
	expiration30Days := time.Now().Add(time.Hour * 24 * 30)

	testCases := []struct {
		name     string
		input    openlaneclient.CreateAPITokenInput
		errorMsg string
	}{
		{
			name: "happy path",
			input: openlaneclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				Scopes:      []string{"read", "write"},
			},
		},
		{
			name: "happy path, set expire",
			input: openlaneclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   &expiration30Days,
			},
		},
		{
			name: "happy path, set org",
			input: openlaneclient.CreateAPITokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   &expiration30Days,
			},
		},
		{
			name: "happy path, name only",
			input: openlaneclient.CreateAPITokenInput{
				Name: "forthethingz",
			},
		},
		{
			name: "empty name",
			input: openlaneclient.CreateAPITokenInput{
				Description: &tokenDescription,
			},
			errorMsg: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.CreateAPIToken(testUser1.UserCtx, tc.input)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateAPIToken.APIToken)

			assert.Equal(t, tc.input.Name, resp.CreateAPIToken.APIToken.Name)
			assert.Equal(t, tc.input.Description, resp.CreateAPIToken.APIToken.Description)
			assert.Equal(t, tc.input.Scopes, resp.CreateAPIToken.APIToken.Scopes)

			// check expiration if set
			if tc.input.ExpiresAt == nil {
				assert.Empty(t, resp.CreateAPIToken.APIToken.ExpiresAt)
			} else {
				assert.True(t, tc.input.ExpiresAt.Equal(*resp.CreateAPIToken.APIToken.ExpiresAt))
			}

			// ensure the owner is the org set in the request
			assert.Equal(t, testUser1.OrganizationID, resp.CreateAPIToken.APIToken.Owner.ID)

			// token should not be redacted on create
			assert.NotEqual(t, redacted, resp.CreateAPIToken.APIToken.Token)

			// ensure the token is prefixed
			assert.Contains(t, resp.CreateAPIToken.APIToken.Token, "tola_")
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateAPIToken() {
	t := suite.T()

	token := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	tokenDescription := gofakeit.Sentence(5)
	tokenName := gofakeit.Word()

	testCases := []struct {
		name     string
		tokenID  string
		input    openlaneclient.UpdateAPITokenInput
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path, update name ",
			tokenID: token.ID,
			input: openlaneclient.UpdateAPITokenInput{
				Name: &tokenName,
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "update name, no access",
			tokenID: token.ID,
			input: openlaneclient.UpdateAPITokenInput{
				Name: &tokenName,
			},
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "happy path, update description",
			tokenID: token.ID,
			input: openlaneclient.UpdateAPITokenInput{
				Description: &tokenDescription,
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "happy path, add scope",
			tokenID: token.ID,
			input: openlaneclient.UpdateAPITokenInput{
				Scopes: []string{"write"},
			},
			ctx: testUser1.UserCtx,
		},
		{
			name:    "invalid token id",
			tokenID: "notvalidtoken",
			input: openlaneclient.UpdateAPITokenInput{
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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateAPIToken.APIToken)

			if tc.input.Name != nil {
				assert.Equal(t, resp.UpdateAPIToken.APIToken.Name, *tc.input.Name)
			}

			if tc.input.Description != nil {
				assert.Equal(t, resp.UpdateAPIToken.APIToken.Description, tc.input.Description)
			}

			// Ensure its added
			if tc.input.Scopes != nil {
				assert.Len(t, resp.UpdateAPIToken.APIToken.Scopes, 1)
			}

			assert.Equal(t, testUser1.OrganizationID, resp.UpdateAPIToken.APIToken.Owner.ID)

			// token should be redacted on update
			assert.Equal(t, redacted, resp.UpdateAPIToken.APIToken.Token)
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteAPIToken() {
	t := suite.T()

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.tokenID, resp.DeleteAPIToken.DeletedID)
		})
	}
}

func (suite *GraphTestSuite) TestLastUsedAPIToken() {
	t := suite.T()

	// create new API token
	token := (&APITokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// check that the last used is empty
	res, err := suite.client.api.GetAPITokenByID(testUser1.UserCtx, token.ID)
	require.NoError(t, err)
	assert.Empty(t, res.APIToken.LastUsedAt)

	// setup graph client using the API token
	authHeader := openlaneclient.Authorization{
		BearerToken: token.Token,
	}

	graphClient, err := testutils.TestClientWithAuth(t, suite.client.db, suite.client.objectStore, openlaneclient.WithCredentials(authHeader))
	require.NoError(t, err)

	// get the token to make sure the last used is updated using the token
	out, err := graphClient.GetAPITokenByID(context.Background(), token.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, out.APIToken.LastUsedAt)
}
