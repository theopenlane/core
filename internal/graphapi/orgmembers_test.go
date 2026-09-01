package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryOrgMembers(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedFreshOrgUsers(t)
	org1Member := localTestOrg.Member

	pm := (&th.ProgramMemberBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)

	childOrg := (&th.OrganizationBuilder{Client: suite.Client, ParentOrgID: localTestOrg.Owner.OrganizationID}).MustNew(localTestOrg.Owner.UserCtx, t)

	childReqCtx := auth.NewTestContextWithOrgID(localTestOrg.Owner.ID, childOrg.ID)

	(&th.OrgMemberBuilder{Client: suite.Client}).MustNew(childReqCtx, t)
	(&th.OrgMemberBuilder{Client: suite.Client, UserID: org1Member.ID}).MustNew(childReqCtx, t)

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
			queryID:     localTestOrg.Owner.OrganizationID,
			client:      suite.Client.API,
			ctx:         localTestOrg.Owner.UserCtx,
			expectedLen: 6,
		},
		{
			name:        "happy path, get org members by org id, member",
			queryID:     localTestOrg.Owner.OrganizationID,
			client:      suite.Client.API,
			ctx:         localTestOrg.Member.UserCtx,
			expectedLen: 6,
		},
		{
			name:        "happy path, get org members by org id, auditor",
			queryID:     localTestOrg.Owner.OrganizationID,
			client:      suite.Client.API,
			ctx:         localTestOrg.Auditor.UserCtx,
			expectedLen: 6,
		},
		{
			name:        "happy path, get org with parent members based on context",
			client:      suite.Client.API,
			ctx:         childReqCtx,
			expectedLen: 7, // 2 from child org, 5 from parent org because we dedupe plus the program member
		},
		{
			name:    "where input, get members in program",
			queryID: localTestOrg.Owner.OrganizationID,
			client:  suite.Client.API,
			ctx:     localTestOrg.Owner.UserCtx,
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
			queryID: localTestOrg.Owner.OrganizationID,
			client:  suite.Client.API,
			ctx:     localTestOrg.Owner.UserCtx,
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
			queryID:             localTestOrg.Owner.OrganizationID,
			client:              suite.Client.API,
			ctx:                 localTestOrg.Owner.UserCtx,
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
			client:      suite.Client.API,
			ctx:         childReqCtx,
			expectedLen: 2, // only child org members will be returned
		},
		{
			name:        "no access",
			queryID:     localTestOrg.Owner.OrganizationID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedLen: 0,
			expectErr:   false, // no org members returned
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.Client.API,
			ctx:         localTestOrg.Owner.UserCtx,
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
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestQueryOrgMembersWithAdditionalRoles(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedFreshOrgUsers(t)
	org1Member := localTestOrg.Member

	// add policy manager and trust center manager role
	suite.AddFunctionalRoleForUser(localTestOrg.Owner.UserCtx, t, org1Member.ID, localTestOrg.Owner.OrganizationID, []string{"policy_manager", "trust_center_manager"})
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
			client:                suite.Client.API,
			ctx:                   localTestOrg.Owner.UserCtx,
			expectAdditionalRoles: true,
		},
		{
			name: "happy path, get org auditor has no additional roles",
			whereInput: &testclient.OrgMembershipWhereInput{
				UserID: &localTestOrg.Auditor.ID,
			},
			client:                suite.Client.API,
			ctx:                   localTestOrg.Owner.UserCtx,
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
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestMutationCreateOrgMembers(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.UserBuilder(context.Background(), t)
	org1ID := localTestOrg.OrganizationID

	userCtx := localTestOrg.UserCtx
	personalOrgCtx := auth.NewTestContextWithOrgID(localTestOrg.ID, localTestOrg.PersonalOrgID)

	user1 := (&th.UserBuilder{Client: suite.Client}).MustNew(userCtx, t)
	user2 := (&th.UserBuilder{Client: suite.Client}).MustNew(userCtx, t)
	user3 := (&th.UserBuilder{Client: suite.Client, Email: "mitb2@anderson.io", FirstName: "FirstName!@"}).MustNew(userCtx, t)

	userWithValidDomain := (&th.UserBuilder{Client: suite.Client, Email: "matt@anderson.net"}).MustNew(userCtx, t)
	userWithAnotherDomain := (&th.UserBuilder{Client: suite.Client, Email: "mitb@example.com"}).MustNew(userCtx, t)

	orgWithRestrictions := (&th.OrganizationBuilder{Client: suite.Client, AllowedDomains: []string{"anderson.io", "anderson.net"}}).MustNew(localTestOrg.UserCtx, t)
	otherOrgCtx := auth.NewTestContextWithOrgID(localTestOrg.ID, orgWithRestrictions.ID)

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
			userID: th.SharedTestUser2.ID,
			role:   enums.RoleMember,
			ctx:    th.SharedTestUser2.UserCtx,
			errMsg: th.NotFoundErrorMsg, // organization is not found because user does not have access to it
		},
		{
			name:   "add user to personal org not allowed",
			orgID:  localTestOrg.PersonalOrgID,
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
			ctx:    th.SharedViewOnlyUser.UserCtx,
			errMsg: th.NotAuthorizedErrorMsg,
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

			resp, err := suite.Client.API.AddUserToOrgWithRole(tc.ctx, input)

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
			suite.assertDefaultOrgUpdate(th.SharedTestUser1.UserCtx, t, tc.userID, tc.orgID, true)
		})
	}

	// delete created org and users
	th.CleanupOrganizationDataWithContext(otherOrgCtx, t)
	th.CleanupOrganizationDataWithContext(localTestOrg.UserCtx, t)
}

func TestMutationUpdateOrgMembers(t *testing.T) {
	// create another user for this test
	// so it doesn't interfere with the other tests
	t.Parallel()

	localTestOrg := suite.SeedOrgOwner(t)

	om := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)
	orgMembers, err := suite.Client.API.GetOrgMembersByOrgID(localTestOrg.Owner.UserCtx, &testclient.OrgMembershipWhereInput{
		OrganizationID: &localTestOrg.Owner.OrganizationID,
	})
	assert.NilError(t, err)

	testUserOrgMember := ""

	for _, edge := range orgMembers.OrgMemberships.Edges {
		if edge.Node.UserID == localTestOrg.Owner.ID {
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

			resp, err := suite.Client.API.UpdateUserRoleInOrg(localTestOrg.Owner.UserCtx, tc.orgMemberID, input)

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
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestMutationUpdateOrgMemberRole(t *testing.T) {
	t.Parallel()

	org := suite.SeedFreshOrgUsers(t)
	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	user := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(org.Owner.UserCtx, t, &user, enums.RoleMember, org.Owner.OrganizationID)

	roleUpdateMember, err := suite.Client.DB.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.Owner.OrganizationID),
			orgmembership.UserID(user.ID),
		).
		Only(allowCtx)
	assert.NilError(t, err)

	ownerMember, err := suite.Client.DB.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.Owner.OrganizationID),
			orgmembership.UserID(org.Owner.ID),
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
			ctx:         org.Admin.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleAdmin,
		},
		{
			name:        "admin cannot update member to super admin",
			ctx:         org.Admin.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleSuperAdmin,
			errMsg:      th.NotAuthorizedErrorMsg,
		},
		{
			name:        "member cannot update admin to member",
			ctx:         org.Member.UserCtx,
			orgMemberID: roleUpdateMember.ID,
			role:        enums.RoleMember,
			errMsg:      th.NotAuthorizedErrorMsg,
		},
		{
			name:        "owner role cannot be changed directly",
			ctx:         org.Admin.UserCtx,
			orgMemberID: ownerMember.ID,
			role:        enums.RoleAdmin,
			errMsg:      hooks.ErrOrgOwnerCannotBeUpdated.Error(),
		},
		{
			name:        "owner role cannot be assigned directly",
			ctx:         org.Owner.UserCtx,
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

			resp, err := suite.Client.API.UpdateUserRoleInOrg(tc.ctx, tc.orgMemberID, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.role, resp.UpdateOrgMembership.OrgMembership.Role))
		})
	}

	th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
}

func TestMutationBulkUpdateOrgMemberRole(t *testing.T) {
	t.Parallel()

	org := suite.SeedFreshOrgUsers(t)
	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	user1 := suite.UserBuilder(context.Background(), t)
	user2 := suite.UserBuilder(context.Background(), t)

	suite.AddUserToOrganization(org.Owner.UserCtx, t, &user1, enums.RoleMember, org.Owner.OrganizationID)
	suite.AddUserToOrganization(org.Owner.UserCtx, t, &user2, enums.RoleMember, org.Owner.OrganizationID)

	currentMembers, err := suite.Client.DB.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.Owner.OrganizationID),
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

	resp, err := suite.Client.API.UpdateBulkOrgMemberRoles(org.Admin.UserCtx, ids, input)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Len(resp.UpdateBulkOrgMembership.UpdatedIDs, len(ids)))
	assert.Check(t, is.Len(resp.UpdateBulkOrgMembership.OrgMemberships, len(ids)))

	updatedMembers, err := suite.Client.DB.OrgMembership.Query().
		Where(orgmembership.IDIn(ids...)).
		All(allowCtx)
	assert.NilError(t, err)

	for _, member := range updatedMembers {
		assert.Check(t, is.Equal(enums.RoleAdmin, member.Role))
	}

	ownerMember, err := suite.Client.DB.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(org.Owner.OrganizationID),
			orgmembership.UserID(org.Owner.ID),
		).
		Only(allowCtx)
	assert.NilError(t, err)

	memberRole := enums.RoleMember
	input.Role = &memberRole

	ownerResp, err := suite.Client.API.UpdateBulkOrgMemberRoles(org.Admin.UserCtx, []string{ownerMember.ID}, input)
	assert.NilError(t, err)
	assert.Assert(t, ownerResp != nil)
	assert.Check(t, is.Len(ownerResp.UpdateBulkOrgMembership.UpdatedIDs, 0))
	assert.Check(t, is.Len(ownerResp.UpdateBulkOrgMembership.OrgMemberships, 0))

	ownerMember, err = suite.Client.DB.OrgMembership.Query().
		Where(orgmembership.ID(ownerMember.ID)).
		Only(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.RoleOwner, ownerMember.Role))

	th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
}

func TestMutationDeleteOrgMembers(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedOrgOwner(t)

	om := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)
	adminOrgMember := (&th.OrgMemberBuilder{Client: suite.Client, Role: string(enums.RoleAdmin)}).MustNew(localTestOrg.Owner.UserCtx, t)

	// create admin user context
	adminUserCtx := auth.NewTestContextWithOrgID(adminOrgMember.UserID, localTestOrg.Owner.OrganizationID)

	resp, err := suite.Client.API.RemoveUserFromOrg(localTestOrg.Owner.UserCtx, om.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(om.ID, resp.DeleteOrgMembership.DeletedID))

	// make sure the user default org is not set to the deleted org
	suite.assertDefaultOrgUpdate(localTestOrg.Owner.UserCtx, t, om.UserID, om.OrganizationID, false)

	// re-adding the user to the org should succeed since the org membership
	// is deleted and the managed group is properly cleaned up
	reAddResp, err := suite.Client.API.AddUserToOrgWithRole(localTestOrg.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: localTestOrg.Owner.OrganizationID,
		UserID:         om.UserID,
		Role:           &enums.RoleAdmin,
	})

	assert.NilError(t, err)
	assert.Assert(t, reAddResp != nil)

	// cant remove self from org and owners cannot be removed
	orgMembers, err := suite.Client.API.GetOrgMembersByOrgID(localTestOrg.Owner.UserCtx, &testclient.OrgMembershipWhereInput{
		OrganizationID: &localTestOrg.Owner.OrganizationID,
	})
	assert.NilError(t, err)

	for _, edge := range orgMembers.OrgMemberships.Edges {
		// cannot delete self
		if edge.Node.UserID == th.SharedAdminUser.ID {
			_, err := suite.Client.API.RemoveUserFromOrg(adminUserCtx, edge.Node.ID)
			assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)
		}

		// organization owner cannot be deleted
		if edge.Node.UserID == localTestOrg.Owner.ID {
			_, err = suite.Client.API.RemoveUserFromOrg(adminUserCtx, edge.Node.ID)
			assert.ErrorContains(t, err, "organization owner cannot be deleted")
			break
		}
	}

	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestMutationLeaveOrganization(t *testing.T) {
	t.Parallel()

	currentOrg := suite.SeedOrgOwner(t)
	orgToLeave := suite.SeedOrgOwner(t)

	memberRole := enums.RoleMember
	member, err := suite.Client.API.AddUserToOrgWithRole(orgToLeave.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToLeave.Owner.OrganizationID,
		UserID:         currentOrg.Owner.ID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	resp, err := suite.Client.API.LeaveOrganization(currentOrg.Owner.UserCtx, orgToLeave.Owner.OrganizationID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(member.CreateOrgMembership.OrgMembership.ID, resp.LeaveOrganization.DeletedID))

	members, err := suite.Client.API.GetOrgMembersByOrgID(orgToLeave.Owner.UserCtx, &testclient.OrgMembershipWhereInput{
		OrganizationID: &orgToLeave.Owner.OrganizationID,
		UserID:         &currentOrg.Owner.ID,
	})
	assert.NilError(t, err)
	assert.Assert(t, members != nil)
	assert.Check(t, is.Len(members.OrgMemberships.Edges, 0))

	suite.assertDefaultOrgUpdate(currentOrg.Owner.UserCtx, t, currentOrg.Owner.ID, orgToLeave.Owner.OrganizationID, false)

	th.CleanupOrganizationDataWithContext(currentOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(orgToLeave.Owner.UserCtx, t)
}

func TestMutationLeaveOrganizationPreservesOtherOrgMemberships(t *testing.T) {
	t.Parallel()

	currentOrg := suite.SeedOrgOwner(t)
	orgToLeave := suite.SeedOrgOwner(t)

	userID := currentOrg.Owner.ID

	memberRole := enums.RoleMember
	member, err := suite.Client.API.AddUserToOrgWithRole(orgToLeave.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToLeave.Owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	// creating a program adds the creator as a program admin, giving the user a
	// program membership in the org they are staying in
	(&th.ProgramBuilder{Client: suite.Client}).MustNew(currentOrg.Owner.UserCtx, t)

	// the user should be a member of the managed groups in both orgs
	groupsBefore := suite.countOrgScopedGroupMemberships(t, userID, currentOrg.Owner.OrganizationID)
	assert.Check(t, groupsBefore > 0)

	programsBefore := suite.countOrgScopedProgramMemberships(t, userID, currentOrg.Owner.OrganizationID)
	assert.Check(t, programsBefore > 0)

	assert.Check(t, suite.countOrgScopedGroupMemberships(t, userID, orgToLeave.Owner.OrganizationID) > 0)

	resp, err := suite.Client.API.LeaveOrganization(currentOrg.Owner.UserCtx, orgToLeave.Owner.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// memberships in the org the user left should be removed
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, userID, orgToLeave.Owner.OrganizationID)))
	assert.Check(t, is.Equal(0, suite.countOrgScopedProgramMemberships(t, userID, orgToLeave.Owner.OrganizationID)))

	// memberships in every other org must be left intact
	assert.Check(t, is.Equal(groupsBefore, suite.countOrgScopedGroupMemberships(t, userID, currentOrg.Owner.OrganizationID)))
	assert.Check(t, is.Equal(programsBefore, suite.countOrgScopedProgramMemberships(t, userID, currentOrg.Owner.OrganizationID)))

	// the user must specifically still be in the system managed groups of the org they stayed in
	suite.assertManagedGroupMembership(t, userID, currentOrg.Owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, userID, currentOrg.Owner.OrganizationID, hooks.AllMembersGroup)

	th.CleanupOrganizationDataWithContext(currentOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(orgToLeave.Owner.UserCtx, t)
}

func TestMutationDeleteOrgMembersPreservesOtherOrgMemberships(t *testing.T) {
	t.Parallel()

	homeOrg := suite.SeedOrgOwner(t)
	otherOrg := suite.SeedOrgOwner(t)

	userID := homeOrg.Owner.ID

	memberRole := enums.RoleMember
	member, err := suite.Client.API.AddUserToOrgWithRole(otherOrg.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: otherOrg.Owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	groupsBefore := suite.countOrgScopedGroupMemberships(t, userID, homeOrg.Owner.OrganizationID)
	assert.Assert(t, groupsBefore > 0)

	// the other org's owner removes the user from their org
	resp, err := suite.Client.API.RemoveUserFromOrg(otherOrg.Owner.UserCtx, member.CreateOrgMembership.OrgMembership.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// memberships in the org the user was removed from should be gone
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, userID, otherOrg.Owner.OrganizationID)))

	// memberships in the user's other org must be left intact
	assert.Check(t, is.Equal(groupsBefore, suite.countOrgScopedGroupMemberships(t, userID, homeOrg.Owner.OrganizationID)))
	suite.assertManagedGroupMembership(t, userID, homeOrg.Owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, userID, homeOrg.Owner.OrganizationID, hooks.AllMembersGroup)

	th.CleanupOrganizationDataWithContext(homeOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(otherOrg.Owner.UserCtx, t)
}

func TestMutationDeleteOrganizationPreservesOtherOrgMemberships(t *testing.T) {
	t.Parallel()

	homeOrg := suite.SeedOrgOwner(t)
	orgToDelete := suite.SeedOrgOwner(t)

	userID := homeOrg.Owner.ID

	memberRole := enums.RoleMember
	member, err := suite.Client.API.AddUserToOrgWithRole(orgToDelete.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToDelete.Owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	groupsBefore := suite.countOrgScopedGroupMemberships(t, userID, homeOrg.Owner.OrganizationID)
	assert.Assert(t, groupsBefore > 0)
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, userID, orgToDelete.Owner.OrganizationID) > 0)

	// delete the entire organization the user is a member of
	resp, err := suite.Client.API.DeleteOrganization(orgToDelete.Owner.UserCtx, orgToDelete.Owner.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// the deleted org's own groups and memberships are removed asynchronously by the
	// organization cascade delete listener, so only the cross-org invariant is asserted here:
	// memberships in the user's other org must be left intact
	assert.Check(t, is.Equal(groupsBefore, suite.countOrgScopedGroupMemberships(t, userID, homeOrg.Owner.OrganizationID)))
	suite.assertManagedGroupMembership(t, userID, homeOrg.Owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, userID, homeOrg.Owner.OrganizationID, hooks.AllMembersGroup)

	th.CleanupOrganizationDataWithContext(homeOrg.Owner.UserCtx, t)
}

func TestMutationDeleteBulkOrgMembers(t *testing.T) {
	t.Parallel()

	homeOrg := suite.SeedOrgOwner(t)
	bulkOrg := suite.SeedOrgOwner(t)

	crossOrgUserID := homeOrg.Owner.ID

	// user with a membership in another org, to verify bulk removal stays org-scoped
	memberRole := enums.RoleMember
	crossOrgMember, err := suite.Client.API.AddUserToOrgWithRole(bulkOrg.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: bulkOrg.Owner.OrganizationID,
		UserID:         crossOrgUserID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)

	otherMember := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(bulkOrg.Owner.UserCtx, t)

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	ownerMembershipID, err := suite.Client.DB.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(bulkOrg.Owner.OrganizationID),
			orgmembership.UserID(bulkOrg.Owner.ID),
		).
		OnlyID(allowCtx)
	assert.NilError(t, err)

	homeGroupsBefore := suite.countOrgScopedGroupMemberships(t, crossOrgUserID, homeOrg.Owner.OrganizationID)
	assert.Assert(t, homeGroupsBefore > 0)
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, crossOrgUserID, bulkOrg.Owner.OrganizationID) > 0)
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, otherMember.UserID, bulkOrg.Owner.OrganizationID) > 0)

	crossOrgMembershipID := crossOrgMember.CreateOrgMembership.OrgMembership.ID

	resp, err := suite.Client.API.RemoveBulkUsersFromOrg(bulkOrg.Owner.UserCtx, []string{
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
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, crossOrgUserID, bulkOrg.Owner.OrganizationID)))
	assert.Check(t, is.Equal(0, suite.countOrgScopedGroupMemberships(t, otherMember.UserID, bulkOrg.Owner.OrganizationID)))

	// the owner keeps their memberships in the bulk org
	assert.Check(t, suite.countOrgScopedGroupMemberships(t, bulkOrg.Owner.ID, bulkOrg.Owner.OrganizationID) > 0)

	// memberships in the removed user's other org must be left intact
	assert.Check(t, is.Equal(homeGroupsBefore, suite.countOrgScopedGroupMemberships(t, crossOrgUserID, homeOrg.Owner.OrganizationID)))
	suite.assertManagedGroupMembership(t, crossOrgUserID, homeOrg.Owner.OrganizationID, hooks.AdminsGroup)
	suite.assertManagedGroupMembership(t, crossOrgUserID, homeOrg.Owner.OrganizationID, hooks.AllMembersGroup)

	th.CleanupOrganizationDataWithContext(homeOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(bulkOrg.Owner.UserCtx, t)
}

func TestMutationLeaveOrganizationReassignsDefaultOrgToMemberOrg(t *testing.T) {
	t.Parallel()

	user := suite.SeedOrgOwner(t)
	orgToLeave := suite.SeedOrgOwner(t)

	userID := user.Owner.ID

	memberRole := enums.RoleMember
	member, err := suite.Client.API.AddUserToOrgWithRole(orgToLeave.Owner.UserCtx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgToLeave.Owner.OrganizationID,
		UserID:         userID,
		Role:           &memberRole,
	})
	assert.NilError(t, err)
	assert.Assert(t, member != nil)

	// make the org they are about to leave their default org
	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	_, err = suite.Client.DB.UserSetting.Update().
		Where(usersetting.UserID(userID)).
		SetDefaultOrgID(orgToLeave.Owner.OrganizationID).
		Save(allowCtx)
	assert.NilError(t, err)

	resp, err := suite.Client.API.LeaveOrganization(user.Owner.UserCtx, orgToLeave.Owner.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// the reassigned default org must be an org the user is actually a member of,
	// not an arbitrary org picked from the unscoped privacy-allowed query
	setting, err := suite.Client.DB.UserSetting.Query().
		Where(usersetting.UserID(userID)).
		WithDefaultOrg().
		Only(allowCtx)
	assert.NilError(t, err)
	assert.Assert(t, setting.Edges.DefaultOrg != nil)

	newDefaultOrgID := setting.Edges.DefaultOrg.ID
	assert.Check(t, newDefaultOrgID != orgToLeave.Owner.OrganizationID)

	isMember, err := suite.Client.DB.OrgMembership.Query().
		Where(
			orgmembership.UserID(userID),
			orgmembership.OrganizationID(newDefaultOrgID),
		).
		Exist(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, isMember, "default org %s was reassigned to an org the user is not a member of", newDefaultOrgID)

	th.CleanupOrganizationDataWithContext(user.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(orgToLeave.Owner.UserCtx, t)
}

func (suite *graphTestSuite) countOrgScopedGroupMemberships(t *testing.T, userID, orgID string) int {
	t.Helper()

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	count, err := suite.Client.DB.GroupMembership.Query().
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

func (suite *graphTestSuite) countOrgScopedProgramMemberships(t *testing.T, userID, orgID string) int {
	t.Helper()

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	count, err := suite.Client.DB.ProgramMembership.Query().
		Where(
			programmembership.UserID(userID),
			programmembership.HasProgramWith(program.OwnerID(orgID)),
		).
		Count(allowCtx)
	assert.NilError(t, err)

	return count
}

func (suite *graphTestSuite) assertManagedGroupMembership(t *testing.T, userID, orgID, groupName string) {
	t.Helper()

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

	exists, err := suite.Client.DB.GroupMembership.Query().
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

func (suite *graphTestSuite) assertDefaultOrgUpdate(ctx context.Context, t *testing.T, userID, orgID string, isEqual bool) {
	// when an org membership is deleted, the user default org should be updated
	// we need to allow the request because this is not for the user making the request
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	where := testclient.UserSettingWhereInput{
		UserID: &userID,
	}

	userSettingResp, err := suite.Client.API.GetUserSettings(allowCtx, where)
	assert.NilError(t, err)
	assert.Assert(t, userSettingResp != nil)
	assert.Check(t, is.Len(userSettingResp.UserSettings.Edges, 1))

	if isEqual {
		assert.Check(t, is.Equal(orgID, userSettingResp.UserSettings.Edges[0].Node.DefaultOrg.ID))
	} else {
		assert.Check(t, orgID != userSettingResp.UserSettings.Edges[0].Node.DefaultOrg.ID)
	}
}
