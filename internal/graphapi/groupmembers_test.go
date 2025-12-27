package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryGroupMembers(t *testing.T) {
	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	groupMember := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		expected    *ent.GroupMembership
		errExpected bool
	}{
		{
			name:     "happy path, get group member by group id",
			queryID:  group.ID,
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			expected: groupMember,
		},
		{
			name:     "happy path, get group member by group id using api token",
			queryID:  group.ID,
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			expected: groupMember,
		},
		{
			name:     "happy path, get group member by group id using personal access token",
			queryID:  group.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			expected: groupMember,
		},
		{
			name:        "get group member by group id, no access",
			queryID:     group.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expected:    nil, // no results are returned because the group provided is not found for that user
			errExpected: false,
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expected:    nil, // no results are returned because the group provided is not found for that user
			errExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			groupID := tc.queryID
			whereInput := testclient.GroupMembershipWhereInput{
				GroupID: &groupID,
			}
			resp, err := tc.client.GetGroupMembersByGroupID(tc.ctx, &whereInput)

			if tc.errExpected {
				assert.ErrorContains(t, err, notFoundErrorMsg)

				return
			}

			assert.NilError(t, err)

			if tc.expected == nil {
				assert.Check(t, is.Len(resp.GroupMemberships.Edges, 0))

				return
			}

			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.GroupMemberships.Edges != nil)
			assert.Check(t, is.Equal(tc.expected.UserID, resp.GroupMemberships.Edges[0].Node.GetUser().GetID()))
			assert.Check(t, is.Equal(tc.expected.Role, resp.GroupMemberships.Edges[0].Node.Role))
		})
	}

	// delete created group
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group.ID}).MustDelete(testUser1.UserCtx, t)
	// delete created group member
	(&Cleanup[*generated.GroupMembershipDeleteOne]{client: suite.client.db.GroupMembership, ID: groupMember.ID}).MustDelete(testUser1.UserCtx, t)
	// delete org member
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{groupMember.Edges.OrgMembership.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateGroupMembers(t *testing.T) {
	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	checkCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	groupMember, err := group1.QueryMembers().All(checkCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(groupMember, 0))

	orgMember1 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgMember2 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgMember3 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name    string
		groupID string
		userID  string
		role    enums.Role
		client  *testclient.TestClient
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
			name:    "happy path, add self (owner) as admin",
			groupID: group1.ID,
			userID:  testUser1.ID,
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
			name:    "cannot add self to group as org member",
			groupID: group1.ID,
			userID:  viewOnlyUser.UserInfo.ID,
			role:    enums.RoleAdmin,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
			errMsg:  notAuthorizedErrorMsg,
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
			input := testclient.CreateGroupMembershipInput{
				GroupID: tc.groupID,
				UserID:  tc.userID,
				Role:    &role,
			}

			resp, err := tc.client.AddUserToGroupWithRole(tc.ctx, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.userID, resp.CreateGroupMembership.GroupMembership.UserID))
			assert.Check(t, is.Equal(tc.groupID, resp.CreateGroupMembership.GroupMembership.GroupID))
			assert.Check(t, is.Equal(tc.role, resp.CreateGroupMembership.GroupMembership.Role))
		})
	}

	// delete created groups and org members
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{orgMember1.ID, orgMember2.ID, orgMember3.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateGroupMembers(t *testing.T) {
	gm := (&GroupMemberBuilder{client: suite.client, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	// add self to group as admin
	testUser1GroupMember := (&GroupMemberBuilder{client: suite.client, GroupID: testUser1.GroupID, UserID: testUser1.UserInfo.ID, Role: enums.RoleAdmin.String()}).MustNew(testUser1.UserCtx, t)

	gmCtx := auth.NewTestContextWithOrgID(gm.UserID, testUser1.OrganizationID)

	testCases := []struct {
		name          string
		groupMemberID string
		role          enums.Role
		client        *testclient.TestClient
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
			name:          "update self from admin to member ok",
			groupMemberID: testUser1GroupMember.ID,
			role:          enums.RoleMember,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:          "update self from member to admin not allowed",
			groupMemberID: gm.ID,
			role:          enums.RoleMember,
			client:        suite.client.api,
			ctx:           gmCtx,
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
			input := testclient.UpdateGroupMembershipInput{
				Role: &role,
			}

			resp, err := tc.client.UpdateUserRoleInGroup(tc.ctx, tc.groupMemberID, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.role, resp.UpdateGroupMembership.GroupMembership.Role))
		})
	}

	// delete created group member
	(&Cleanup[*generated.GroupMembershipDeleteOne]{client: suite.client.db.GroupMembership, ID: gm.ID}).MustDelete(testUser1.UserCtx, t)
	// delete org member
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{gm.Edges.OrgMembership.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteGroupMembers(t *testing.T) {
	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	gm1 := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)
	gm2 := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)
	gm3 := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	// add self to group as admin
	testUser1GroupMember := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID, UserID: testUser1.UserInfo.ID, Role: enums.RoleAdmin.String()}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
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
			name:       "allowed to delete self as org admin",
			idToDelete: testUser1GroupMember.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
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
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteGroupMembership.DeletedID))
		})
	}

	// delete org members
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{gm1.Edges.OrgMembership.ID, gm2.Edges.OrgMembership.ID, gm3.Edges.OrgMembership.ID}}).MustDelete(testUser1.UserCtx, t)
	// delete the group
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group.ID}).MustDelete(testUser1.UserCtx, t)
}
