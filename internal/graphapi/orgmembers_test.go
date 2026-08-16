package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/program"
	"github.com/theopenlane/core/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
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

			// no org role set, so should return empty array
			assert.Check(t, is.Len(resp.OrgMemberships.Edges[0].Node.AdditionalRoles, 0))
		})
	}

	// delete created org
	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
}

func TestQueryOrgMembersWithAdditionalRoles(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedFreshOrgUsers(t)
	org1Member := localTestOrg.member

	// add policy manager and trust center manager role
	suite.addFunctionalRoleForUser(localTestOrg.owner.UserCtx, t, org1Member.ID, localTestOrg.owner.OrganizationID, []string{"policy_manager", "trust_center_manager"})
	testCases := []struct {
		name                  string
		whereInput            *testclient.OrgMembershipWhereInput
		client                *testclient.TestClient
		ctx                   context.Context
		expectErr             bool
		expectAdditionalRoles bool
	}{
		{
			name: "happy path, get org member with additional roles",
			whereInput: &testclient.OrgMembershipWhereInput{
				UserID: &org1Member.ID,
			},
			client:                suite.client.api,
			ctx:                   localTestOrg.owner.UserCtx,
			expectAdditionalRoles: true,
		},
		{
			name: "happy path, get org auditor has no additional roles",
			whereInput: &testclient.OrgMembershipWhereInput{
				UserID: &localTestOrg.auditor.ID,
			},
			client:                suite.client.api,
			ctx:                   localTestOrg.owner.UserCtx,
			expectAdditionalRoles: false,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			if tc.whereInput == nil {
				tc.whereInput = &testclient.OrgMembershipWhereInput{}
			}

			resp, err := tc.client.GetOrgMembersByOrgID(tc.ctx, tc.whereInput)

			if tc.expectErr {
				assert.Assert(t, err != nil)
				assert.Assert(t, is.Nil(resp))
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Assert(t, is.Len(resp.OrgMemberships.Edges, 1))

			if tc.expectAdditionalRoles {
				assert.Check(t, is.Len(resp.OrgMemberships.Edges[0].Node.AdditionalRoles, 2))
				assert.Check(t, is.Contains(resp.OrgMemberships.Edges[0].Node.AdditionalRoles, "Policy Manager"))
				assert.Check(t, is.Contains(resp.OrgMemberships.Edges[0].Node.AdditionalRoles, "Trust Center Manager"))
			} else {
				assert.Check(t, is.Len(resp.OrgMemberships.Edges[0].Node.AdditionalRoles, 0))
			}
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
	userWithAnotherDomain := (&UserBuilder{client: suite.client, Email: "mitb@example.com"}).MustNew(userCtx, t)

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
			name:   "add member with another domain, allowed because allowed domains is only enforce for auto join",
			orgID:  orgWithRestrictions.ID,
			userID: userWithAnotherDomain.ID,
			ctx:    otherOrgCtx,
			role:   enums.RoleAuditor,
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
			errMsg: "constraint failed",
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
			errMsg:      hooks.ErrOrgOwnerCannotBeUpdated.Error(),
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

func TestMutationBulkUpdateOrgMemberRole(t *testing.T) {
	t.Parallel()

	org := suite.seedFreshOrgUsers(t)
	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	user1 := suite.userBuilder(context.Background(), t)
	user2 := suite.userBuilder(context.Background(), t)

	suite.addUserToOrganization(org.owner.UserCtx, t, &user1, enums.RoleMember, org.owner.OrganizationID)
	suite.addUserToOrganization(org.owner.UserCtx, t, &user2, enums.RoleMember, org.owner.OrganizationID)

	currentMembers, err := suite.client.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.owner.OrganizationID),
			orgmembership.UserIDIn(user1.ID, user2.ID),
		).
		All(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(currentMembers, 2))

	ids := make([]string, len(currentMembers))
	for i, member := range currentMembers {
		ids[i] = member.ID
	}

	adminRole := enums.RoleAdmin

	input := testclient.UpdateOrgMembershipInput{
		Role: &adminRole,
	}

	resp, err := suite.client.api.UpdateBulkOrgMemberRoles(org.admin.UserCtx, ids, input)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Len(resp.UpdateBulkOrgMembership.UpdatedIDs, len(ids)))
	assert.Check(t, is.Len(resp.UpdateBulkOrgMembership.OrgMemberships, len(ids)))

	updatedMembers, err := suite.client.db.OrgMembership.Query().
		Where(orgmembership.IDIn(ids...)).
		All(allowCtx)
	assert.NilError(t, err)

	for _, member := range updatedMembers {
		assert.Check(t, is.Equal(enums.RoleAdmin, member.Role))
	}

	ownerMember, err := suite.client.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.owner.OrganizationID),
			orgmembership.UserID(org.owner.ID),
		).
		Only(allowCtx)
	assert.NilError(t, err)

	memberRole := enums.RoleMember
	input.Role = &memberRole

	ownerResp, err := suite.client.api.UpdateBulkOrgMemberRoles(org.admin.UserCtx, []string{ownerMember.ID}, input)
	assert.NilError(t, err)
	assert.Assert(t, ownerResp != nil)
	assert.Check(t, is.Len(ownerResp.UpdateBulkOrgMembership.UpdatedIDs, 0))
	assert.Check(t, is.Len(ownerResp.UpdateBulkOrgMembership.OrgMemberships, 0))

	ownerMember, err = suite.client.db.OrgMembership.Query().
		Where(orgmembership.ID(ownerMember.ID)).
		Only(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.RoleOwner, ownerMember.Role))

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

func TestMutationLeaveOrganization(t *testing.T) {
	t.Parallel()

	currentOrg := suite.seedOrgOwner(t)
	orgToLeave := suite.seedOrgOwner(t)

	memberRole := enums.RoleMember
	member, err := suite.client.api.AddUserToOrgWithRole(orgToLeave.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToLeave.owner.OrganizationID,
		UserID:         currentOrg.owner.ID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	resp, err := suite.client.api.LeaveOrganization(currentOrg.owner.UserCtx, orgToLeave.owner.OrganizationID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(member.CreateOrgMembership.OrgMembership.ID, resp.LeaveOrganization.DeletedID))

	members, err := suite.client.api.GetOrgMembersByOrgID(orgToLeave.owner.UserCtx, &testclient.OrgMembershipWhereInput{
		OrganizationID: &orgToLeave.owner.OrganizationID,
		UserID:         &currentOrg.owner.ID,
	})
	assert.NilError(t, err)
	assert.Assert(t, members != nil)
	assert.Check(t, is.Len(members.OrgMemberships.Edges, 0))

	suite.assertDefaultOrgUpdate(currentOrg.owner.UserCtx, t, currentOrg.owner.ID, orgToLeave.owner.OrganizationID, false)

	cleanupOrganizationDataWithContext(currentOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(orgToLeave.owner.UserCtx, t)
}

func TestMutationLeaveOrganizationPreservesOtherOrgMemberships(t *testing.T) {
	t.Parallel()

	currentOrg := suite.seedOrgOwner(t)
	orgToLeave := suite.seedOrgOwner(t)

	userID := currentOrg.owner.ID

	memberRole := enums.RoleMember
	member, err := suite.client.api.AddUserToOrgWithRole(orgToLeave.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToLeave.owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	// creating a program adds the creator as a program admin, giving the user a
	// program membership in the org they are staying in
	(&ProgramBuilder{client: suite.client}).MustNew(currentOrg.owner.UserCtx, t)

	// the user should be a member of the managed groups in both orgs
	groupsBefore := suite.countOrgScopedGroupMemberships(t, userID, currentOrg.owner.OrganizationID)
	assert.Check(t, groupsBefore > 0)

	programsBefore := suite.countOrgScopedProgramMemberships(t, userID, currentOrg.owner.OrganizationID)
	assert.Check(t, programsBefore > 0)

	assert.Check(t, suite.countOrgScopedGroupMemberships(t, userID, orgToLeave.owner.OrganizationID) > 0)

	resp, err := suite.client.api.LeaveOrganization(currentOrg.owner.UserCtx, orgToLeave.owner.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// memberships in the org the user left should be removed
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, userID, orgToLeave.owner.OrganizationID)))
	assert.Check(t, is.Equal(0, suite.countOrgScopedProgramMemberships(t, userID, orgToLeave.owner.OrganizationID)))

	// memberships in every other org must be left intact
	assert.Check(t, is.Equal(groupsBefore, suite.countOrgScopedGroupMemberships(t, userID, currentOrg.owner.OrganizationID)))
	assert.Check(t, is.Equal(programsBefore, suite.countOrgScopedProgramMemberships(t, userID, currentOrg.owner.OrganizationID)))

	// the user must specifically still be in the system managed groups of the org they stayed in
	suite.assertManagedGroupMembership(t, userID, currentOrg.owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, userID, currentOrg.owner.OrganizationID, hooks.AllMembersGroup)

	cleanupOrganizationDataWithContext(currentOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(orgToLeave.owner.UserCtx, t)
}

func TestMutationDeleteOrgMembersPreservesOtherOrgMemberships(t *testing.T) {
	t.Parallel()

	homeOrg := suite.seedOrgOwner(t)
	otherOrg := suite.seedOrgOwner(t)

	userID := homeOrg.owner.ID

	memberRole := enums.RoleMember
	member, err := suite.client.api.AddUserToOrgWithRole(otherOrg.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: otherOrg.owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	groupsBefore := suite.countOrgScopedGroupMemberships(t, userID, homeOrg.owner.OrganizationID)
	assert.Assert(t, groupsBefore > 0)

	// the other org's owner removes the user from their org
	resp, err := suite.client.api.RemoveUserFromOrg(otherOrg.owner.UserCtx, member.CreateOrgMembership.OrgMembership.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// memberships in the org the user was removed from should be gone
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, userID, otherOrg.owner.OrganizationID)))

	// memberships in the user's other org must be left intact
	assert.Check(t, is.Equal(groupsBefore, suite.countOrgScopedGroupMemberships(t, userID, homeOrg.owner.OrganizationID)))
	suite.assertManagedGroupMembership(t, userID, homeOrg.owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, userID, homeOrg.owner.OrganizationID, hooks.AllMembersGroup)

	cleanupOrganizationDataWithContext(homeOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(otherOrg.owner.UserCtx, t)
}

func TestMutationDeleteOrganizationPreservesOtherOrgMemberships(t *testing.T) {
	t.Parallel()

	homeOrg := suite.seedOrgOwner(t)
	orgToDelete := suite.seedOrgOwner(t)

	userID := homeOrg.owner.ID

	memberRole := enums.RoleMember
	member, err := suite.client.api.AddUserToOrgWithRole(orgToDelete.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToDelete.owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	groupsBefore := suite.countOrgScopedGroupMemberships(t, userID, homeOrg.owner.OrganizationID)
	assert.Assert(t, groupsBefore > 0)
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, userID, orgToDelete.owner.OrganizationID) > 0)

	// delete the entire organization the user is a member of
	resp, err := suite.client.api.DeleteOrganization(orgToDelete.owner.UserCtx, orgToDelete.owner.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// the deleted org's own groups and memberships are removed asynchronously by the
	// organization cascade delete listener, so only the cross-org invariant is asserted here:
	// memberships in the user's other org must be left intact
	assert.Check(t, is.Equal(groupsBefore, suite.countOrgScopedGroupMemberships(t, userID, homeOrg.owner.OrganizationID)))
	suite.assertManagedGroupMembership(t, userID, homeOrg.owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, userID, homeOrg.owner.OrganizationID, hooks.AllMembersGroup)

	cleanupOrganizationDataWithContext(homeOrg.owner.UserCtx, t)
}

func TestMutationDeleteBulkOrgMembers(t *testing.T) {
	t.Parallel()

	homeOrg := suite.seedOrgOwner(t)
	bulkOrg := suite.seedOrgOwner(t)

	crossOrgUserID := homeOrg.owner.ID

	// user with a membership in another org, to verify bulk removal stays org-scoped
	memberRole := enums.RoleMember
	crossOrgMember, err := suite.client.api.AddUserToOrgWithRole(bulkOrg.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: bulkOrg.owner.OrganizationID,
		UserID:         crossOrgUserID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)

	otherMember := (&OrgMemberBuilder{client: suite.client}).MustNew(bulkOrg.owner.UserCtx, t)

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	ownerMembershipID, err := suite.client.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(bulkOrg.owner.OrganizationID),
			orgmembership.UserID(bulkOrg.owner.ID),
		).
		OnlyID(allowCtx)
	assert.NilError(t, err)

	homeGroupsBefore := suite.countOrgScopedGroupMemberships(t, crossOrgUserID, homeOrg.owner.OrganizationID)
	assert.Assert(t, homeGroupsBefore > 0)
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, crossOrgUserID, bulkOrg.owner.OrganizationID) > 0)
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, otherMember.UserID, bulkOrg.owner.OrganizationID) > 0)

	crossOrgMembershipID := crossOrgMember.CreateOrgMembership.OrgMembership.ID

	resp, err := suite.client.api.RemoveBulkUsersFromOrg(bulkOrg.owner.UserCtx, []string{
		crossOrgMembershipID,
		otherMember.ID,
		ownerMembershipID,
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// the two regular members are removed, the owner is protected
	assert.Check(t, is.Contains(resp.DeleteBulkOrgMembership.DeletedIDs, crossOrgMembershipID))
	assert.Check(t, is.Contains(resp.DeleteBulkOrgMembership.DeletedIDs, otherMember.ID))
	assert.Check(t, is.Contains(resp.DeleteBulkOrgMembership.NotDeletedIDs, ownerMembershipID))
	assert.Check(t, resp.DeleteBulkOrgMembership.Error != nil)

	// group memberships in the bulk org are cleaned up for the removed members
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, crossOrgUserID, bulkOrg.owner.OrganizationID)))
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, otherMember.UserID, bulkOrg.owner.OrganizationID)))

	// the owner keeps their memberships in the bulk org
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, bulkOrg.owner.ID, bulkOrg.owner.OrganizationID) > 0)

	// memberships in the removed user's other org must be left intact
	assert.Check(t, is.Equal(homeGroupsBefore, suite.countOrgScopedGroupMemberships(t, crossOrgUserID, homeOrg.owner.OrganizationID)))
	suite.assertManagedGroupMembership(t, crossOrgUserID, homeOrg.owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, crossOrgUserID, homeOrg.owner.OrganizationID, hooks.AllMembersGroup)

	cleanupOrganizationDataWithContext(homeOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(bulkOrg.owner.UserCtx, t)
}

func TestMutationLeaveOrganizationReassignsDefaultOrgToMemberOrg(t *testing.T) {
	t.Parallel()

	user := suite.seedOrgOwner(t)
	orgToLeave := suite.seedOrgOwner(t)

	userID := user.owner.ID

	memberRole := enums.RoleMember
	member, err := suite.client.api.AddUserToOrgWithRole(orgToLeave.owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToLeave.owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	// make the org they are about to leave their default org
	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	_, err = suite.client.db.UserSetting.Update().
		Where(usersetting.UserID(userID)).
		SetDefaultOrgID(orgToLeave.owner.OrganizationID).
		Save(allowCtx)
	assert.NilError(t, err)

	resp, err := suite.client.api.LeaveOrganization(user.owner.UserCtx, orgToLeave.owner.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// the reassigned default org must be an org the user is actually a member of,
	// not an arbitrary org picked from the unscoped privacy-allowed query
	setting, err := suite.client.db.UserSetting.Query().
		Where(usersetting.UserID(userID)).
		WithDefaultOrg().
		Only(allowCtx)
	assert.NilError(t, err)
	assert.Assert(t, setting.Edges.DefaultOrg != nil)

	newDefaultOrgID := setting.Edges.DefaultOrg.ID
	assert.Check(t, newDefaultOrgID != orgToLeave.owner.OrganizationID)

	isMember, err := suite.client.db.OrgMembership.Query().
		Where(
			orgmembership.UserID(userID),
			orgmembership.OrganizationID(newDefaultOrgID),
		).
		Exist(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, isMember, "default org %s was reassigned to an org the user is not a member of", newDefaultOrgID)

	cleanupOrganizationDataWithContext(user.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(orgToLeave.owner.UserCtx, t)
}

func (suite *GraphTestSuite) countOrgScopedGroupMemberships(t *testing.T, userID, orgID string) int {
	t.Helper()

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	count, err := suite.client.db.GroupMembership.Query().
		Where(
			groupmembership.UserID(userID),
			groupmembership.HasGroupWith(
				group.OwnerID(orgID),
				group.DeletedAtIsNil(),
			),
		).
		Count(allowCtx)
	assert.NilError(t, err)

	return count
}

func (suite *GraphTestSuite) countOrgScopedProgramMemberships(t *testing.T, userID, orgID string) int {
	t.Helper()

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	count, err := suite.client.db.ProgramMembership.Query().
		Where(
			programmembership.UserID(userID),
			programmembership.HasProgramWith(program.OwnerID(orgID)),
		).
		Count(allowCtx)
	assert.NilError(t, err)

	return count
}

func (suite *GraphTestSuite) assertManagedGroupMembership(t *testing.T, userID, orgID, groupName string) {
	t.Helper()

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	exists, err := suite.client.db.GroupMembership.Query().
		Where(
			groupmembership.UserID(userID),
			groupmembership.HasGroupWith(
				group.OwnerID(orgID),
				group.IsManaged(true),
				group.Name(groupName),
			),
		).
		Exist(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, exists, "expected user %s to be in managed group %q of org %s", userID, groupName, orgID)
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
