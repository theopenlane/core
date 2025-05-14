package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func TestQueryInvite(t *testing.T) {
	invite := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite2 := (&InviteBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name    string
		queryID string
		client  openlaneclient.OpenlaneClient
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
			name:    "api token, no access",
			queryID: invite2.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: true,
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

				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.Invite.ID != "")
		})
	}

	// delete created invite
	(&Cleanup[*generated.InviteDeleteOne]{client: suite.client.db.Invite, ID: invite.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.InviteDeleteOne]{client: suite.client.db.Invite, ID: invite2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateInvite(t *testing.T) {
	// existing user to invite to org
	existingUser := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// existing user already a member of org
	existingUser2 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	om := (&OrgMemberBuilder{client: suite.client, UserID: existingUser2.ID}).MustNew(testUser1.UserCtx, t)

	orgWithRestrictions := (&OrganizationBuilder{client: suite.client, AllowedDomains: []string{"meow.net"}}).MustNew(testUserCreator.UserCtx, t)

	orgWithRestrictionsCtx := auth.NewTestContextWithOrgID(testUserCreator.ID, orgWithRestrictions.ID)

	user1Context := auth.NewTestContextWithOrgID(testUserCreator.ID, testUserCreator.OrganizationID)

	testCases := []struct {
		name             string
		recipient        string
		orgID            string
		role             enums.Role
		client           openlaneclient.OpenlaneClient
		ctx              context.Context
		requestorID      string
		expectedStatus   enums.InviteStatus
		expectedAttempts int64
		wantErr          bool
	}{
		{
			name:             "happy path, new user as member",
			recipient:        "meow@theopenlane.io",
			orgID:            testUserCreator.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              user1Context,
			requestorID:      testUserCreator.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
			wantErr:          false,
		},
		{
			name:             "happy path, new user as member in restricted domain org",
			recipient:        "meow@meow.net",
			orgID:            orgWithRestrictions.ID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              orgWithRestrictionsCtx,
			requestorID:      testUserCreator.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
			wantErr:          false,
		},
		{
			name:        "new user as member in restricted domain org with invalid domain",
			recipient:   "meow@meow.io",
			orgID:       orgWithRestrictions.ID,
			role:        enums.RoleMember,
			client:      suite.client.api,
			ctx:         orgWithRestrictionsCtx,
			requestorID: testUserCreator.ID,
			wantErr:     true,
		},
		{
			name:             "invite new user as member using api token",
			recipient:        "meow@theopenlane.io",
			orgID:            testUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
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
			expectedAttempts: 2,
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
			expectedAttempts: 1,
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
			expectedAttempts: 1,
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
			ctx:       user1Context,
			wantErr:   true,
		},
		{
			name:      "new user with invalid email",
			recipient: "woof",
			orgID:     testUser1.OrganizationID,
			role:      enums.RoleOwner,
			client:    suite.client.api,
			ctx:       user1Context,
			wantErr:   true,
		},
		{
			name:             "happy path, existing user as member",
			recipient:        existingUser.Email,
			orgID:            testUserCreator.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              user1Context,
			requestorID:      testUserCreator.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
			wantErr:          false,
		},
		{
			name:             "user already a member, will still send an invite",
			recipient:        existingUser2.Email,
			orgID:            testUserCreator.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.client.api,
			ctx:              user1Context,
			requestorID:      testUserCreator.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
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
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Assert matching fields
			assert.Check(t, is.Equal(tc.orgID, resp.CreateInvite.Invite.Owner.ID))
			assert.Check(t, is.Equal(tc.role, resp.CreateInvite.Invite.Role))
			assert.Check(t, is.Equal(tc.requestorID, *resp.CreateInvite.Invite.RequestorID))
			assert.Check(t, is.Equal(tc.expectedStatus, resp.CreateInvite.Invite.Status))
			assert.Check(t, is.Equal(tc.expectedAttempts, resp.CreateInvite.Invite.SendAttempts))

			assert.Assert(t, resp.CreateInvite.Invite.Expires != nil)
			diff := resp.CreateInvite.Invite.Expires.Sub(time.Now().UTC().AddDate(0, 0, 14))
			assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
		})
	}

	// delete organization created
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, ID: orgWithRestrictions.ID}).MustDelete(testUserCreator.UserCtx, t)
	// delete org member created
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: om.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateBulkInvite(t *testing.T) {
	invites := []string{}
	testCases := []struct {
		name             string
		recipients       []string
		client           openlaneclient.OpenlaneClient
		ctx              context.Context
		requestorID      string
		expectedStatus   enums.InviteStatus
		expectedAttempts int64
		wantErr          bool
	}{
		{
			name:             "happy path, new user with defaults",
			recipients:       []string{"meow-meow-meow@theopenlane.io", "kitty@theopenlane.io"},
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
			wantErr:          false,
		},
		{
			name:             "happy path, resend with defaults",
			recipients:       []string{"meow-meow-meow@theopenlane.io", "kitty@theopenlane.io"},
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 2,
			wantErr:          false,
		},
		{
			name:             "happy path, resend again with defaults",
			recipients:       []string{"meow-meow-meow@theopenlane.io", "kitty@theopenlane.io"},
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			requestorID:      testUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 3,
			wantErr:          false,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			input := []*openlaneclient.CreateInviteInput{}

			for _, recipient := range tc.recipients {
				input = append(input, &openlaneclient.CreateInviteInput{
					Recipient: recipient,
				})
			}

			resp, err := tc.client.CreateBulkInvite(tc.ctx, input)
			if tc.wantErr {

				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.CreateBulkInvite.Invites, len(tc.recipients)))

			for _, invite := range resp.CreateBulkInvite.Invites {
				assert.Check(t, is.Equal(enums.RoleMember, invite.Role))
				assert.Check(t, is.Equal(testUser1.ID, *invite.RequestorID))
				assert.Check(t, is.Equal(tc.expectedStatus, invite.Status))
				assert.Check(t, is.Equal(tc.expectedAttempts, invite.SendAttempts))
			}

			// delete created invites
			invites := []string{}
			for _, invite := range resp.CreateBulkInvite.Invites {
				invites = append(invites, invite.ID)
			}
		})
	}

	(&Cleanup[*generated.InviteDeleteOne]{client: suite.client.db.Invite, IDs: invites}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteInvite(t *testing.T) {
	invite1 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite2 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite3 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite4 := (&InviteBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	invite5 := (&InviteBuilder{client: suite.client, Role: fgax.AdminRelation}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name    string
		queryID string
		client  openlaneclient.OpenlaneClient
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

				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// assert equal
			assert.Check(t, is.Equal(tc.queryID, resp.DeleteInvite.DeletedID))
		})
	}
}
