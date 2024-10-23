package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryOrgMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org1Member := (&OrgMemberBuilder{client: suite.client, OrgID: testOrgID}).MustNew(reqCtx, t)

	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: testOrgID}).MustNew(reqCtx, t)

	childReqCtx, err := auth.NewTestContextWithOrgID(testUser.ID, childOrg.ID)
	require.NoError(t, err)

	orgMember2 := (&OrgMemberBuilder{client: suite.client, OrgID: childOrg.ID}).MustNew(childReqCtx, t)
	orgMember3 := (&OrgMemberBuilder{client: suite.client, OrgID: childOrg.ID, UserID: org1Member.UserID}).MustNew(childReqCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		expectedLen int
		expectErr   bool
	}{
		{
			name:        "happy path, get org members by org id",
			queryID:     testOrgID,
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedLen: 2,
		},
		{
			name:        "happy path, get org with parent members based on context",
			client:      suite.client.api,
			ctx:         childReqCtx,
			allowed:     true,
			expectedLen: 3, // 2 from child org, 1 from parent org because we dedupe
		},
		{
			name:        "happy path, get org with parent members using org ID, only direct members will be returned",
			queryID:     childOrg.ID,
			client:      suite.client.api,
			ctx:         childReqCtx,
			allowed:     true,
			expectedLen: 2, // only child org members will be returned
		},
		{
			name:        "no access",
			queryID:     testOrgID,
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedLen: 0,
			expectErr:   true,
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedLen: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			orgID := tc.queryID
			whereInput := openlaneclient.OrgMembershipWhereInput{}

			if orgID != "" {
				whereInput.OrganizationID = &orgID

				// if thee user is providing an org id, we check if they have access to the org
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			if tc.expectedLen > 0 {
				// mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID})
				mock_fga.ListUsersAny(t, suite.client.fga, []string{org1Member.UserID,
					orgMember2.UserID,
					orgMember3.UserID,
					testUser.ID,
				}, nil)
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
	(&OrganizationCleanup{client: suite.client, ID: childOrg.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationCreateOrgMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	// allow access to organization
	checkCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	orgMember, err := org1.Members(checkCtx)
	require.NoError(t, err)
	require.Len(t, orgMember, 1)

	testUser1 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	testUser2 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name      string
		orgID     string
		userID    string
		role      enums.Role
		checkOrg  bool
		checkRole bool
		errMsg    string
	}{
		{
			name:      "happy path, add admin",
			orgID:     org1.ID,
			userID:    testUser1.ID,
			role:      enums.RoleAdmin,
			checkRole: true,
			checkOrg:  true,
		},
		{
			name:      "happy path, add member",
			orgID:     org1.ID,
			userID:    testUser2.ID,
			role:      enums.RoleMember,
			checkRole: true,
			checkOrg:  true,
		},
		{
			name:      "duplicate user, different role",
			orgID:     org1.ID,
			userID:    testUser1.ID,
			role:      enums.RoleMember,
			checkOrg:  true,
			checkRole: true,
			errMsg:    "already exists",
		},
		{
			name:      "add user to personal org not allowed",
			orgID:     testPersonalOrgID,
			userID:    testUser1.ID,
			role:      enums.RoleMember,
			checkOrg:  true,
			checkRole: true,
			errMsg:    hooks.ErrPersonalOrgsNoMembers.Error(),
		},
		{
			name:      "invalid user",
			orgID:     org1.ID,
			userID:    ulids.New().String(),
			role:      enums.RoleMember,
			checkOrg:  true,
			checkRole: true,
			errMsg:    "constraint failed, unable to complete the action",
		},
		{
			name:      "invalid org",
			orgID:     ulids.New().String(),
			userID:    testUser1.ID,
			role:      enums.RoleMember,
			checkOrg:  true,
			checkRole: true,
			errMsg:    "organization not found",
		},
		{
			name:      "invalid role",
			orgID:     org1.ID,
			userID:    testUser1.ID,
			role:      enums.RoleInvalid,
			checkOrg:  false,
			checkRole: false,
			errMsg:    "not a valid OrgMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errMsg == "" {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			// checks role in org to ensure user has ability to add other members
			if tc.checkRole {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			if tc.checkOrg {
				mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + org1.ID, "organization:" + testPersonalOrgID})
			}

			role := tc.role
			input := openlaneclient.CreateOrgMembershipInput{
				OrganizationID: tc.orgID,
				UserID:         tc.userID,
				Role:           &role,
			}

			resp, err := suite.client.api.AddUserToOrgWithRole(reqCtx, input)

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
			suite.assertDefaultOrgUpdate(reqCtx, tc.userID, tc.orgID, true)
		})
	}

	// delete created org and users
	(&OrganizationCleanup{client: suite.client, ID: org1.ID}).MustDelete(reqCtx, t)
	(&UserCleanup{client: suite.client, ID: testUser1.ID}).MustDelete(reqCtx, t)
	(&UserCleanup{client: suite.client, ID: testUser2.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationUpdateOrgMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, om.OrganizationID)
	require.NoError(t, err)

	testCases := []struct {
		name       string
		role       enums.Role
		tupleWrite bool
		errMsg     string
	}{
		{
			name:       "happy path, update to admin from member",
			tupleWrite: true,
			role:       enums.RoleAdmin,
		},
		{
			name:       "happy path, update to member from admin",
			tupleWrite: true,
			role:       enums.RoleMember,
		},
		{
			name:       "update to same role",
			tupleWrite: false, // nothing should change
			role:       enums.RoleMember,
		},
		{
			name:       "invalid role",
			role:       enums.RoleInvalid,
			tupleWrite: false,
			errMsg:     "not a valid OrgMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.tupleWrite {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			if tc.errMsg == "" {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			role := tc.role
			input := openlaneclient.UpdateOrgMembershipInput{
				Role: &role,
			}

			resp, err := suite.client.api.UpdateUserRoleInOrg(reqCtx, om.ID, input)

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
	(&OrgMemberCleanup{client: suite.client, ID: om.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationDeleteOrgMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(reqCtx, t)

	mock_fga.WriteAny(t, suite.client.fga)
	mock_fga.CheckAny(t, suite.client.fga, true)

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, om.OrganizationID)
	require.NoError(t, err)

	resp, err := suite.client.api.RemoveUserFromOrg(reqCtx, om.ID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.DeleteOrgMembership)
	assert.Equal(t, om.ID, resp.DeleteOrgMembership.DeletedID)

	// make sure the user default org is not set to the deleted org
	suite.assertDefaultOrgUpdate(reqCtx, om.UserID, om.OrganizationID, false)
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
