package graphapi_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryInvite(t *testing.T) {
	invite := (&th.InviteBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	invite2 := (&th.InviteBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	testCases := []struct {
		name    string
		queryID string
		client  *testclient.TestClient
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "happy path",
			queryID: invite.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path with api token",
			queryID: invite.ID,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "api token, no access",
			queryID: invite2.ID,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "invalid id",
			queryID: "allthefooandbar",
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			wantErr: true,
		},
		{
			name:    "no access",
			queryID: invite.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser2.UserCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetInviteByID(tc.ctx, tc.queryID)

			if tc.wantErr {
				assert.ErrorContains(t, err, th.NotFoundErrorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.Invite.ID != "")
		})
	}

	// delete created invite
	(&th.Cleanup[*generated.InviteDeleteOne]{Client: suite.Client.DB.Invite, ID: invite.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InviteDeleteOne]{Client: suite.Client.DB.Invite, ID: invite2.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}

func TestMutationCreateInvite(t *testing.T) {
	// existing user to invite to org
	existingUser := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// existing user already a member of org
	existingUser2 := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	(&th.OrgMemberBuilder{Client: suite.Client, UserID: existingUser2.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create another organization
	localTestOrg := suite.UserBuilder(context.Background(), t)

	// setup one more with restrictions on allowed domains
	orgWithRestrictions := (&th.OrganizationBuilder{Client: suite.Client, AllowedDomains: []string{"meow.net"}}).MustNew(localTestOrg.UserCtx, t)

	orgWithRestrictionsCtx := auth.NewTestContextWithOrgID(localTestOrg.ID, orgWithRestrictions.ID)

	user1Context := localTestOrg.UserCtx

	// create a group to add to the invite
	meows := (&th.GroupBuilder{Client: suite.Client, Name: "meows"}).MustNew(user1Context, t)
	anotherMeows := (&th.GroupBuilder{Client: suite.Client, Name: "another-meows"}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name             string
		recipient        string
		orgID            string
		groupID          *string
		role             enums.Role
		client           *testclient.TestClient
		ctx              context.Context
		requestorID      string
		expectedStatus   enums.InviteStatus
		expectedAttempts int64
		expectedErr      string
	}{
		{
			name:             "happy path, new user as member with a group set",
			recipient:        "meow@theopenlane.io",
			orgID:            localTestOrg.OrganizationID,
			groupID:          &meows.ID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              user1Context,
			requestorID:      localTestOrg.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:             "happy path, another new user as member with a group set",
			recipient:        "meowmeow@theopenlane.io",
			orgID:            localTestOrg.OrganizationID,
			groupID:          &meows.ID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              user1Context,
			requestorID:      localTestOrg.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:        "new user as member, with invalid group",
			recipient:   "meow-another@theopenlane.io",
			orgID:       localTestOrg.OrganizationID,
			groupID:     &anotherMeows.ID,
			role:        enums.RoleMember,
			client:      suite.Client.API,
			ctx:         user1Context,
			requestorID: localTestOrg.ID,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:             "happy path, new user as member in restricted domain org",
			recipient:        "meow@meow.net",
			orgID:            orgWithRestrictions.ID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              orgWithRestrictionsCtx,
			requestorID:      localTestOrg.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:             "new user as member in allowed domains set, direct invite to another org should be allowed",
			recipient:        "meow@meow.io",
			orgID:            orgWithRestrictions.ID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              orgWithRestrictionsCtx,
			requestorID:      localTestOrg.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:             "invite new user as member using api token",
			recipient:        "meow@theopenlane.io",
			orgID:            th.SharedTestUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			requestorID:      th.SharedTestUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:             "re-invite new user as member using api token",
			recipient:        "meow@theopenlane.io",
			orgID:            th.SharedTestUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.Client.APIWithToken,
			ctx:              context.Background(),
			requestorID:      th.SharedTestUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 2,
		},
		{
			name:             "happy path, new user as admin using pat",
			recipient:        "woof@theopenlane.io",
			orgID:            th.SharedTestUser1.OrganizationID,
			role:             enums.RoleAdmin,
			client:           suite.Client.APIWithPAT,
			ctx:              context.Background(),
			requestorID:      th.SharedTestUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:             "happy path, new user as member, by member",
			recipient:        "meow-meow@theopenlane.io",
			orgID:            th.SharedTestUser1.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              th.SharedViewOnlyUser.UserCtx,
			requestorID:      th.SharedViewOnlyUser.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:        "new user as admin, by member, not allowed",
			recipient:   "meow-meow@theopenlane.io",
			orgID:       th.SharedTestUser1.OrganizationID,
			role:        enums.RoleAdmin,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			requestorID: th.SharedViewOnlyUser.ID,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "new user with invalid email",
			recipient:   "woof",
			orgID:       th.SharedTestUser1.OrganizationID,
			role:        enums.RoleMember,
			client:      suite.Client.API,
			ctx:         user1Context,
			expectedErr: th.InvalidInputErrorMsg,
		},
		{
			name:             "happy path, existing user as member",
			recipient:        existingUser.Email,
			orgID:            localTestOrg.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              user1Context,
			requestorID:      localTestOrg.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
		{
			name:             "user already a member, will still send an invite",
			recipient:        existingUser2.Email,
			orgID:            localTestOrg.OrganizationID,
			role:             enums.RoleMember,
			client:           suite.Client.API,
			ctx:              user1Context,
			requestorID:      localTestOrg.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			role := tc.role
			input := testclient.CreateInviteInput{
				Recipient: tc.recipient,
				OwnerID:   &tc.orgID,
				Role:      &role,
			}

			if tc.groupID != nil {
				input.GroupIDs = []string{*tc.groupID}
			}

			resp, err := tc.client.CreateInvite(tc.ctx, input)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

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

			if tc.groupID != nil {
				assert.Check(t, is.Len(resp.CreateInvite.Invite.Groups.Edges, 1))
			} else {
				assert.Check(t, is.Len(resp.CreateInvite.Invite.Groups.Edges, 0))
			}

			assert.Assert(t, resp.CreateInvite.Invite.Expires != nil)
			diff := resp.CreateInvite.Invite.Expires.Sub(time.Now().UTC().AddDate(0, 0, 14))
			assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
		})
	}

	// delete organization created
	th.CleanupOrganizationDataWithContext(localTestOrg.UserCtx, t)
	th.CleanupOrganizationDataWithContext(orgWithRestrictionsCtx, t)
}

func TestMutationCreateBulkInvite(t *testing.T) {
	invites := []string{}
	testCases := []struct {
		name             string
		recipients       []string
		client           *testclient.TestClient
		ctx              context.Context
		requestorID      string
		expectedStatus   enums.InviteStatus
		expectedAttempts int64
		wantErr          bool
	}{
		{
			name:             "happy path, new user with defaults",
			recipients:       []string{"meow-meow-meow@theopenlane.io", "kitty@theopenlane.io"},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			requestorID:      th.SharedTestUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 1,
			wantErr:          false,
		},
		{
			name:             "happy path, resend with defaults",
			recipients:       []string{"meow-meow-meow@theopenlane.io", "kitty@theopenlane.io"},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			requestorID:      th.SharedTestUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 2,
			wantErr:          false,
		},
		{
			name:             "happy path, resend again with defaults",
			recipients:       []string{"meow-meow-meow@theopenlane.io", "kitty@theopenlane.io"},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			requestorID:      th.SharedTestUser1.ID,
			expectedStatus:   enums.InvitationSent,
			expectedAttempts: 3,
			wantErr:          false,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			input := []*testclient.CreateInviteInput{}

			for _, recipient := range tc.recipients {
				input = append(input, &testclient.CreateInviteInput{
					Recipient: recipient,
				})
			}

			resp, err := tc.client.CreateBulkInvite(tc.ctx, input)
			if tc.wantErr {
				assert.ErrorContains(t, err, "failed to create invite")

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.CreateBulkInvite.Invites, len(tc.recipients)))

			for _, invite := range resp.CreateBulkInvite.Invites {
				assert.Check(t, is.Equal(enums.RoleMember, invite.Role))
				assert.Check(t, is.Equal(th.SharedTestUser1.ID, *invite.RequestorID))
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

	(&th.Cleanup[*generated.InviteDeleteOne]{Client: suite.Client.DB.Invite, IDs: invites}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteInvite(t *testing.T) {
	invite1 := (&th.InviteBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	invite2 := (&th.InviteBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	invite3 := (&th.InviteBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	invite4 := (&th.InviteBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	invite5 := (&th.InviteBuilder{Client: suite.Client, Role: fgax.AdminRelation}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path",
			queryID: invite1.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, using api token",
			queryID: invite2.ID,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, using personal access token",
			queryID: invite3.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, org member deleting member invite",
			queryID: invite4.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:        "org member deleting admin invite",
			queryID:     invite5.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:    "org owner deleting admin invite",
			queryID: invite5.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:        "invalid id",
			queryID:     "allthefooandbar",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteInvite(tc.ctx, tc.queryID)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// assert equal
			assert.Check(t, is.Equal(tc.queryID, resp.DeleteInvite.DeletedID))
		})
	}
}
