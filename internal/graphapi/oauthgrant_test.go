package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/oauthserver"
)

func TestMutationUpdateOauthGrant(t *testing.T) {
	activeGrant := (&th.OauthGrantBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	revokeGrant := (&th.OauthGrantBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	revokedGrant := (&th.OauthGrantBuilder{Client: suite.Client, Revoked: true}).MustNew(th.SharedTestUser1.UserCtx, t)

	activeToken := (&th.OauthTokenBuilder{Client: suite.Client, GrantID: activeGrant.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	revokeToken := (&th.OauthTokenBuilder{Client: suite.Client, GrantID: revokeGrant.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.OauthGrantDeleteOne]{Client: suite.Client.DB.OauthGrant, IDs: []string{activeGrant.ID, revokeGrant.ID, revokedGrant.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	testCases := []struct {
		name     string
		grantID  string
		tokenID  string
		input    testclient.UpdateOauthGrantInput
		ctx      context.Context
		revoked  bool
		errorMsg string
	}{
		{
			name:    "setting is active on an active grant with other changes",
			grantID: activeGrant.ID,
			tokenID: activeToken.ID,
			input: testclient.UpdateOauthGrantInput{
				IsActive: new(true),
				Tags:     []string{"mcp"},
			},
			ctx: th.SharedTestUser1.UserCtx,
		},
		{
			name:    "revoke grant",
			grantID: revokeGrant.ID,
			tokenID: revokeToken.ID,
			input: testclient.UpdateOauthGrantInput{
				IsActive:      new(false),
				RevokedReason: new("no longer used"),
			},
			ctx:     th.SharedTestUser1.UserCtx,
			revoked: true,
		},
		{
			name:    "update revoked grant",
			grantID: revokedGrant.ID,
			input: testclient.UpdateOauthGrantInput{
				Tags: []string{"mcp"},
			},
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: hooks.ErrOAuthGrantRevoked.Error(),
		},
		{
			name:    "reactivate revoked grant",
			grantID: revokedGrant.ID,
			input: testclient.UpdateOauthGrantInput{
				IsActive: new(true),
			},
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: hooks.ErrOAuthGrantRevoked.Error(),
		},
		{
			name:    "revoke revoked grant again",
			grantID: revokedGrant.ID,
			input: testclient.UpdateOauthGrantInput{
				IsActive: new(false),
			},
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: hooks.ErrOAuthGrantRevoked.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.UpdateOauthGrant(tc.ctx, tc.grantID, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			grant := resp.UpdateOauthGrant.OauthGrant
			assert.Check(t, is.Equal(grant.IsActive, !tc.revoked))

			token, err := suite.Client.DB.OauthToken.Get(th.SetInternalContext(th.SharedTestUser1.UserCtx, suite.Client.DB), tc.tokenID)
			assert.NilError(t, err)

			if tc.revoked {
				assert.Check(t, grant.RevokedAt != nil)
				assert.Check(t, is.DeepEqual(grant.RevokedBy, &th.SharedTestUser1.ID))
				assert.Check(t, token.ConsumedAt != nil)

				return
			}

			assert.Check(t, grant.RevokedAt == nil)
			assert.Check(t, token.ConsumedAt == nil)
		})
	}
}

func TestOauthGrantScopeEnforcement(t *testing.T) {
	grant := (&th.OauthGrantBuilder{Client: suite.Client, Scopes: []string{"control:write", "group:read", "program:read"}}).MustNew(th.SharedTestUser1.UserCtx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.OauthGrantDeleteOne]{Client: suite.Client.DB.OauthGrant, ID: grant.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	ctx := oauthserver.WithCallerGrant(th.SharedTestUser1.UserCtx, oauthserver.CallerGrant{
		ID:     grant.ID,
		Kind:   enums.OAuthGrantKindUser,
		Access: enums.OAuthAccessLevelWrite,
		Scopes: grant.Scopes,
	})

	control, err := suite.Client.API.CreateControl(ctx, testclient.CreateControlInput{RefCode: "CTL-" + ulids.New().String()})
	assert.NilError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.CreateControl.Control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	_, err = suite.Client.API.GetAllControls(ctx)
	assert.NilError(t, err)

	_, err = suite.Client.API.CreateRisk(ctx, testclient.CreateRiskInput{Name: gofakeit.Word()})
	assert.ErrorContains(t, err, th.MissingScopeErrorMsg)

	_, err = suite.Client.API.GetAllRisks(ctx, nil, nil, nil, nil, nil)
	assert.ErrorContains(t, err, th.MissingScopeErrorMsg)
}

func TestOauthGrantSoftDeleteRevokesTokens(t *testing.T) {
	grant := (&th.OauthGrantBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	token := (&th.OauthTokenBuilder{Client: suite.Client, GrantID: grant.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	internalCtx := th.SetInternalContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	t.Cleanup(func() {
		err := suite.Client.DB.OauthToken.DeleteOneID(token.ID).Exec(internalCtx)
		assert.NilError(t, err)
	})

	err := suite.Client.DB.OauthGrant.DeleteOneID(grant.ID).Exec(th.SetUserContext(th.SharedTestUser1.UserCtx, suite.Client.DB))
	assert.NilError(t, err)

	got, err := suite.Client.DB.OauthToken.Get(internalCtx, token.ID)
	assert.NilError(t, err)
	assert.Check(t, got.ConsumedAt != nil)
}
