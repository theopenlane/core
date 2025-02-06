package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryGroupMembers() {
	t := suite.T()

	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	checkCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	groupMember, err := group.Members(checkCtx)
	require.NoError(t, err)
	require.Len(t, groupMember, 1)

	testCases := []struct {
		name        string
		queryID     string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expected    *ent.GroupMembership
		errExpected bool
	}{
		{
			name:     "happy path, get group member by group id",
			queryID:  group.ID,
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			expected: groupMember[0],
		},
		{
			name:     "happy path, get group member by group id using api token",
			queryID:  group.ID,
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			expected: groupMember[0],
		},
		{
			name:     "happy path, get group member by group id using personal access token",
			queryID:  group.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			expected: groupMember[0],
		},
		{
			name:        "get group member by group id, no access",
			queryID:     group.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expected:    nil,
			errExpected: true,
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expected:    nil,
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			groupID := tc.queryID
			whereInput := openlaneclient.GroupMembershipWhereInput{
				GroupID: &groupID,
			}
			resp, err := tc.client.GetGroupMembersByGroupID(tc.ctx, &whereInput)

			if tc.errExpected {
				require.Error(t, err)
				assert.ErrorContains(t, err, notFoundErrorMsg)

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
	(&GroupCleanup{client: suite.client, ID: group.ID}).MustDelete(testUser1.UserCtx, t)
}

func (suite *GraphTestSuite) TestMutationCreateGroupMembers() {
	t := suite.T()

	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	checkCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	groupMember, err := group1.Members(checkCtx)
	require.NoError(t, err)
	require.Len(t, groupMember, 1)

	orgMember1 := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	orgMember2 := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	orgMember3 := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name    string
		groupID string
		userID  string
		role    enums.Role
		client  *openlaneclient.OpenlaneClient
		ctx     context.Context
		errMsg  string
	}{
		{
			name:    "happy path, add admin",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleAdmin,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, add member using api token",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleMember,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, add member using personal access token",
			groupID: group1.ID,
			userID:  orgMember3.UserID,
			role:    enums.RoleMember,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:    "add member, no access",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
			errMsg:  notAuthorizedErrorMsg,
		},
		{
			name:    "owner relation not valid for groups",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleOwner,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			errMsg:  "OWNER is not a valid GroupMembershipRole",
		},
		{
			name:    "duplicate user, different role",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			errMsg:  "already exists",
		},
		{
			name:    "invalid user",
			groupID: group1.ID,
			userID:  "not-a-valid-user-id",
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			errMsg:  "user not in organization",
		},
		{
			name:    "invalid group",
			groupID: "not-a-valid-group-id",
			userID:  orgMember1.UserID,
			role:    enums.RoleMember,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			errMsg:  notAuthorizedErrorMsg,
		},
		{
			name:    "invalid role",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleInvalid,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			errMsg:  "not a valid GroupMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
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
	(&GroupCleanup{client: suite.client, ID: group1.ID}).MustDelete(testUser1.UserCtx, t)
	(&OrgMemberCleanup{client: suite.client, ID: orgMember1.ID}).MustDelete(testUser1.UserCtx, t)
	(&OrgMemberCleanup{client: suite.client, ID: orgMember2.ID}).MustDelete(testUser1.UserCtx, t)
}

func (suite *GraphTestSuite) TestMutationUpdateGroupMembers() {
	t := suite.T()

	gm := (&GroupMemberBuilder{client: suite.client, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// get all group members so we know the id of the test user group member
	groupMembers, err := suite.client.api.GetGroupMembersByGroupID(testUser1.UserCtx, &openlaneclient.GroupMembershipWhereInput{
		GroupID: &testUser1.GroupID,
	})

	require.NoError(t, err)

	testUser1GroupMember := ""
	for _, gm := range groupMembers.GroupMemberships.Edges {
		if gm.Node.UserID == testUser1.UserInfo.ID {
			testUser1GroupMember = gm.Node.ID
			break
		}
	}

	testCases := []struct {
		name          string
		groupMemberID string
		role          enums.Role
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errMsg        string
	}{
		{
			name:          "happy path, update to admin from member",
			groupMemberID: gm.ID,
			role:          enums.RoleAdmin,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:          "update self from admin to member, not allowed",
			groupMemberID: testUser1GroupMember,
			role:          enums.RoleMember,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			errMsg:        notAuthorizedErrorMsg,
		},
		{
			name:          "happy path, update to member from admin using api token",
			groupMemberID: gm.ID,
			role:          enums.RoleMember,
			client:        suite.client.apiWithToken,
			ctx:           context.Background(),
		},
		{
			name:          "happy path, update to admin from member using personal access token",
			groupMemberID: gm.ID,
			role:          enums.RoleAdmin,
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
		},
		{
			name:          "invalid role",
			groupMemberID: gm.ID,
			role:          enums.RoleInvalid,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			errMsg:        "not a valid GroupMembershipRole",
		},
		{
			name:          "no access",
			groupMemberID: gm.ID,
			role:          enums.RoleMember,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
			errMsg:        notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			role := tc.role
			input := openlaneclient.UpdateGroupMembershipInput{
				Role: &role,
			}

			resp, err := tc.client.UpdateUserRoleInGroup(tc.ctx, tc.groupMemberID, input)

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
}

func (suite *GraphTestSuite) TestMutationDeleteGroupMembers() {
	t := suite.T()

	gm1 := (&GroupMemberBuilder{client: suite.client, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	gm2 := (&GroupMemberBuilder{client: suite.client, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	gm3 := (&GroupMemberBuilder{client: suite.client, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// get all group members so we know the id of the test user group member
	groupMembers, err := suite.client.api.GetGroupMembersByGroupID(testUser1.UserCtx, &openlaneclient.GroupMembershipWhereInput{
		GroupID: &testUser1.GroupID,
	})

	require.NoError(t, err)

	testUser1GroupMember := ""
	for _, gm := range groupMembers.GroupMemberships.Edges {
		if gm.Node.UserID == testUser1.UserInfo.ID {
			testUser1GroupMember = gm.Node.ID
			break
		}
	}

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  gm1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete self",
			idToDelete:  testUser1GroupMember,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete, not found",
			idToDelete:  gm1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete group member using api token",
			idToDelete: gm2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete org member",
			idToDelete: gm1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:       "happy path, delete group member using personal access token",
			idToDelete: gm3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown group member, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "group member already deleted, not found",
			idToDelete:  gm1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
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
