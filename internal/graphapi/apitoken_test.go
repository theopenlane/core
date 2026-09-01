package graphapi_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	fgamodel "github.com/theopenlane/core/v2/fga/model"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/testutils"
)

func TestQueryApiToken(t *testing.T) {
	apiToken := (&th.APITokenBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: apiToken.ID,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:     "not found, no access",
			queryID:  apiToken.ID,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     th.NotFoundErrorMsg,
			queryID:  "notfound",
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.GetAPITokenByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(th.Redacted, resp.APIToken.Token))
			assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, resp.APIToken.Owner.ID))
		})
	}

	(&th.Cleanup[*generated.APITokenDeleteOne]{Client: suite.Client.DB.APIToken, ID: apiToken.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryAPITokens(t *testing.T) {
	token1 := (&th.APITokenBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	token2 := (&th.APITokenBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			resp, err := suite.Client.API.GetAllAPITokens(th.SharedTestUser1.UserCtx)

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

	(&th.Cleanup[*generated.APITokenDeleteOne]{Client: suite.Client.DB.APIToken, ID: token1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.APITokenDeleteOne]{Client: suite.Client.DB.APIToken, ID: token2.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
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
				Scopes:      []string{"evidence:read", "evidence:write"},
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
			resp, err := suite.Client.API.CreateAPIToken(th.SharedTestUser1.UserCtx, tc.input)

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
			assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, *resp.CreateAPIToken.APIToken.OwnerID))

			// token should not be th.Redacted on create
			assert.Check(t, th.Redacted != resp.CreateAPIToken.APIToken.Token)

			// ensure the token is prefixed
			assert.Check(t, is.Contains(resp.CreateAPIToken.APIToken.Token, "tola_"))

			(&th.Cleanup[*generated.APITokenDeleteOne]{Client: suite.Client.DB.APIToken, ID: resp.CreateAPIToken.APIToken.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateAPIToken(t *testing.T) {
	token := (&th.APITokenBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			ctx: th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, update expiration",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Name:      &tokenName,
				ExpiresAt: lo.ToPtr(time.Now().Add(time.Hour)),
			},
			ctx: th.SharedTestUser1.UserCtx,
		},
		{
			name:    "update name, no access",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Name: &tokenName,
			},
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:    "happy path, update description",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Description: &tokenDescription,
			},
			ctx: th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, add scope",
			tokenID: token.ID,
			input: testclient.UpdateAPITokenInput{
				Scopes: []string{"evidence:write"},
			},
			ctx: th.SharedTestUser1.UserCtx,
		},
		{
			name:    "invalid token id",
			tokenID: "notvalidtoken",
			input: testclient.UpdateAPITokenInput{
				Description: &tokenDescription,
			},
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.UpdateAPIToken(tc.ctx, tc.tokenID, tc.input)

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

			assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, *resp.UpdateAPIToken.APIToken.OwnerID))

			// token should be th.Redacted on update
			assert.Check(t, is.Equal(th.Redacted, resp.UpdateAPIToken.APIToken.Token))
		})
	}

	(&th.Cleanup[*generated.APITokenDeleteOne]{Client: suite.Client.DB.APIToken, ID: token.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteAPIToken(t *testing.T) {
	// create user to make tokens
	user := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	user2 := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	orgID := user.Edges.Setting.Edges.DefaultOrg.ID
	orgID2 := user2.Edges.Setting.Edges.DefaultOrg.ID

	reqCtx := auth.NewTestContextWithOrgID(user.ID, orgID)

	token := (&th.APITokenBuilder{Client: suite.Client}).MustNew(reqCtx, t)

	reqCtx2 := auth.NewTestContextWithOrgID(user2.ID, orgID2)

	token2 := (&th.APITokenBuilder{Client: suite.Client}).MustNew(reqCtx2, t)

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
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.DeleteAPIToken(reqCtx, tc.tokenID)

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
	token := (&th.APITokenBuilder{Client: suite.Client, Scopes: []string{"evidence:read", "api_token:read"}}).MustNew(th.SharedTestUser1.UserCtx, t)

	// check that the last used is empty
	res, err := suite.Client.API.GetAPITokenByID(th.SharedTestUser1.UserCtx, token.ID)
	assert.NilError(t, err)
	assert.Check(t, res.APIToken.LastUsedAt == nil)

	// setup graph client using the API token
	authHeader := testclient.Authorization{
		BearerToken: token.Token,
	}

	graphClient, err := testutils.TestClientWithAuth(suite.Client.DB, suite.Client.ObjectStore,
		testclient.WithCredentials(authHeader),
	)
	assert.NilError(t, err)

	// get the token to make sure the last used is updated using the token
	out, err := graphClient.GetAPITokenByID(context.Background(), token.ID)
	assert.NilError(t, err)
	assert.Check(t, !out.APIToken.LastUsedAt.IsZero())
}

func TestAPITokenScopeEnforcement(t *testing.T) {
	t.Parallel()
	localTestUser := suite.SeedOrgOwner(t)
	orgCtx := auth.NewTestContextWithOrgID(localTestUser.Owner.ID, localTestUser.Owner.OrganizationID)

	// create scoped tokens (read-only vs write)
	// the non-obvious scopes are required because of the query being used in the test-client having edges to other fields
	readToken := (&th.APITokenBuilder{Client: suite.Client, Scopes: []string{"organization:read", "group:read", "org_subscription:read", "org_membership:read", "file:read"}}).MustNew(orgCtx, t)
	writeToken := (&th.APITokenBuilder{Client: suite.Client, Scopes: []string{"group:write"}}).MustNew(orgCtx, t)

	makeClient := func(token string) *testclient.TestClient {
		authHeader := testclient.Authorization{
			BearerToken: token,
		}

		c, err := testutils.TestClientWithAuth(
			suite.Client.DB,
			suite.Client.ObjectStore,
			testclient.WithCredentials(authHeader),
		)
		th.RequireNoError(t, err)

		return c
	}

	readClient := makeClient(readToken.Token)
	writeClient := makeClient(writeToken.Token)

	// read-only scope can fetch org details, this query includes groups so the token must have read:group scope as well
	_, err := readClient.GetOrganizationByID(context.Background(), localTestUser.Owner.OrganizationID)
	assert.NilError(t, err)

	// read-only scope cannot create a group (requires edit)
	_, err = readClient.CreateGroupSimple(context.Background(), testclient.CreateGroupInput{
		Name: gofakeit.AppName(),
	})
	assert.ErrorContains(t, err, th.MissingScopeErrorMsg)

	// write scope can create a group
	groupResp, err := writeClient.CreateGroupSimple(context.Background(), testclient.CreateGroupInput{
		Name: gofakeit.AppName(),
	})
	assert.NilError(t, err)
	assert.Assert(t, groupResp != nil)
	assert.Check(t, groupResp.CreateGroup.Group.ID != "")

	th.CleanupOrganizationDataWithContext(orgCtx, t)
}

func TestAPITokenObjectScopeTuples(t *testing.T) {
	t.Parallel()
	localTestUser := suite.SeedOrgOwner(t)
	orgCtx := auth.NewTestContextWithOrgID(localTestUser.Owner.ID, localTestUser.Owner.OrganizationID)
	orgUser := localTestUser.Owner

	evidence := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(orgCtx, t)

	var tokensToCleanup []string

	makeTokenClient := func(scopes []string) (*testclient.APIToken, *testclient.TestClient) {
		resp, err := suite.Client.API.CreateAPIToken(orgCtx, testclient.CreateAPITokenInput{
			Name:   gofakeit.AppName(),
			Scopes: scopes,
		})
		assert.NilError(t, err)

		token := resp.CreateAPIToken.APIToken
		tokensToCleanup = append(tokensToCleanup, token.ID)

		authHeader := testclient.Authorization{
			BearerToken: token.Token,
		}

		client, err := testutils.TestClientWithAuth(
			suite.Client.DB,
			suite.Client.ObjectStore,
			testclient.WithCredentials(authHeader),
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
		resp, err := suite.Client.DB.Authz.ListObjectsRequest(context.Background(), fgax.ListRequest{
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

	viewRelation := fgamodel.NormalizeScope("evidence:read")
	editRelation := fgamodel.NormalizeScope("evidence:write")

	t.Run("read-only evidence scope", func(t *testing.T) {
		token, client := makeTokenClient([]string{"evidence:read", "file:read", "control:read", "task:read", "subcontrol:read", "program:read", "control_objective:read", "internal_policy:read", "procedure:read"})

		ids := listScopedOrgIDs(token.ID, viewRelation)
		assert.Check(t, lo.Contains(ids, orgUser.OrganizationID))

		ids = listScopedOrgIDs(token.ID, editRelation)
		assert.Check(t, !lo.Contains(ids, orgUser.OrganizationID))

		_, err := client.GetEvidenceByID(context.Background(), evidence.ID)
		assert.NilError(t, err)

		_, err = client.UpdateEvidence(context.Background(), evidence.ID, testclient.UpdateEvidenceInput{
			Name: lo.ToPtr(gofakeit.Word()),
		}, nil)
		assert.ErrorContains(t, err, th.MissingScopeErrorMsg)
	})

	t.Run("scope addition and removal update tuples", func(t *testing.T) {
		token, client := makeTokenClient([]string{"evidence:read", "file:read", "control:read", "task:read", "subcontrol:read", "program:read", "control_objective:read", "internal_policy:read", "procedure:read"})

		assert.Check(t, lo.Contains(listScopedOrgIDs(token.ID, viewRelation), orgUser.OrganizationID))
		assert.Check(t, !lo.Contains(listScopedOrgIDs(token.ID, editRelation), orgUser.OrganizationID))

		_, err := suite.Client.API.UpdateAPIToken(orgCtx, token.ID, testclient.UpdateAPITokenInput{
			AppendScopes: []string{"evidence:write"},
		})
		assert.NilError(t, err)

		assert.Check(t, lo.Contains(listScopedOrgIDs(token.ID, editRelation), orgUser.OrganizationID))

		updatedName := gofakeit.Word()
		_, err = client.UpdateEvidence(context.Background(), evidence.ID, testclient.UpdateEvidenceInput{
			Name: &updatedName,
		}, nil)
		assert.NilError(t, err)

		_, err = suite.Client.API.UpdateAPIToken(orgCtx, token.ID, testclient.UpdateAPITokenInput{
			Scopes: []string{"evidence:read"},
		})
		assert.NilError(t, err)

		assert.Check(t, !lo.Contains(listScopedOrgIDs(token.ID, editRelation), orgUser.OrganizationID))

		_, err = client.UpdateEvidence(context.Background(), evidence.ID, testclient.UpdateEvidenceInput{
			Name: lo.ToPtr(gofakeit.Word()),
		}, nil)
		assert.ErrorContains(t, err, th.MissingScopeErrorMsg)
	})

	th.CleanupOrganizationDataWithContext(orgCtx, t)
}
