package resolvers_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi/testclient"

	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/testutils"
)

func TestQueryPersonalAccessToken(t *testing.T) {
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
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(redacted, resp.PersonalAccessToken.Token))
		})
	}

	// cleanup
	(&Cleanup[*generated.PersonalAccessTokenDeleteOne]{
		client: suite.client.db.PersonalAccessToken,
		ID:     token.ID,
	}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryPersonalAccessTokens(t *testing.T) {
	(&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create a token for another user
	(&PersonalAccessTokenBuilder{client: suite.client, OrganizationIDs: []string{testUser2.OrganizationID}}).MustNew(testUser2.UserCtx, t)

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
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.PersonalAccessTokens.Edges, 3)) // there is an additional token from the seed test data for this user
		})
	}
}

func TestMutationCreatePersonalAccessToken(t *testing.T) {
	tokenDescription := gofakeit.Sentence()
	expiration30Days := time.Now().Add(time.Hour * 24 * 30)

	testCases := []struct {
		name     string
		input    testclient.CreatePersonalAccessTokenInput
		errorMsg string
	}{
		{
			name: "happy path",
			input: testclient.CreatePersonalAccessTokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
			},
		},
		{
			name: "bad path, set expire to the past",
			input: testclient.CreatePersonalAccessTokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   lo.ToPtr(time.Now().Add(-time.Hour)),
			},
			errorMsg: hooks.ErrPastTimeNotAllowed.Error(),
		},
		{
			name: "happy path, set expire",
			input: testclient.CreatePersonalAccessTokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
				ExpiresAt:   &expiration30Days,
			},
		},
		{
			name: "happy path, set orgs",
			input: testclient.CreatePersonalAccessTokenInput{
				Name:            "forthethingz",
				Description:     &tokenDescription,
				ExpiresAt:       &expiration30Days,
				OrganizationIDs: []string{testUser1.OrganizationID, testUser1.PersonalOrgID},
			},
		},
		{
			name: "happy path, name only",
			input: testclient.CreatePersonalAccessTokenInput{
				Name: "forthethingz",
			},
		},
		{
			name: "empty name",
			input: testclient.CreatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
			errorMsg: "value is less than the required length",
		},
		{
			name: "setting other user id",
			input: testclient.CreatePersonalAccessTokenInput{
				Name:        "forthethingz",
				Description: &tokenDescription,
			},
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.CreatePersonalAccessToken(testUser1.UserCtx, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(resp.CreatePersonalAccessToken.PersonalAccessToken.Name, tc.input.Name))
			assert.Check(t, is.DeepEqual(resp.CreatePersonalAccessToken.PersonalAccessToken.Description, tc.input.Description))

			// check expiration if set
			if tc.input.ExpiresAt == nil {
				assert.Check(t, resp.CreatePersonalAccessToken.PersonalAccessToken.ExpiresAt == nil)
			} else {
				assert.Check(t, tc.input.ExpiresAt.Equal(*resp.CreatePersonalAccessToken.PersonalAccessToken.ExpiresAt))
			}

			// check organization is set if provided
			if tc.input.OrganizationIDs != nil {
				assert.Check(t, is.Len(resp.CreatePersonalAccessToken.PersonalAccessToken.Organizations.Edges, len(tc.input.OrganizationIDs)))

				for _, orgID := range resp.CreatePersonalAccessToken.PersonalAccessToken.Organizations.Edges {
					assert.Check(t, is.Contains(tc.input.OrganizationIDs, orgID.Node.ID))
				}
			} else {
				assert.Check(t, is.Len(resp.CreatePersonalAccessToken.PersonalAccessToken.Organizations.Edges, 0))
			}

			// ensure the owner is the user that made the request
			assert.Check(t, is.Equal(testUser1.ID, resp.CreatePersonalAccessToken.PersonalAccessToken.Owner.ID))

			// token should not be redacted on create
			assert.Check(t, redacted != resp.CreatePersonalAccessToken.PersonalAccessToken.Token)

			// ensure the token is prefixed
			assert.Check(t, is.Contains(resp.CreatePersonalAccessToken.PersonalAccessToken.Token, "tolp_"))

			// cleanup
			(&Cleanup[*generated.PersonalAccessTokenDeleteOne]{
				client: suite.client.db.PersonalAccessToken,
				ID:     resp.CreatePersonalAccessToken.PersonalAccessToken.ID,
			}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdatePersonalAccessToken(t *testing.T) {
	token := (&PersonalAccessTokenBuilder{
		client:          suite.client,
		OrganizationIDs: []string{testUser1.PersonalOrgID},
		ExpiresAt:       lo.ToPtr(time.Now().Add(time.Hour * 24 * 30))}).
		MustNew(testUser1.UserCtx, t)

	tokenOther := (&PersonalAccessTokenBuilder{
		client: suite.client, OrganizationIDs: []string{testUser2.OrganizationID}}).
		MustNew(testUser2.UserCtx, t)

	tokenDescription := gofakeit.Sentence()
	tokenName := gofakeit.Word()

	testCases := []struct {
		name     string
		tokenID  string
		input    testclient.UpdatePersonalAccessTokenInput
		errorMsg string
	}{
		{
			name:    "happy path, update name",
			tokenID: token.ID,
			input: testclient.UpdatePersonalAccessTokenInput{
				Name: &tokenName,
			},
		},
		{
			name:    "happy path, update description",
			tokenID: token.ID,
			input: testclient.UpdatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
		},
		{
			name:    "happy path, add org",
			tokenID: token.ID,
			input: testclient.UpdatePersonalAccessTokenInput{
				AddOrganizationIDs: []string{testUser1.OrganizationID},
			},
		},
		{
			name:    "happy path, remove org",
			tokenID: token.ID,
			input: testclient.UpdatePersonalAccessTokenInput{
				RemoveOrganizationIDs: []string{testUser1.OrganizationID},
			},
		},
		{
			name:    "invalid token id",
			tokenID: "notvalidtoken",
			input: testclient.UpdatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "not authorized",
			tokenID: tokenOther.ID,
			input: testclient.UpdatePersonalAccessTokenInput{
				Description: &tokenDescription,
			},
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.UpdatePersonalAccessToken(testUser1.UserCtx, tc.tokenID, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.input.Name != nil {
				assert.Check(t, is.Equal(resp.UpdatePersonalAccessToken.PersonalAccessToken.Name, *tc.input.Name))
			}

			if tc.input.Description != nil {
				assert.Check(t, is.DeepEqual(resp.UpdatePersonalAccessToken.PersonalAccessToken.Description, tc.input.Description))
			}

			// make sure these fields did not get updated
			if token.ExpiresAt != nil {
				assert.Assert(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.ExpiresAt != nil)
				diff := resp.UpdatePersonalAccessToken.PersonalAccessToken.ExpiresAt.Sub(*token.ExpiresAt)
				assert.Check(t, diff >= -1*time.Second && diff <= 1*time.Second, "time difference is not within 1 second, got %v", diff)
			} else {
				assert.Check(t, resp.UpdatePersonalAccessToken.PersonalAccessToken.ExpiresAt == nil)
			}

			assert.Check(t, is.Len(resp.UpdatePersonalAccessToken.PersonalAccessToken.Organizations.Edges, len(tc.input.AddOrganizationIDs)+1))

			// Ensure its removed
			if tc.input.RemoveOrganizationIDs != nil {
				assert.Check(t, is.Len(resp.UpdatePersonalAccessToken.PersonalAccessToken.Organizations.Edges, 1))
			}

			assert.Check(t, is.Equal(testUser1.ID, resp.UpdatePersonalAccessToken.PersonalAccessToken.Owner.ID))

			// token should be redacted on update
			assert.Check(t, is.Equal(redacted, resp.UpdatePersonalAccessToken.PersonalAccessToken.Token))
		})
	}

	// update expiration date
	_, err := suite.client.api.UpdatePersonalAccessToken(testUser1.UserCtx, token.ID, testclient.UpdatePersonalAccessTokenInput{
		ExpiresAt: lo.ToPtr(time.Now().Add(time.Hour)),
	})
	assert.NilError(t, err)

	// cleanup
	(&Cleanup[*generated.PersonalAccessTokenDeleteOne]{
		client: suite.client.db.PersonalAccessToken,
		ID:     token.ID,
	}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.PersonalAccessTokenDeleteOne]{
		client: suite.client.db.PersonalAccessToken,
		ID:     tokenOther.ID,
	}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationDeletePersonalAccessToken(t *testing.T) {
	token := (&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// token for another user
	tokenOther := (&PersonalAccessTokenBuilder{client: suite.client, OrganizationIDs: []string{testUser2.OrganizationID}}).MustNew(testUser2.UserCtx, t)

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
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.tokenID, resp.DeletePersonalAccessToken.DeletedID)
		})
	}

	// cleanup
	(&Cleanup[*generated.PersonalAccessTokenDeleteOne]{
		client: suite.client.db.PersonalAccessToken,
		ID:     tokenOther.ID,
	}).MustDelete(testUser2.UserCtx, t)
}

func TestLastUsedPersonalAccessToken(t *testing.T) {
	// create new personal access token
	token := (&PersonalAccessTokenBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// check that the last used is empty
	res, err := suite.client.api.GetPersonalAccessTokenByID(testUser1.UserCtx, token.ID)
	assert.NilError(t, err)
	assert.Check(t, res.PersonalAccessToken.LastUsedAt == nil)

	// setup graph client using the personal access token
	authHeader := openlaneclient.Authorization{
		BearerToken: token.Token,
	}

	graphClient, err := testutils.TestClientWithAuth(suite.client.db, suite.client.objectStore, openlaneclient.WithCredentials(authHeader))
	assert.NilError(t, err)

	// get the token to make sure the last used is updated using the token
	out, err := graphClient.GetPersonalAccessTokenByID(context.Background(), token.ID)
	assert.NilError(t, err)
	assert.Check(t, !out.PersonalAccessToken.LastUsedAt.IsZero())

	// cleanup
	(&Cleanup[*generated.PersonalAccessTokenDeleteOne]{
		client: suite.client.db.PersonalAccessToken,
		ID:     token.ID,
	}).MustDelete(testUser1.UserCtx, t)
}
