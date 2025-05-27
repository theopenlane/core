package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryOrgMembers(t *testing.T) {
	t.Parallel()

	testOrgMemberUser := suite.userBuilder(context.Background(), t)
	org1Member := (&OrgMemberBuilder{client: suite.client}).MustNew(testOrgMemberUser.UserCtx, t)

	pm := (&ProgramMemberBuilder{client: suite.client}).MustNew(testOrgMemberUser.UserCtx, t)

	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: testOrgMemberUser.OrganizationID}).MustNew(testOrgMemberUser.UserCtx, t)

	childReqCtx := auth.NewTestContextWithOrgID(testOrgMemberUser.ID, childOrg.ID)

	(&OrgMemberBuilder{client: suite.client}).MustNew(childReqCtx, t)
	(&OrgMemberBuilder{client: suite.client, UserID: org1Member.UserID}).MustNew(childReqCtx, t)

	testCases := []struct {
		name                string
		queryID             string
		deleteProgramMember bool
		whereInput          *openlaneclient.OrgMembershipWhereInput
		client              *openlaneclient.OpenlaneClient
		ctx                 context.Context
		expectedLen         int
		expectErr           bool
	}{
		{
			name:        "happy path, get org members by org id",
			queryID:     testOrgMemberUser.OrganizationID,
			client:      suite.client.api,
			ctx:         testOrgMemberUser.UserCtx,
			expectedLen: 3,
		},
		{
			name:        "happy path, get org with parent members based on context",
			client:      suite.client.api,
			ctx:         childReqCtx,
			expectedLen: 4, // 2 from child org, 2 from parent org because we dedupe plus the program member
		},
		{
			name:    "where input, get members in program",
			queryID: testOrgMemberUser.OrganizationID,
			client:  suite.client.api,
			ctx:     testOrgMemberUser.UserCtx,
			whereInput: &openlaneclient.OrgMembershipWhereInput{
				HasUserWith: []*openlaneclient.UserWhereInput{
					{
						HasProgramMembershipsWith: []*openlaneclient.ProgramMembershipWhereInput{
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
			queryID: testOrgMemberUser.OrganizationID,
			client:  suite.client.api,
			ctx:     testOrgMemberUser.UserCtx,
			whereInput: &openlaneclient.OrgMembershipWhereInput{
				Not: &openlaneclient.OrgMembershipWhereInput{
					HasUserWith: []*openlaneclient.UserWhereInput{
						{
							HasProgramMembershipsWith: []*openlaneclient.ProgramMembershipWhereInput{
								{
									ProgramID: &pm.ProgramID,
								},
							},
						},
					},
				},
			},
			expectedLen: 1, // everyone not the owner and the program member
		},
		{
			name:                "where input, get members in program, after deleting a member",
			deleteProgramMember: true,
			queryID:             testOrgMemberUser.OrganizationID,
			client:              suite.client.api,
			ctx:                 testOrgMemberUser.UserCtx,
			whereInput: &openlaneclient.OrgMembershipWhereInput{
				HasUserWith: []*openlaneclient.UserWhereInput{
					{
						HasProgramMembershipsWith: []*openlaneclient.ProgramMembershipWhereInput{
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
			queryID:     testOrgMemberUser.OrganizationID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedLen: 0,
			expectErr:   false, // no org members returned
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.client.api,
			ctx:         testOrgMemberUser.UserCtx,
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
				tc.whereInput = &openlaneclient.OrgMembershipWhereInput{}
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
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, IDs: []string{childOrg.ID, testOrgMemberUser.OrganizationID}}).MustDelete(testOrgMemberUser.UserCtx, t)
}

func TestMutationCreateOrgMembers(t *testing.T) {
	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	userCtx := auth.NewTestContextWithOrgID(testUser1.ID, org1.ID)

	user1 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	user2 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	userWithValidDomain := (&UserBuilder{client: suite.client, Email: "matt@anderson.net"}).MustNew(testUser1.UserCtx, t)
	userWithInvalidDomain := (&UserBuilder{client: suite.client, Email: "mitb@example.com"}).MustNew(testUser1.UserCtx, t)

	orgWithRestrictions := (&OrganizationBuilder{client: suite.client, AllowedDomains: []string{"anderson.io", "anderson.net"}}).MustNew(testUser1.UserCtx, t)
	otherOrgCtx := auth.NewTestContextWithOrgID(testUser1.ID, orgWithRestrictions.ID)

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
			orgID:  org1.ID,
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
			name:   "happy path, add member in org with allowed domains",
			orgID:  org1.ID,
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
			orgID:  org1.ID,
			userID: user1.ID,
			role:   enums.RoleMember,
			ctx:    userCtx,
			errMsg: "already exists",
		},
		{
			name:   "cannot add self to organization",
			orgID:  org1.ID,
			userID: testUser2.ID,
			role:   enums.RoleMember,
			ctx:    testUser2.UserCtx,
			errMsg: notAuthorizedErrorMsg,
		},
		{
			name:   "add user to personal org not allowed",
			orgID:  testUser1.PersonalOrgID,
			userID: user1.ID,
			role:   enums.RoleMember,
			ctx:    userCtx,
			errMsg: hooks.ErrPersonalOrgsNoMembers.Error(),
		},
		{
			name:   "invalid user",
			orgID:  org1.ID,
			userID: ulids.New().String(),
			role:   enums.RoleMember,
			ctx:    userCtx,
			errMsg: "user not found",
		},
		{
			name:   "no access",
			orgID:  org1.ID,
			userID: user1.ID,
			role:   enums.RoleMember,
			ctx:    viewOnlyUser.UserCtx,
			errMsg: notAuthorizedErrorMsg,
		},
		{
			name:   "invalid role",
			orgID:  org1.ID,
			userID: user1.ID,
			role:   enums.RoleInvalid,
			ctx:    userCtx,
			errMsg: "not a valid OrgMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			input := openlaneclient.CreateOrgMembershipInput{
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
			suite.assertDefaultOrgUpdate(testUser1.UserCtx, t, tc.userID, tc.orgID, true)
		})
	}

	// delete created org and users
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, IDs: []string{org1.ID, orgWithRestrictions.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.UserDeleteOne]{client: suite.client.db.User, IDs: []string{user1.ID, user2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateOrgMembers(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	testUserOrg := suite.userBuilder(context.Background(), t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUserOrg.UserCtx, t)

	orgMembers, err := suite.client.api.GetOrgMembersByOrgID(testUserOrg.UserCtx, &openlaneclient.OrgMembershipWhereInput{
		OrganizationID: &testUserOrg.OrganizationID,
	})
	assert.NilError(t, err)

	testUserOrgMember := ""

	for _, edge := range orgMembers.OrgMemberships.Edges {
		if edge.Node.UserID == testUserOrg.ID {
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
			orgMemberID: om.ID,
			role:        enums.RoleInvalid,
			errMsg:      "not a valid OrgMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			input := openlaneclient.UpdateOrgMembershipInput{
				Role: &tc.role,
			}

			resp, err := suite.client.api.UpdateUserRoleInOrg(testUserOrg.UserCtx, tc.orgMemberID, input)

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
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: om.ID}).MustDelete(testUserOrg.UserCtx, t)
}

func TestMutationDeleteOrgMembers(t *testing.T) {
	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	resp, err := suite.client.api.RemoveUserFromOrg(testUser1.UserCtx, om.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(om.ID, resp.DeleteOrgMembership.DeletedID))

	// make sure the user default org is not set to the deleted org
	suite.assertDefaultOrgUpdate(testUser1.UserCtx, t, om.UserID, om.OrganizationID, false)

	// test re-adding the user to the org
	_, err = suite.client.api.AddUserToOrgWithRole(testUser1.UserCtx, openlaneclient.CreateOrgMembershipInput{
		OrganizationID: om.OrganizationID,
		UserID:         om.UserID,
		Role:           &om.Role,
	})

	assert.NilError(t, err)

	// cant remove self from org and owners cannot be removed
	orgMembers, err := suite.client.api.GetOrgMembersByOrgID(testUser1.UserCtx, &openlaneclient.OrgMembershipWhereInput{
		OrganizationID: &testUser1.OrganizationID,
	})
	assert.NilError(t, err)

	for _, edge := range orgMembers.OrgMemberships.Edges {
		// cannot delete self
		if edge.Node.UserID == adminUser.ID {
			_, err := suite.client.api.RemoveUserFromOrg(adminUser.UserCtx, edge.Node.ID)
			assert.ErrorContains(t, err, notAuthorizedErrorMsg)
		}

		// organization owner cannot be deleted
		if edge.Node.UserID == testUser1.ID {
			_, err = suite.client.api.RemoveUserFromOrg(adminUser.UserCtx, edge.Node.ID)
			assert.ErrorContains(t, err, notAuthorizedErrorMsg)
			break
		}
	}
}

func (suite *GraphTestSuite) assertDefaultOrgUpdate(ctx context.Context, t *testing.T, userID, orgID string, isEqual bool) {
	// when an org membership is deleted, the user default org should be updated
	// we need to allow the request because this is not for the user making the request
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	where := openlaneclient.UserSettingWhereInput{
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
