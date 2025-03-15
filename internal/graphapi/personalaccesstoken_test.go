package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/openlaneclient"

	"github.com/theopenlane/core/pkg/testutils"
)

func (suite *GraphTestSuite) TestQueryPersonalAccessToken() {
	t := suite.T()

	token := (&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path pat",
			queryID: token.ID,
			ctx:     testUser1.UserCtx,
		},
		{
			name:     notFoundErrorMsg,
			queryID:  "notfound",
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     notFoundErrorMsg,
			queryID:  "notfound",
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetPersonalAccessTokenByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.PersonalAccessToken)
			assert.Equal(t, redacted, resp.PersonalAccessToken.Token)
		})
	}
}

func (suite *GraphTestSuite) TestQueryPersonalAccessTokens() {
	t := suite.T()

	(&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create a token for another user
	(&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name     string
		errorMsg string
	}{
		{
			name: "happy path, all pats",
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetAllPersonalAccessTokens(testUser1.UserCtx)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.PersonalAccessTokens.Edges, 3) // there is an additional token from the seed test data for this user
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreatePersonalAccessToken() {
	t := suite.T()

	tokenDescription := gofakeit.Sentence(5)
	expiration30Days := time.Now().Add(time.Hour * 24 * 30)

	testCases := []struct {
		name     string
		input    openlaneclient.CreatePersonalAccessTokenInput
		errorMsg string
	}{
		{
			name: "happy path",
			input: openlaneclient.CreatePersonalAccessTokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
			},
		},
		{
			name: "happy path, set expire",
			input: openlaneclient.CreatePersonalAccessTokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   &expiration30Days,
			},
		},
		{
			name: "happy path, set orgs",
			input: openlaneclient.CreatePersonalAccessTokenInput{
				Name:            "forthethingz",
				Description:     &tokenDescription,
				ExpiresAt:       &expiration30Days,
				OrganizationIDs: []string{testUser1.OrganizationID, testUser1.PersonalOrgID},
			},
		},
		{
			name: "happy path, name only",
			input: openlaneclient.CreatePersonalAccessTokenInput{
				Name: "forthethingz",
			},
		},
		{
			name: "empty name",
			input: openlaneclient.CreatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
			errorMsg: "value is less than the required length",
		},
		{
			name: "setting other user id",
			input: openlaneclient.CreatePersonalAccessTokenInput{
				OwnerID:     testUser2.ID, // this should get ignored
				Name:        "forthethingz",
				Description: &tokenDescription,
			},
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.CreatePersonalAccessToken(testUser1.UserCtx, tc.input)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreatePersonalAccessToken.PersonalAccessToken)
			assert.Equal(t, resp.CreatePersonalAccessToken.PersonalAccessToken.Name, tc.input.Name)
			assert.Equal(t, resp.CreatePersonalAccessToken.PersonalAccessToken.Description, tc.input.Description)

			// check expiration if set
			if tc.input.ExpiresAt == nil {
				assert.Empty(t, resp.CreatePersonalAccessToken.PersonalAccessToken.ExpiresAt)
			} else {
				assert.True(t, tc.input.ExpiresAt.Equal(*resp.CreatePersonalAccessToken.PersonalAccessToken.ExpiresAt))
			}

			// check organization is set if provided
			if tc.input.OrganizationIDs != nil {
				assert.Len(t, resp.CreatePersonalAccessToken.PersonalAccessToken.Organizations, len(tc.input.OrganizationIDs))

				for _, orgID := range resp.CreatePersonalAccessToken.PersonalAccessToken.Organizations.Edges {
					assert.Contains(t, tc.input.OrganizationIDs, orgID.Node.ID)
				}
			} else {
				assert.Len(t, resp.CreatePersonalAccessToken.PersonalAccessToken.Organizations, 0)
			}

			// ensure the owner is the user that made the request
			assert.Equal(t, testUser1.ID, resp.CreatePersonalAccessToken.PersonalAccessToken.Owner.ID)

			// token should not be redacted on create
			assert.NotEqual(t, redacted, resp.CreatePersonalAccessToken.PersonalAccessToken.Token)

			// ensure the token is prefixed
			assert.Contains(t, resp.CreatePersonalAccessToken.PersonalAccessToken.Token, "tolp_")
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdatePersonalAccessToken() {
	t := suite.T()

	token := (&PersonalAccessTokenBuilder{
		client:          suite.client,
		OrganizationIDs: []string{testUser1.PersonalOrgID},
		ExpiresAt:       lo.ToPtr(time.Now().Add(time.Hour * 24 * 30))}).
		MustNew(testUser1.UserCtx, t)

	tokenOther := (&PersonalAccessTokenBuilder{
		client: suite.client}).
		MustNew(testUser2.UserCtx, t)

	tokenDescription := gofakeit.Sentence(5)
	tokenName := gofakeit.Word()

	testCases := []struct {
		name     string
		tokenID  string
		input    openlaneclient.UpdatePersonalAccessTokenInput
		errorMsg string
	}{
		{
			name:    "happy path, update name",
			tokenID: token.ID,
			input: openlaneclient.UpdatePersonalAccessTokenInput{
				Name: &tokenName,
			},
		},
		{
			name:    "happy path, update description",
			tokenID: token.ID,
			input: openlaneclient.UpdatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
		},
		{
			name:    "happy path, add org",
			tokenID: token.ID,
			input: openlaneclient.UpdatePersonalAccessTokenInput{
				AddOrganizationIDs: []string{testUser1.OrganizationID},
			},
		},
		{
			name:    "happy path, remove org",
			tokenID: token.ID,
			input: openlaneclient.UpdatePersonalAccessTokenInput{
				RemoveOrganizationIDs: []string{testUser1.OrganizationID},
			},
		},
		{
			name:    "invalid token id",
			tokenID: "notvalidtoken",
			input: openlaneclient.UpdatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "not authorized",
			tokenID: tokenOther.ID,
			input: openlaneclient.UpdatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.UpdatePersonalAccessToken(testUser1.UserCtx, tc.tokenID, tc.input)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdatePersonalAccessToken.PersonalAccessToken)

			if tc.input.Name != nil {
				assert.Equal(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.Name, *tc.input.Name)
			}

			if tc.input.Description != nil {
				assert.Equal(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.Description, tc.input.Description)
			}

			// make sure these fields did not get updated
			if token.ExpiresAt != nil {
				assert.WithinDuration(t, *token.ExpiresAt, *resp.UpdatePersonalAccessToken.PersonalAccessToken.ExpiresAt, 1*time.Second)
			} else {
				assert.Empty(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.ExpiresAt)
			}

			assert.Len(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.Organizations, len(tc.input.AddOrganizationIDs)+1)

			// Ensure its removed
			if tc.input.RemoveOrganizationIDs != nil {
				assert.Len(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.Organizations, 1)
			}

			assert.Equal(t, testUser1.ID, resp.UpdatePersonalAccessToken.PersonalAccessToken.Owner.ID)

			// token should be redacted on update
			assert.Equal(t, redacted, resp.UpdatePersonalAccessToken.PersonalAccessToken.Token)
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeletePersonalAccessToken() {
	t := suite.T()

	token := (&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// token for another user
	tokenOther := (&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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
			tokenID:  tokenOther.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.DeletePersonalAccessToken(testUser1.UserCtx, tc.tokenID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.tokenID, resp.DeletePersonalAccessToken.DeletedID)
		})
	}
}

func (suite *GraphTestSuite) TestLastUsedPersonalAccessToken() {
	t := suite.T()

	// create new personal access token
	token := (&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// check that the last used is empty
	res, err := suite.client.api.GetPersonalAccessTokenByID(testUser1.UserCtx, token.ID)
	require.NoError(t, err)
	assert.Empty(t, res.PersonalAccessToken.LastUsedAt)

	// setup graph client using the personal access token
	authHeader := openlaneclient.Authorization{
		BearerToken: token.Token,
	}

	graphClient, err := testutils.TestClientWithAuth(t, suite.client.db, suite.client.objectStore, openlaneclient.WithCredentials(authHeader))
	require.NoError(t, err)

	// get the token to make sure the last used is updated using the token
	out, err := graphClient.GetPersonalAccessTokenByID(context.Background(), token.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, out.PersonalAccessToken.LastUsedAt)
}
