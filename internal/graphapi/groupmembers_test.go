package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryGroupMembers(t *testing.T) {
	group := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: group.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			expected: groupMember,
		},
		{
			name:     "happy path, get group member by group id using api token",
			queryID:  group.ID,
			client:   suite.Client.APIWithToken,
			ctx:      context.Background(),
			expected: groupMember,
		},
		{
			name:     "happy path, get group member as auditor",
			queryID:  group.ID,
			client:   suite.Client.API,
			ctx:      th.SharedAuditorUser.UserCtx,
			expected: groupMember,
		},
		{
			name:     "happy path, get group member by group id using personal access token",
			queryID:  group.ID,
			client:   suite.Client.APIWithPAT,
			ctx:      context.Background(),
			expected: groupMember,
		},
		{
			name:        "get group member by group id, no access",
			queryID:     group.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expected:    nil, // no results are returned because the group provided is not found for that user
			errExpected: true,
		},
		{
			name:        "invalid-id",
			queryID:     "tacos-for-dinner",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
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
				assert.ErrorContains(t, err, th.NotFoundErrorMsg)

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
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: group.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete created group member
	(&th.Cleanup[*generated.GroupMembershipDeleteOne]{Client: suite.Client.DB.GroupMembership, ID: groupMember.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete org member
	(&th.Cleanup[*generated.OrgMembershipDeleteOne]{Client: suite.Client.DB.OrgMembership, IDs: []string{groupMember.Edges.OrgMembership.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateGroupMembers(t *testing.T) {
	group1 := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	checkCtx := privacy.DecisionContext(th.SharedTestUser1.UserCtx, privacy.Allow)

	groupMember, err := group1.QueryMembers().All(checkCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(groupMember, 0))

	orgMember1 := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	orgMember2 := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	orgMember3 := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, add self (owner) as admin",
			groupID: group1.ID,
			userID:  th.SharedTestUser1.ID,
			role:    enums.RoleAdmin,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, add member using api token",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleMember,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, add member using personal access token",
			groupID: group1.ID,
			userID:  orgMember3.UserID,
			role:    enums.RoleMember,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:    "cannot add self to group as org member",
			groupID: group1.ID,
			userID:  th.SharedViewOnlyUser.UserInfo.ID,
			role:    enums.RoleAdmin,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
			errMsg:  th.NotAuthorizedErrorMsg,
		},
		{
			name:    "add member, no access",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleMember,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
			errMsg:  th.NotAuthorizedErrorMsg,
		},
		{
			name:    "owner relation not valid for groups",
			groupID: group1.ID,
			userID:  orgMember2.UserID,
			role:    enums.RoleOwner,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			errMsg:  "OWNER is not a valid GroupMembershipRole",
		},
		{
			name:    "duplicate user, different role",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleMember,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			errMsg:  "already exists",
		},
		{
			name:    "invalid user",
			groupID: group1.ID,
			userID:  "not-a-valid-user-id",
			role:    enums.RoleMember,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			errMsg:  th.NotAuthorizedErrorMsg,
		},
		{
			name:    "invalid group",
			groupID: "not-a-valid-group-id",
			userID:  orgMember1.UserID,
			role:    enums.RoleMember,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			errMsg:  th.NotFoundErrorMsg,
		},
		{
			name:    "invalid role",
			groupID: group1.ID,
			userID:  orgMember1.UserID,
			role:    enums.RoleInvalid,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
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
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: group1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.OrgMembershipDeleteOne]{Client: suite.Client.DB.OrgMembership, IDs: []string{orgMember1.ID, orgMember2.ID, orgMember3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateGroupMembers(t *testing.T) {
	gm := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)
	// add self to group as admin
	sharedTestUser1GroupMember := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: th.SharedTestUser1.GroupID, UserID: th.SharedTestUser1.UserInfo.ID, Role: enums.RoleAdmin.String()}).MustNew(th.SharedTestUser1.UserCtx, t)

	gmCtx := auth.NewTestContextWithOrgID(gm.UserID, th.SharedTestUser1.OrganizationID)

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
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
		},
		{
			name:          "update self from admin to member ok",
			groupMemberID: sharedTestUser1GroupMember.ID,
			role:          enums.RoleMember,
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
		},
		{
			name:          "update self from member to admin not allowed",
			groupMemberID: gm.ID,
			role:          enums.RoleMember,
			client:        suite.Client.API,
			ctx:           gmCtx,
			errMsg:        th.NotAuthorizedErrorMsg,
		},
		{
			name:          "happy path, update to member from admin using api token",
			groupMemberID: gm.ID,
			role:          enums.RoleMember,
			client:        suite.Client.APIWithToken,
			ctx:           context.Background(),
		},
		{
			name:          "happy path, update to admin from member using personal access token",
			groupMemberID: gm.ID,
			role:          enums.RoleAdmin,
			client:        suite.Client.APIWithPAT,
			ctx:           context.Background(),
		},
		{
			name:          "invalid role",
			groupMemberID: gm.ID,
			role:          enums.RoleInvalid,
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			errMsg:        "not a valid GroupMembershipRole",
		},
		{
			name:          "no access",
			groupMemberID: gm.ID,
			role:          enums.RoleMember,
			client:        suite.Client.API,
			ctx:           th.SharedViewOnlyUser.UserCtx,
			errMsg:        th.NotAuthorizedErrorMsg,
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
	(&th.Cleanup[*generated.GroupMembershipDeleteOne]{Client: suite.Client.DB.GroupMembership, ID: gm.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete org member
	(&th.Cleanup[*generated.OrgMembershipDeleteOne]{Client: suite.Client.DB.OrgMembership, IDs: []string{gm.Edges.OrgMembership.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteGroupMembers(t *testing.T) {
	group := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	gm1 := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: group.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	gm2 := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: group.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	gm3 := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: group.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add self to group as admin
	sharedTestUser1GroupMember := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: group.ID, UserID: th.SharedTestUser1.UserInfo.ID, Role: enums.RoleAdmin.String()}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:       "allowed to delete self as org admin",
			idToDelete: sharedTestUser1GroupMember.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "not allowed to delete, in another org, not found",
			idToDelete:  gm1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete group member using api token",
			idToDelete: gm2.ID,
			client:     suite.Client.APIWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete org member",
			idToDelete: gm1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:       "happy path, delete group member using personal access token",
			idToDelete: gm3.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown group member, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "group member already deleted, not found",
			idToDelete:  gm1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
	(&th.Cleanup[*generated.OrgMembershipDeleteOne]{Client: suite.Client.DB.OrgMembership, IDs: []string{gm1.Edges.OrgMembership.ID, gm2.Edges.OrgMembership.ID, gm3.Edges.OrgMembership.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete the group
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: group.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
