package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/testutils"
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
				Scopes:      []string{"read:evidence", "write:evidence"},
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
				Scopes: []string{"write:evidence"},
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
	token := (&APITokenBuilder{client: suite.client, Scopes: []string{"read:evidence", "read:api_token"}}).MustNew(testUser1.UserCtx, t)

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

func TestAPITokenScopeEnforcement(t *testing.T) {
	orgUser := suite.userBuilder(context.Background(), t)
	orgCtx := auth.NewTestContextWithOrgID(orgUser.ID, orgUser.OrganizationID)

	// create scoped tokens (read-only vs write)
	readToken := (&APITokenBuilder{client: suite.client, Scopes: []string{"read:organization"}}).MustNew(orgCtx, t)
	writeToken := (&APITokenBuilder{client: suite.client, Scopes: []string{"write:group"}}).MustNew(orgCtx, t)

	makeClient := func(token string) *testclient.TestClient {
		authHeader := openlaneclient.Authorization{
			BearerToken: token,
		}

		c, err := testutils.TestClientWithAuth(
			suite.client.db,
			suite.client.objectStore,
			openlaneclient.WithCredentials(authHeader),
		)
		requireNoError(err)

		return c
	}

	readClient := makeClient(readToken.Token)
	writeClient := makeClient(writeToken.Token)

	// read-only scope can fetch org details
	_, err := readClient.GetOrganizationByID(context.Background(), orgUser.OrganizationID)
	assert.NilError(t, err)

	// read-only scope cannot create a group (requires edit)
	_, err = readClient.CreateGroup(context.Background(), testclient.CreateGroupInput{
		Name: gofakeit.AppName(),
	})
	assert.ErrorContains(t, err, "permission denied")

	// write scope can create a group
	groupResp, err := writeClient.CreateGroup(context.Background(), testclient.CreateGroupInput{
		Name: gofakeit.AppName(),
	})
	assert.NilError(t, err)
	assert.Assert(t, groupResp != nil)
	assert.Check(t, groupResp.CreateGroup.Group.ID != "")

	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{groupResp.CreateGroup.Group.ID}}).MustDelete(orgCtx, t)
	(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, IDs: []string{readToken.ID, writeToken.ID}}).MustDelete(orgCtx, t)
}

func TestAPITokenObjectScopeTuples(t *testing.T) {
	orgUser := suite.userBuilder(context.Background(), t)
	orgCtx := auth.NewTestContextWithOrgID(orgUser.ID, orgUser.OrganizationID)

	evidence := (&EvidenceBuilder{client: suite.client}).MustNew(orgCtx, t)

	var tokensToCleanup []string

	defer (&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, IDs: []string{evidence.ID}}).MustDelete(orgCtx, t)
	defer func() {
		if len(tokensToCleanup) > 0 {
			(&Cleanup[*generated.APITokenDeleteOne]{client: suite.client.db.APIToken, IDs: tokensToCleanup}).MustDelete(orgCtx, t)
		}
	}()

	makeTokenClient := func(scopes []string) (*testclient.APIToken, *testclient.TestClient) {
		resp, err := suite.client.api.CreateAPIToken(orgCtx, testclient.CreateAPITokenInput{
			Name:   gofakeit.AppName(),
			Scopes: scopes,
		})
		assert.NilError(t, err)

		token := resp.CreateAPIToken.APIToken
		tokensToCleanup = append(tokensToCleanup, token.ID)

		authHeader := openlaneclient.Authorization{
			BearerToken: token.Token,
		}

		client, err := testutils.TestClientWithAuth(
			suite.client.db,
			suite.client.objectStore,
			openlaneclient.WithCredentials(authHeader),
		)
		assert.NilError(t, err)

		apiToken := &testclient.APIToken{
			ID:          token.ID,
			Name:        token.Name,
			Description: token.Description,
			Token:       token.Token,
			Scopes:      token.Scopes,
			ExpiresAt:   token.ExpiresAt,
			OwnerID:     token.OwnerID,
			LastUsedAt:  token.LastUsedAt,
		}

		return apiToken, client
	}

	listScopedOrgIDs := func(tokenID string, relation string) []string {
		resp, err := suite.client.db.Authz.ListObjectsRequest(context.Background(), fgax.ListRequest{
			SubjectID:   tokenID,
			SubjectType: auth.ServiceSubjectType,
			Relation:    relation,
			ObjectType:  generated.TypeOrganization,
		})
		assert.NilError(t, err)

		ids, err := fgax.GetEntityIDs(resp)
		assert.NilError(t, err)

		return ids
	}

	viewRelation := fgamodel.NormalizeScope("read:evidence")
	editRelation := fgamodel.NormalizeScope("write:evidence")

	t.Run("read-only evidence scope", func(t *testing.T) {
		token, client := makeTokenClient([]string{"read:evidence"})

		ids := listScopedOrgIDs(token.ID, viewRelation)
		assert.Check(t, lo.Contains(ids, orgUser.OrganizationID))

		ids = listScopedOrgIDs(token.ID, editRelation)
		assert.Check(t, !lo.Contains(ids, orgUser.OrganizationID))

		_, err := client.GetEvidenceByID(context.Background(), evidence.ID)
		assert.NilError(t, err)

		_, err = client.UpdateEvidence(context.Background(), evidence.ID, testclient.UpdateEvidenceInput{
			Name: lo.ToPtr(gofakeit.Word()),
		}, nil)
		assert.ErrorContains(t, err, "permission denied")
	})

	t.Run("scope addition and removal update tuples", func(t *testing.T) {
		token, client := makeTokenClient([]string{"read:evidence"})

		assert.Check(t, lo.Contains(listScopedOrgIDs(token.ID, viewRelation), orgUser.OrganizationID))
		assert.Check(t, !lo.Contains(listScopedOrgIDs(token.ID, editRelation), orgUser.OrganizationID))

		_, err := suite.client.api.UpdateAPIToken(orgCtx, token.ID, testclient.UpdateAPITokenInput{
			AppendScopes: []string{"write:evidence"},
		})
		assert.NilError(t, err)

		assert.Check(t, lo.Contains(listScopedOrgIDs(token.ID, editRelation), orgUser.OrganizationID))

		updatedName := gofakeit.Word()
		_, err = client.UpdateEvidence(context.Background(), evidence.ID, testclient.UpdateEvidenceInput{
			Name: &updatedName,
		}, nil)
		assert.NilError(t, err)

		_, err = suite.client.api.UpdateAPIToken(orgCtx, token.ID, testclient.UpdateAPITokenInput{
			Scopes: []string{"read:evidence"},
		})
		assert.NilError(t, err)

		assert.Check(t, !lo.Contains(listScopedOrgIDs(token.ID, editRelation), orgUser.OrganizationID))

		_, err = client.UpdateEvidence(context.Background(), evidence.ID, testclient.UpdateEvidenceInput{
			Name: lo.ToPtr(gofakeit.Word()),
		}, nil)
		assert.ErrorContains(t, err, "permission denied")
	})
}
