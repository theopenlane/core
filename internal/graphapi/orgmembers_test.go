package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryOrgMembers(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedFreshOrgUsers(t)
	org1Member := localTestOrg.member

	pm := (&ProgramMemberBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)

	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: localTestOrg.owner.OrganizationID}).MustNew(localTestOrg.owner.UserCtx, t)

	childReqCtx := auth.NewTestContextWithOrgID(localTestOrg.owner.ID, childOrg.ID)

	(&OrgMemberBuilder{client: suite.client}).MustNew(childReqCtx, t)
	(&OrgMemberBuilder{client: suite.client, UserID: org1Member.ID}).MustNew(childReqCtx, t)

	testCases := []struct {
		name                string
		queryID             string
		deleteProgramMember bool
		whereInput          *testclient.OrgMembershipWhereInput
		client              *testclient.TestClient
		ctx                 context.Context
		expectedLen         int
		expectErr           bool
	}{
		{
			name:        "happy path, get org members by org id",
			queryID:     localTestOrg.owner.OrganizationID,
			client:      suite.client.api,
			ctx:         localTestOrg.owner.UserCtx,
			expectedLen: 6,
		},
		{
			name:        "happy path, get org members by org id, member",
			queryID:     localTestOrg.owner.OrganizationID,
			client:      suite.client.api,
			ctx:         localTestOrg.member.UserCtx,
			expectedLen: 6,
		},
		{
			name:        "happy path, get org members by org id, auditor",
			queryID:     localTestOrg.owner.OrganizationID,
			client:      suite.client.api,
			ctx:         localTestOrg.auditor.UserCtx,
			expectedLen: 6,
		},
		{
			name:        "happy path, get org with parent members based on context",
			client:      suite.client.api,
			ctx:         childReqCtx,
			expectedLen: 7, // 2 from child org, 5 from parent org because we dedupe plus the program member
		},
		{
			name:    "where input, get members in program",
			queryID: localTestOrg.owner.OrganizationID,
			client:  suite.client.api,
			ctx:     localTestOrg.owner.UserCtx,
			whereInput: &testclient.OrgMembershipWhereInput{
				HasUserWith: []*testclient.UserWhereInput{
					{
						HasProgramMembershipsWith: []*testclient.ProgramMembershipWhereInput{
							{
								ProgramID: &pm.ProgramID,
							},
						},
					},
				},
			},
			expectedLen: 2, // owner and program member
		},
		{
			name:    "where input, get members not in program",
			queryID: localTestOrg.owner.OrganizationID,
			client:  suite.client.api,
			ctx:     localTestOrg.owner.UserCtx,
			whereInput: &testclient.OrgMembershipWhereInput{
				Not: &testclient.OrgMembershipWhereInput{
					HasUserWith: []*testclient.UserWhereInput{
						{
							HasProgramMembershipsWith: []*testclient.ProgramMembershipWhereInput{
								{
									ProgramID: &pm.ProgramID,
								},
							},
						},
					},
				},
			},
			expectedLen: 4,
		},
		{
			name:                "where input, get members in program, after deleting a member",
			deleteProgramMember: true,
			queryID:             localTestOrg.owner.OrganizationID,
			client:              suite.client.api,
			ctx:                 localTestOrg.owner.UserCtx,
			whereInput: &testclient.OrgMembershipWhereInput{
				HasUserWith: []*testclient.UserWhereInput{
					{
						HasProgramMembershipsWith: []*testclient.ProgramMembershipWhereInput{
							{
								ProgramID: &pm.ProgramID,
							},
						},
					},
				},
			},
			expectedLen: 1, // only the owner remains
		},
		{
			name:        "happy path, get org with parent members using org ID, only direct members will be returned",
			queryID:     childOrg.ID,
			client:      suite.client.api,
			ctx:         childReqCtx,
			expectedLen: 2, // only child org members will be returned
		},
		{
			name:        "no access",
			queryID:     localTestOrg.owner.OrganizationID,
			client:      suite.client.api,
			ctx:         sharedTestUser2.UserCtx,
			expectedLen: 0,
			expectErr:   false, // no org members returned
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.client.api,
			ctx:         localTestOrg.owner.UserCtx,
			expectedLen: 0,
			expectErr:   false, // no org members returned
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			if tc.deleteProgramMember {
				// delete the program member to test the where input
				_, err := tc.client.DeleteProgramMembership(tc.ctx, pm.ID)
				assert.NilError(t, err)
			}

			orgID := tc.queryID

			if tc.whereInput == nil {
				tc.whereInput = &testclient.OrgMembershipWhereInput{}
			}

			if orgID != "" {
				tc.whereInput.OrganizationID = &orgID
			}

			resp, err := tc.client.GetOrgMembersByOrgID(tc.ctx, tc.whereInput)

			if tc.expectErr {
				assert.Assert(t, err != nil)
				assert.Assert(t, is.Nil(resp))
				return
			}

			assert.NilError(t, err)

			if tc.expectedLen == 0 {
				assert.Check(t, is.Len(resp.OrgMemberships.Edges, 0))

				return
			}

			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.OrgMemberships.Edges, tc.expectedLen))
		})
	}

	// delete created org
	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
}

func TestMutationCreateOrgMembers(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedFreshOrgUsers(t)
	org1ID := localTestOrg.owner.OrganizationID

	userCtx := localTestOrg.owner.UserCtx
	personalOrgCtx := auth.NewTestContextWithOrgID(localTestOrg.owner.ID, localTestOrg.owner.PersonalOrgID)

	user1 := (&UserBuilder{client: suite.client}).MustNew(userCtx, t)
	user2 := (&UserBuilder{client: suite.client}).MustNew(userCtx, t)
	user3 := (&UserBuilder{client: suite.client, Email: "mitb2@anderson.io", FirstName: "FirstName!@"}).MustNew(userCtx, t)

	userWithValidDomain := (&UserBuilder{client: suite.client, Email: "matt@anderson.net"}).MustNew(userCtx, t)
	userWithInvalidDomain := (&UserBuilder{client: suite.client, Email: "mitb@example.com"}).MustNew(userCtx, t)

	orgWithRestrictions := (&OrganizationBuilder{client: suite.client, AllowedDomains: []string{"anderson.io", "anderson.net"}}).MustNew(localTestOrg.owner.UserCtx, t)
	otherOrgCtx := auth.NewTestContextWithOrgID(localTestOrg.owner.ID, orgWithRestrictions.ID)

	testCases := []struct {
		name   string
		orgID  string
		userID string
		role   enums.Role
		ctx    context.Context
		errMsg string
	}{
		{
			name:   "happy path, add admin",
			orgID:  org1ID,
			userID: user1.ID,
			ctx:    userCtx,
			role:   enums.RoleAdmin,
		},
		{
			name:   "happy path, add member",
			orgID:  orgWithRestrictions.ID,
			userID: userWithValidDomain.ID,
			ctx:    otherOrgCtx,
			role:   enums.RoleMember,
		},
		{
			// it will be a managed group so it passes
			name:   "happy path, add member with invalid name",
			orgID:  orgWithRestrictions.ID,
			userID: user3.ID,
			ctx:    otherOrgCtx,
			role:   enums.RoleMember,
		},
		{
			name:   "happy path, add member in org with allowed domains",
			orgID:  org1ID,
			userID: user2.ID,
			ctx:    userCtx,
			role:   enums.RoleMember,
		},
		{
			name:   "add member with invalid domain",
			orgID:  orgWithRestrictions.ID,
			userID: userWithInvalidDomain.ID,
			ctx:    otherOrgCtx,
			role:   enums.RoleMember,
			errMsg: "email domain not allowed in organization",
		},
		{
			name:   "duplicate user, different role",
			orgID:  org1ID,
			userID: user1.ID,
			role:   enums.RoleMember,
			ctx:    userCtx,
			errMsg: "already exists",
		},
		{
			name:   "cannot add self to organization",
			orgID:  org1ID,
			userID: sharedTestUser2.ID,
			role:   enums.RoleMember,
			ctx:    sharedTestUser2.UserCtx,
			errMsg: notFoundErrorMsg, // organization is not found because user does not have access to it
		},
		{
			name:   "add user to personal org not allowed",
			orgID:  localTestOrg.owner.PersonalOrgID,
			userID: user1.ID,
			role:   enums.RoleMember,
			ctx:    personalOrgCtx,
			errMsg: hooks.ErrPersonalOrgsNoMembers.Error(),
		},
		{
			name:   "invalid user",
			orgID:  org1ID,
			userID: ulids.New().String(),
			role:   enums.RoleMember,
			ctx:    userCtx,
			errMsg: "user not found",
		},
		{
			name:   "no access",
			orgID:  org1ID,
			userID: user1.ID,
			role:   enums.RoleMember,
			ctx:    sharedViewOnlyUser.UserCtx,
			errMsg: notAuthorizedErrorMsg,
		},
		{
			name:   "invalid role",
			orgID:  org1ID,
			userID: user1.ID,
			role:   enums.RoleInvalid,
			ctx:    userCtx,
			errMsg: "not a valid OrgMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			input := testclient.CreateOrgMembershipInput{
				OrganizationID: tc.orgID,
				UserID:         tc.userID,
				Role:           &tc.role,
			}

			resp, err := suite.client.api.AddUserToOrgWithRole(tc.ctx, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.userID, resp.CreateOrgMembership.OrgMembership.UserID))
			assert.Check(t, is.Equal(tc.orgID, resp.CreateOrgMembership.OrgMembership.OrganizationID))
			assert.Check(t, is.Equal(tc.role, resp.CreateOrgMembership.OrgMembership.Role))

			// make sure the user default org is set to the new org
			suite.assertDefaultOrgUpdate(sharedTestUser1.UserCtx, t, tc.userID, tc.orgID, true)
		})
	}

	// delete created org and users
	cleanupOrganizationDataWithContext(otherOrgCtx, t)
	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
}

func TestMutationUpdateOrgMembers(t *testing.T) {
	// create another user for this test
	// so it doesn't interfere with the other tests
	t.Parallel()

	localTestOrg := suite.seedOrgOwner(t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
	orgMembers, err := suite.client.api.GetOrgMembersByOrgID(localTestOrg.owner.UserCtx, &testclient.OrgMembershipWhereInput{
		OrganizationID: &localTestOrg.owner.OrganizationID,
	})
	assert.NilError(t, err)

	testUserOrgMember := ""

	for _, edge := range orgMembers.OrgMemberships.Edges {
		if edge.Node.UserID == localTestOrg.owner.ID {
			testUserOrgMember = edge.Node.ID
			break
		}
	}

	testCases := []struct {
		name        string
		orgMemberID string
		role        enums.Role
		errMsg      string
	}{
		{
			name:        "happy path, update to admin from member",
			orgMemberID: om.ID,
			role:        enums.RoleAdmin,
		},
		{
			name:        "happy path, update to member from admin",
			orgMemberID: om.ID,
			role:        enums.RoleMember,
		},
		{
			name:        "update to same role",
			orgMemberID: om.ID,
			role:        enums.RoleMember,
		},
		{
			name:        "update self from admin to member, not allowed",
			orgMemberID: testUserOrgMember,
			role:        enums.RoleMember,
			errMsg:      notAuthorizedErrorMsg,
		},
		{
			name:        "invalid role",
			orgMemberID: testUserOrgMember,
			role:        enums.RoleInvalid,
			errMsg:      "not a valid OrgMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			input := testclient.UpdateOrgMembershipInput{
				Role: &tc.role,
			}

			resp, err := suite.client.api.UpdateUserRoleInOrg(localTestOrg.owner.UserCtx, tc.orgMemberID, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.role, resp.UpdateOrgMembership.OrgMembership.Role))
		})
	}

	// delete created org members
	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
}

func TestMutationUpdateOrgMemberRole(t *testing.T) {
	t.Parallel()

	org := suite.seedFreshOrgUsers(t)
	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	user := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(org.owner.UserCtx, t, &user, enums.RoleMember, org.owner.OrganizationID)

	roleUpdateMember, err := suite.client.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.owner.OrganizationID),
			orgmembership.UserID(user.ID),
		).
		Only(allowCtx)
	assert.NilError(t, err)

	ownerMember, err := suite.client.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.owner.OrganizationID),
			orgmembership.UserID(org.owner.ID),
		).
		Only(allowCtx)
	assert.NilError(t, err)

	cases := []struct {
		name        string
		ctx         context.Context
		orgMemberID string
		role        enums.Role
		errMsg      string
	}{
		{
			name:        "admin can update member to admin",
			ctx:         org.admin.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleAdmin,
		},
		{
			name:        "admin cannot update member to super admin",
			ctx:         org.admin.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleSuperAdmin,
			errMsg:      notAuthorizedErrorMsg,
		},
		{
			name:        "member cannot update admin to member",
			ctx:         org.member.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleMember,
			errMsg:      notAuthorizedErrorMsg,
		},
		{
			name:        "owner role cannot be changed directly",
			ctx:         org.admin.UserCtx,
			orgMemberID: ownerMember.ID,
			role:        enums.RoleAdmin,
			errMsg:      hooks.ErrOrgOwnerCannotBeUpdated.Error(),
		},
		{
			name:        "owner role cannot be assigned directly",
			ctx:         org.owner.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleOwner,
			errMsg:      hooks.ErrOrgOwnerCannotBeUpdated.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := testclient.UpdateOrgMembershipInput{
				Role: &tc.role,
			}

			resp, err := suite.client.api.UpdateUserRoleInOrg(tc.ctx, tc.orgMemberID, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.role, resp.UpdateOrgMembership.OrgMembership.Role))
		})
	}

	cleanupOrganizationDataWithContext(org.owner.UserCtx, t)
}

func TestMutationDeleteOrgMembers(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedOrgOwner(t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
	adminOrgMember := (&OrgMemberBuilder{client: suite.client, Role: string(enums.RoleAdmin)}).MustNew(localTestOrg.owner.UserCtx, t)

	// create admin user context
	adminUserCtx := auth.NewTestContextWithOrgID(adminOrgMember.UserID, localTestOrg.owner.OrganizationID)

	resp, err := suite.client.api.RemoveUserFromOrg(localTestOrg.owner.UserCtx, om.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(om.ID, resp.DeleteOrgMembership.DeletedID))

	// make sure the user default org is not set to the deleted org
	suite.assertDefaultOrgUpdate(localTestOrg.owner.UserCtx, t, om.UserID, om.OrganizationID, false)

	// re-adding the user to the org should succeed since the org membership
	// is deleted and the managed group is properly cleaned up
	reAddResp, err := suite.client.api.AddUserToOrgWithRole(localTestOrg.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: localTestOrg.owner.OrganizationID,
		UserID:         om.UserID,
		Role:           &enums.RoleAdmin,
	})

	assert.NilError(t, err)
	assert.Assert(t, reAddResp != nil)

	// cant remove self from org and owners cannot be removed
	orgMembers, err := suite.client.api.GetOrgMembersByOrgID(localTestOrg.owner.UserCtx, &testclient.OrgMembershipWhereInput{
		OrganizationID: &localTestOrg.owner.OrganizationID,
	})
	assert.NilError(t, err)

	for _, edge := range orgMembers.OrgMemberships.Edges {
		// cannot delete self
		if edge.Node.UserID == sharedAdminUser.ID {
			_, err := suite.client.api.RemoveUserFromOrg(adminUserCtx, edge.Node.ID)
			assert.ErrorContains(t, err, notAuthorizedErrorMsg)
		}

		// organization owner cannot be deleted
		if edge.Node.UserID == localTestOrg.owner.ID {
			_, err = suite.client.api.RemoveUserFromOrg(adminUserCtx, edge.Node.ID)
			assert.ErrorContains(t, err, "organization owner cannot be deleted")
			break
		}
	}

	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
}

func (suite *GraphTestSuite) assertDefaultOrgUpdate(ctx context.Context, t *testing.T, userID, orgID string, isEqual bool) {
	// when an org membership is deleted, the user default org should be updated
	// we need to allow the request because this is not for the user making the request
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	where := testclient.UserSettingWhereInput{
		UserID: &userID,
	}

	userSettingResp, err := suite.client.api.GetUserSettings(allowCtx, where)
	assert.NilError(t, err)
	assert.Assert(t, userSettingResp != nil)
	assert.Check(t, is.Len(userSettingResp.UserSettings.Edges, 1))

	if isEqual {
		assert.Check(t, is.Equal(orgID, userSettingResp.UserSettings.Edges[0].Node.DefaultOrg.ID))
	} else {
		assert.Check(t, orgID != userSettingResp.UserSettings.Edges[0].Node.DefaultOrg.ID)
	}
}
