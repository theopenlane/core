package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryOrgMembers() {
	t := suite.T()

	org1Member := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	childReqCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, childOrg.ID)
	require.NoError(t, err)

	(&OrgMemberBuilder{client: suite.client, OrgID: childOrg.ID}).MustNew(childReqCtx, t)
	(&OrgMemberBuilder{client: suite.client, OrgID: childOrg.ID, UserID: org1Member.UserID}).MustNew(childReqCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedLen int
		expectErr   bool
	}{
		{
			name:        "happy path, get org members by org id",
			queryID:     testUser1.OrganizationID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedLen: 4, // account for the seeded org members
		},
		{
			name:        "happy path, get org with parent members based on context",
			client:      suite.client.api,
			ctx:         childReqCtx,
			expectedLen: 5, // 2 from child org, 2 from parent org because we dedupe
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
			queryID:     testUser1.OrganizationID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedLen: 0,
			expectErr:   true,
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedLen: 0,
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			orgID := tc.queryID
			whereInput := openlaneclient.OrgMembershipWhereInput{}

			if orgID != "" {
				whereInput.OrganizationID = &orgID
			}

			resp, err := tc.client.GetOrgMembersByOrgID(tc.ctx, &whereInput)

			if tc.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			if tc.expectedLen == 0 {
				assert.Empty(t, resp.OrgMemberships.Edges)

				return
			}

			require.NotNil(t, resp)
			require.NotNil(t, resp.OrgMemberships)
			assert.Len(t, resp.OrgMemberships.Edges, tc.expectedLen)
		})
	}

	// delete created org
	(&OrganizationCleanup{client: suite.client, ID: childOrg.ID}).MustDelete(testUser1.UserCtx, t)
}

func (suite *GraphTestSuite) TestMutationCreateOrgMembers() {
	t := suite.T()

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// allow access to organization
	checkCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	orgMember, err := org1.Members(checkCtx)
	require.NoError(t, err)
	require.Len(t, orgMember, 1)

	userCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, org1.ID)
	require.NoError(t, err)

	user1 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	user2 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	userWithValidDomain := (&UserBuilder{client: suite.client, Email: "matt@anderson.net"}).MustNew(testUser1.UserCtx, t)
	userWithInvalidDomain := (&UserBuilder{client: suite.client, Email: "mitb@example.com"}).MustNew(testUser1.UserCtx, t)

	orgWithRestrictions := (&OrganizationBuilder{client: suite.client, AllowedDomains: []string{"anderson.io", "anderson.net"}}).MustNew(testUser1.UserCtx, t)
	otherOrgCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, orgWithRestrictions.ID)
	require.NoError(t, err)

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateOrgMembership)
			assert.Equal(t, tc.userID, resp.CreateOrgMembership.OrgMembership.UserID)
			assert.Equal(t, tc.orgID, resp.CreateOrgMembership.OrgMembership.OrganizationID)
			assert.Equal(t, tc.role, resp.CreateOrgMembership.OrgMembership.Role)

			// make sure the user default org is set to the new org
			suite.assertDefaultOrgUpdate(testUser1.UserCtx, tc.userID, tc.orgID, true)
		})
	}

	// delete created org and users
	(&OrganizationCleanup{client: suite.client, ID: org1.ID}).MustDelete(testUser1.UserCtx, t)
	(&UserCleanup{client: suite.client, ID: testUser1.ID}).MustDelete(testUser1.UserCtx, t)
	(&UserCleanup{client: suite.client, ID: testUser2.ID}).MustDelete(testUser1.UserCtx, t)
}

func (suite *GraphTestSuite) TestMutationUpdateOrgMembers() {
	t := suite.T()

	om := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	orgMembers, err := suite.client.api.GetOrgMembersByOrgID(testUser1.UserCtx, &openlaneclient.OrgMembershipWhereInput{
		OrganizationID: &testUser1.OrganizationID,
	})
	require.NoError(t, err)

	testUser1OrgMember := ""

	for _, edge := range orgMembers.OrgMemberships.Edges {
		if edge.Node.UserID == testUser1.ID {
			testUser1OrgMember = edge.Node.ID
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
			orgMemberID: testUser1OrgMember,
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

			resp, err := suite.client.api.UpdateUserRoleInOrg(testUser1.UserCtx, tc.orgMemberID, input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateOrgMembership)
			assert.Equal(t, tc.role, resp.UpdateOrgMembership.OrgMembership.Role)
		})
	}

	// delete created org and users
	(&OrgMemberCleanup{client: suite.client, ID: om.ID}).MustDelete(testUser1.UserCtx, t)
}

func (suite *GraphTestSuite) TestMutationDeleteOrgMembers() {
	t := suite.T()

	om := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	resp, err := suite.client.api.RemoveUserFromOrg(testUser1.UserCtx, om.ID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.DeleteOrgMembership)
	assert.Equal(t, om.ID, resp.DeleteOrgMembership.DeletedID)

	// make sure the user default org is not set to the deleted org
	suite.assertDefaultOrgUpdate(testUser1.UserCtx, om.UserID, om.OrganizationID, false)

	// test re-adding the user to the org
	_, err = suite.client.api.AddUserToOrgWithRole(testUser1.UserCtx, openlaneclient.CreateOrgMembershipInput{
		OrganizationID: om.OrganizationID,
		UserID:         om.UserID,
		Role:           &om.Role,
	})

	require.NoError(t, err)

	// cant remove self from org and owners cannot be removed
	orgMembers, err := suite.client.api.GetOrgMembersByOrgID(testUser1.UserCtx, &openlaneclient.OrgMembershipWhereInput{
		OrganizationID: &testUser1.OrganizationID,
	})
	require.NoError(t, err)

	for _, edge := range orgMembers.OrgMemberships.Edges {
		// cannot delete self
		if edge.Node.UserID == adminUser.ID {
			_, err := suite.client.api.RemoveUserFromOrg(adminUser.UserCtx, edge.Node.ID)
			require.Error(t, err)

		}

		// organization owner cannot be deleted
		if edge.Node.UserID == testUser1.ID {
			_, err = suite.client.api.RemoveUserFromOrg(adminUser.UserCtx, edge.Node.ID)
			require.Error(t, err)
			break
		}
	}
}

func (suite *GraphTestSuite) assertDefaultOrgUpdate(ctx context.Context, userID, orgID string, isEqual bool) {
	t := suite.T()

	// when an org membership is deleted, the user default org should be updated
	// we need to allow the request because this is not for the user making the request
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	where := openlaneclient.UserSettingWhereInput{
		UserID: &userID,
	}

	userSettingResp, err := suite.client.api.GetUserSettings(allowCtx, where)
	require.NoError(t, err)
	require.NotNil(t, userSettingResp)
	require.Len(t, userSettingResp.UserSettings.Edges, 1)

	if isEqual {
		assert.Equal(t, orgID, userSettingResp.UserSettings.Edges[0].Node.DefaultOrg.ID)
	} else {
		assert.NotEqual(t, orgID, userSettingResp.UserSettings.Edges[0].Node.DefaultOrg.ID)
	}
}
