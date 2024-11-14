package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryInvite() {
	t := suite.T()

	invite := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name    string
		queryID string
		client  *openlaneclient.OpenlaneClient
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "happy path",
			queryID: invite.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path with api token",
			queryID: invite.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "invalid id",
			queryID: "allthefooandbar",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
		{
			name:    "no access",
			queryID: invite.ID,
			client:  suite.client.api,
			ctx:     testUser2.UserCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetInviteByID(tc.ctx, tc.queryID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Invite)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateInvite() {
	t := suite.T()

	// existing user to invite to org
	existingUser := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// existing user already a member of org
	existingUser2 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	_ = (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID, UserID: existingUser2.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name             string
		recipient        string
		orgID            string
		role             enums.Role
		client           *openlaneclient.OpenlaneClient
		ctx              context.Context
		requestorID      string
		expectedStatus   enums.InviteStatus
		expectedAttempts int64
		wantErr          bool
	}{
		{
			name:             "happy path, new user as member",
			recipient:        "meow@theopenlane.io",
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 0,
			wantErr:          false,
		},
		{
			name:             "re-invite new user as member using api token",
			recipient:        "meow@theopenlane.io",
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.apiWithToken,
			ctx:              context.Background(),
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
			wantErr:          false,
		},
		{
			name:             "happy path, new user as admin using pat",
			recipient:        "woof@theopenlane.io",
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleAdmin,
			client:           suite.client.apiWithPAT,
			ctx:              context.Background(),
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 0,
			wantErr:          false,
		},
		{
			name:             "happy path, new user as member, by member",
			recipient:        "meow-meow@theopenlane.io",
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              viewOnlyUser.UserCtx,
			requestorID:      viewOnlyUser.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 0,
			wantErr:          false,
		},
		{
			name:        "new user as admin, by member, not allowed",
			recipient:   "meow-meow@theopenlane.io",
			orgID:       testUser1.OrganizationID,
			role:        enums.RoleAdmin,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			requestorID: viewOnlyUser.ID,
			wantErr:     true,
		},
		{
			name:      "new user as owner should fail",
			recipient: "woof@theopenlane.io",
			orgID:     testUser1.OrganizationID,
			role:      enums.RoleOwner,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
			wantErr:   true,
		},
		{
			name:             "happy path, existing user as member",
			recipient:        existingUser.Email,
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 0,
			wantErr:          false,
		},
		{
			name:             "user already a member, will still send an invite",
			recipient:        existingUser2.Email,
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 0,
			wantErr:          false,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			role := tc.role
			input := openlaneclient.CreateInviteInput{
				Recipient: tc.recipient,
				OwnerID:   &tc.orgID,
				Role:      &role,
			}

			resp, err := tc.client.CreateInvite(tc.ctx, input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Assert matching fields
			assert.Equal(t, tc.orgID, resp.CreateInvite.Invite.Owner.ID)
			assert.Equal(t, tc.role, resp.CreateInvite.Invite.Role)
			assert.Equal(t, tc.requestorID, *resp.CreateInvite.Invite.RequestorID)
			assert.Equal(t, tc.expectedStatus, resp.CreateInvite.Invite.Status)
			assert.Equal(t, tc.expectedAttempts, resp.CreateInvite.Invite.SendAttempts)
			assert.WithinDuration(t, time.Now().UTC().AddDate(0, 0, 14), *resp.CreateInvite.Invite.Expires, time.Minute)
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteInvite() {
	t := suite.T()

	invite1 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite2 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite3 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite4 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite5 := (&InviteBuilder{client: suite.client, Role: fgax.AdminRelation}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name    string
		queryID string
		client  *openlaneclient.OpenlaneClient
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "happy path",
			queryID: invite1.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: false,
		},
		{
			name:    "happy path, using api token",
			queryID: invite2.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "happy path, using personal access token",
			queryID: invite3.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "happy path, org member deleting member invite",
			queryID: invite4.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
			wantErr: false,
		},
		{
			name:    "org member deleting admin invite",
			queryID: invite5.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
			wantErr: true,
		},
		{
			name:    "org owner deleting admin invite",
			queryID: invite5.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: false,
		},
		{
			name:    "invalid id",
			queryID: "allthefooandbar",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteInvite(tc.ctx, tc.queryID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// assert equal
			assert.Equal(t, tc.queryID, resp.DeleteInvite.DeletedID)
		})
	}
}
