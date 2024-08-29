package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryGroupMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	group := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	// allow access to group
	checkCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	groupMember, err := group.Members(checkCtx)
	require.NoError(t, err)
	require.Len(t, groupMember, 1)

	testCases := []struct {
		name        string
		queryID     string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		expected    *ent.GroupMembership
		errExpected bool
	}{
		{
			name:     "happy path, get group member by group id",
			queryID:  group.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			allowed:  true,
			expected: groupMember[0],
		},
		{
			name:     "happy path, get group member by group id using api token",
			queryID:  group.ID,
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			allowed:  true,
			expected: groupMember[0],
		},
		{
			name:     "happy path, get group member by group id using personal access token",
			queryID:  group.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			allowed:  true,
			expected: groupMember[0],
		},
		{
			name:        "get group member by group id, no access",
			queryID:     group.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expected:    nil,
			errExpected: true,
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			client:   suite.client.api,
			ctx:      reqCtx,
			allowed:  true,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			groupID := tc.queryID
			whereInput := openlaneclient.GroupMembershipWhereInput{
				GroupID: &groupID,
			}

			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			if tc.expected != nil {
				// list groups in order to determine access to group level data
				mock_fga.ListAny(t, suite.client.fga, []string{fmt.Sprintf("group:%s", group.ID)})
			}

			resp, err := tc.client.GetGroupMembersByGroupID(tc.ctx, &whereInput)

			if tc.errExpected {
				require.Error(t, err)
				assert.ErrorContains(t, err, "deny rule")

				return
			}

			require.NoError(t, err)

			if tc.expected == nil {
				assert.Empty(t, resp.GroupMemberships.Edges)

				return
			}

			require.NotNil(t, resp)
			require.NotNil(t, resp.GroupMemberships)
			assert.Equal(t, tc.expected.UserID, resp.GroupMemberships.Edges[0].Node.GetUser().GetID())
			assert.Equal(t, tc.expected.Role, resp.GroupMemberships.Edges[0].Node.Role)
		})
	}

	// delete created group
	(&GroupCleanup{client: suite.client, ID: group.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationCreateGroupMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	group1 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	// allow access to group
	checkCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	groupMember, err := group1.Members(checkCtx)
	require.NoError(t, err)
	require.Len(t, groupMember, 1)

	orgMember1 := (&OrgMemberBuilder{client: suite.client, OrgID: testOrgID}).MustNew(reqCtx, t)
	orgMember2 := (&OrgMemberBuilder{client: suite.client, OrgID: testOrgID}).MustNew(reqCtx, t)
	orgMember3 := (&OrgMemberBuilder{client: suite.client, OrgID: testOrgID}).MustNew(reqCtx, t)

	testCases := []struct {
		name    string
		groupID string
		userID  string
		role    enums.Role
		client  *openlaneclient.OpenlaneClient
		ctx     context.Context
		allowed bool
		check   bool
		list    bool
		errMsg  string
	}{
		{
			name:    "happy path, add admin",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleAdmin,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   true,
			list:    true,
		},
		{
			name:    "happy path, add member using api token",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleMember,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
			check:   true,
			list:    true,
		},
		{
			name:    "happy path, add member using personal access token",
			groupID: group1.ID,
			userID:  orgMember3.UserID,
			role:    enums.RoleMember,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
			check:   true,
			list:    true,
		},
		{
			name:    "add member, no access",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: false,
			check:   true,
			list:    false,
			errMsg:  "you are not authorized to perform this action",
		},
		{
			name:    "owner relation not valid for groups",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleOwner,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   false,
			list:    false,
			errMsg:  "OWNER is not a valid GroupMembershipRole",
		},
		{
			name:    "duplicate user, different role",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   true,
			list:    true,
			errMsg:  "already exists",
		},
		{
			name:    "invalid user",
			groupID: group1.ID,
			userID:  "not-a-valid-user-id",
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   true,
			list:    true,
			errMsg:  "constraint failed",
		},
		{
			name:    "invalid group",
			groupID: "not-a-valid-group-id",
			userID:  orgMember1.UserID,
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   true,
			list:    true,
			errMsg:  "constraint failed",
		},
		{
			name:    "invalid role",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleInvalid,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   false,
			list:    false,
			errMsg:  "not a valid GroupMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errMsg == "" {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			if tc.check {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			if tc.list {
				mock_fga.ListOnce(t, suite.client.fga, []string{fmt.Sprintf("organization:%s", group1.OwnerID)}, nil)
			}

			if tc.errMsg == "" {
				mock_fga.ListOnce(t, suite.client.fga, []string{fmt.Sprintf("group:%s", group1.ID)}, nil)
			}

			role := tc.role
			input := openlaneclient.CreateGroupMembershipInput{
				GroupID: tc.groupID,
				UserID:  tc.userID,
				Role:    &role,
			}

			resp, err := tc.client.AddUserToGroupWithRole(tc.ctx, input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateGroupMembership)
			assert.Equal(t, tc.userID, resp.CreateGroupMembership.GroupMembership.UserID)
			assert.Equal(t, tc.groupID, resp.CreateGroupMembership.GroupMembership.GroupID)
			assert.Equal(t, tc.role, resp.CreateGroupMembership.GroupMembership.Role)
		})
	}

	// delete created group and users
	(&GroupCleanup{client: suite.client, ID: group1.ID}).MustDelete(reqCtx, t)
	(&OrgMemberCleanup{client: suite.client, ID: orgMember1.ID}).MustDelete(reqCtx, t)
	(&OrgMemberCleanup{client: suite.client, ID: orgMember2.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationUpdateGroupMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	orgMember := (&OrgMemberBuilder{client: suite.client, OrgID: testOrgID}).MustNew(reqCtx, t)

	gm := (&GroupMemberBuilder{client: suite.client, UserID: orgMember.UserID}).MustNew(reqCtx, t)

	testCases := []struct {
		name    string
		role    enums.Role
		client  *openlaneclient.OpenlaneClient
		ctx     context.Context
		allowed bool
		check   bool
		errMsg  string
	}{
		{
			name:    "happy path, update to admin from member",
			role:    enums.RoleAdmin,
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			check:   true,
		},
		{
			name:    "happy path, update to member from admin using api token",
			role:    enums.RoleMember,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
			check:   true,
		},
		{
			name:    "happy path, update to admin from member using personal access token",
			role:    enums.RoleAdmin,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
			check:   true,
		},
		{
			name:    "invalid role",
			role:    enums.RoleInvalid,
			client:  suite.client.api,
			ctx:     reqCtx,
			errMsg:  "not a valid GroupMembershipRole",
			allowed: true,
			check:   false,
		},
		{
			name:    "no access",
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     reqCtx,
			errMsg:  "you are not authorized to perform this action",
			allowed: false,
			check:   true,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errMsg == "" {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			if tc.check {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			if tc.errMsg == "" {
				mock_fga.ListOnce(t, suite.client.fga, []string{fmt.Sprintf("group:%s", gm.GroupID)}, nil)
			}

			role := tc.role
			input := openlaneclient.UpdateGroupMembershipInput{
				Role: &role,
			}

			resp, err := tc.client.UpdateUserRoleInGroup(tc.ctx, gm.ID, input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateGroupMembership)
			assert.Equal(t, tc.role, resp.UpdateGroupMembership.GroupMembership.Role)
		})
	}

	// delete created objects
	(&GroupMemberCleanup{client: suite.client, ID: gm.ID}).MustDelete(reqCtx, t)
	(&OrgMemberCleanup{client: suite.client, ID: orgMember.ID}).MustDelete(reqCtx, t)
	(&GroupCleanup{client: suite.client, ID: gm.GroupID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationDeleteGroupMembers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	gm1 := (&GroupMemberBuilder{client: suite.client}).MustNew(reqCtx, t)
	gm2 := (&GroupMemberBuilder{client: suite.client}).MustNew(reqCtx, t)
	gm3 := (&GroupMemberBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		checkAccess bool
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  gm1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: delete on groupmembership",
		},
		{
			name:        "happy path, delete org member",
			idToDelete:  gm1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "group member already deleted, not found",
			idToDelete:  gm1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "group_membership not found",
		},
		{
			name:        "happy path, delete group member using api token",
			idToDelete:  gm2.ID,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "happy path, delete group member using personal access token",
			idToDelete:  gm3.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "unknown group member, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "group_membership not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.expectedErr == "" {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			if tc.checkAccess {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			resp, err := tc.client.RemoveUserFromGroup(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.DeleteGroupMembership)
			assert.Equal(t, tc.idToDelete, resp.DeleteGroupMembership.DeletedID)
		})
	}

}
