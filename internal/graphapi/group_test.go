package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/enums"
)

func TestQueryGroup(t *testing.T) {
	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	privateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	privateGroupWithSetting, err := suite.client.api.GetGroupByID(testUser1.UserCtx, privateGroup.ID)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateGroupSetting(testUser1.UserCtx, privateGroupWithSetting.Group.Setting.ID, testclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{

		{
			name:    "happy path group",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			queryID: group1.ID,
		},
		{
			name:    "happy path group, using personal access token",
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			queryID: group1.ID,
		},
		{
			name:    "happy path private group",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			queryID: privateGroup.ID,
		},
		{
			name:     "private group, no access",
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			queryID:  privateGroup.ID,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			queryID:  group1.ID,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  group1.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetGroupByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {

				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.Group.ID != "")
		})
	}

	// delete created group
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{group1.ID, privateGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryGroupsByOwner(t *testing.T) {
	userAnotherOrg := suite.userBuilder(context.Background(), t)
	org1 := userAnotherOrg.OrganizationID
	reqCtx := userAnotherOrg.UserCtx
	group1 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	userAnotherOrg2 := suite.userBuilder(context.Background(), t)
	org2 := userAnotherOrg2.OrganizationID
	reqCtx2 := userAnotherOrg2.UserCtx
	group2 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx2, t)

	t.Run("Get Groups By Owner", func(t *testing.T) {
		whereInput := &testclient.GroupWhereInput{
			HasOwnerWith: []*testclient.OrganizationWhereInput{
				{
					ID: &org1,
				},
			},
			IsManaged: lo.ToPtr(false),
		}

		resp, err := suite.client.api.GetGroups(reqCtx, whereInput)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Groups.Edges != nil)

		// make sure 2 groups are returned, the first was created as part of the seeding process
		assert.Check(t, is.Len(resp.Groups.Edges, 2))

		group1Found := false
		group2Found := false

		for _, o := range resp.Groups.Edges {
			if o.Node.ID == group1.ID {
				group1Found = true
			} else if o.Node.ID == group2.ID {
				group2Found = true
			}
		}

		// group1 should be returned, group 2 should not be returned
		assert.Check(t, group1Found)
		assert.Check(t, !group2Found)

		whereInput = &testclient.GroupWhereInput{
			HasOwnerWith: []*testclient.OrganizationWhereInput{
				{
					ID: &org2,
				},
			},
			IsManaged: lo.ToPtr(false),
		}

		resp, err = suite.client.api.GetGroups(reqCtx2, whereInput)

		assert.NilError(t, err)
		assert.Assert(t, is.Len(resp.Groups.Edges, 2))

	})

	// delete created groups and orgs
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group1.ID}).MustDelete(reqCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group2.ID}).MustDelete(reqCtx2, t)
}

func TestQueryGroups(t *testing.T) {
	testUser1 := suite.userBuilder(context.Background(), t)
	testUser2 := suite.userBuilder(context.Background(), t)

	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group2 := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	group3 := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	privateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	privateGroupWithSetting, err := suite.client.api.GetGroupByID(testUser1.UserCtx, privateGroup.ID)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateGroupSetting(testUser1.UserCtx, privateGroupWithSetting.Group.Setting.ID, testclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)

	t.Run("Get Groups", func(t *testing.T) {
		resp, err := suite.client.api.GetAllGroups(testUser2.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Groups.Edges != nil)

		// make sure two organizations are returned (group 2 and group 3), the seeded group, and the 3 managed groups
		// and the system managed group for the user
		assert.Check(t, is.Equal(7, len(resp.Groups.Edges)))

		group1Found := false
		group2Found := false
		group3Found := false

		for _, o := range resp.Groups.Edges {
			switch id := o.Node.ID; id {
			case group1.ID:
				group1Found = true
			case group2.ID:
				group2Found = true
			case group3.ID:
				group3Found = true
			}
		}

		// if one of the groups isn't found, fail the test
		assert.Check(t, group2Found)
		assert.Check(t, group3Found)

		// if group 1 (which belongs to an unauthorized org) is found, fail the test
		assert.Assert(t, !group1Found)

		// check groups available to testuser1
		resp, err = suite.client.api.GetAllGroups(testUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		// check groups available to admin user (private group created by testUser1 should not be returned)
		resp, err = suite.client.api.GetAllGroups(adminUser.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		for _, v := range resp.Groups.Edges {
			assert.Assert(t, v.Node.ID != privateGroup.ID)
		}

		// check groups available to admin user (private group created by testUser1 should not be returned for org member)
		resp, err = suite.client.api.GetAllGroups(viewOnlyUser.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		for _, v := range resp.Groups.Edges {
			assert.Assert(t, v.Node.ID != privateGroup.ID)
		}

	})

	// delete created groups
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{group1.ID, privateGroup.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{group2.ID, group3.ID}}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateGroup(t *testing.T) {
	name := gofakeit.Name()

	// group for the view only user
	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	createdGroups := []string{group.ID}

	testCases := []struct {
		name          string
		groupName     string
		description   string
		displayName   string
		owner         string
		settings      *testclient.CreateGroupSettingInput
		addGroupToOrg bool
		client        *testclient.TestClient
		ctx           context.Context
		errorMsg      string
	}{
		{
			name:        "happy path group",
			groupName:   name,
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name:        "invalid group name",
			groupName:   name + "!@",
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			errorMsg:    "field cannot contain special character",
		},
		{
			name:        "duplicate group name, case insensitive",
			groupName:   strings.ToUpper(name),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			errorMsg:    "group already exists",
		},
		{
			name:        "happy path group using api token",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(),
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
		},
		{
			name:        "happy path group using personal access token",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			owner:       testUser1.OrganizationID,
			description: gofakeit.HipsterSentence(),
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
		},
		{
			name:        "happy path group with settings",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(),
			settings: &testclient.CreateGroupSettingInput{
				JoinPolicy: &enums.JoinPolicyInviteOnly,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "no access to create group",
			groupName: gofakeit.Name(),
			client:    suite.client.api,
			ctx:       viewOnlyUser.UserCtx,
			errorMsg:  notAuthorizedErrorMsg,
		},
		{
			name:          "group create access added",
			groupName:     gofakeit.Name(),
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name:        "no access to owner, should ignore the input org",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(),
			owner:       testUser2.OrganizationID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name:      "happy path group, minimum fields",
			groupName: gofakeit.Name(),
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name:     "missing name",
			errorMsg: "validator failed",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddGroupCreatorIDs: []string{group.ID},
					}, nil)
				assert.NilError(t, err)
			}

			input := testclient.CreateGroupInput{
				Name:        tc.groupName,
				Description: &tc.description,
				DisplayName: &tc.displayName,
			}

			if tc.owner != "" {
				input.OwnerID = &tc.owner
			}

			if tc.displayName != "" {
				input.DisplayName = &tc.displayName
			}

			if tc.settings != nil {
				input.CreateGroupSettings = tc.settings
			}

			resp, err := tc.client.CreateGroup(tc.ctx, input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.CreateGroup.Group.ID != "")

			// Make sure provided values match
			assert.Check(t, is.Equal(tc.groupName, resp.CreateGroup.Group.Name))
			assert.Check(t, is.Equal(tc.description, *resp.CreateGroup.Group.Description))

			if tc.displayName != "" {
				assert.Check(t, is.Equal(tc.displayName, resp.CreateGroup.Group.DisplayName))
			} else {
				// display name defaults to the name if not set
				assert.Check(t, is.Equal(tc.groupName, resp.CreateGroup.Group.DisplayName))
			}

			if tc.settings != nil {
				assert.Check(t, is.Equal(resp.CreateGroup.Group.Setting.JoinPolicy, enums.JoinPolicyInviteOnly))
			}

			if tc.owner != "" && tc.ctx == testUser2.UserCtx {
				// make sure the owner is ignored if the user doesn't have access
				assert.Check(t, tc.owner != resp.CreateGroup.Group.Owner.ID)
			}

			createdGroups = append(createdGroups, resp.CreateGroup.Group.ID)
		})
	}

	// cleanup the group creator
	_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
		testclient.UpdateOrganizationInput{
			RemoveGroupCreatorIDs: []string{group.ID},
		}, nil)
	assert.NilError(t, err)

	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: createdGroups}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateGroupWithMembers(t *testing.T) {
	testCases := []struct {
		name        string
		group       testclient.CreateGroupInput
		members     []*testclient.GroupMembersInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedLen int
		errorMsg    string
	}{
		{
			name: "happy path group",
			group: testclient.CreateGroupInput{
				Name: ulids.New().String(),
				CreateGroupSettings: &testclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*testclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedLen: 2,
		},
		{
			name: "happy path private group as org admin including self",
			group: testclient.CreateGroupInput{
				Name: ulids.New().String(),
				CreateGroupSettings: &testclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*testclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client:      suite.client.api,
			expectedLen: 2,
			ctx:         adminUser.UserCtx,
		},
		{
			name: "happy path group as org admin including self",
			group: testclient.CreateGroupInput{
				Name: ulids.New().String(),
				CreateGroupSettings: &testclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*testclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client:      suite.client.api,
			expectedLen: 2,
			ctx:         adminUser.UserCtx,
		},
		{
			name: "happy path group as org admin not including self",
			group: testclient.CreateGroupInput{
				Name: ulids.New().String(),
				CreateGroupSettings: &testclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*testclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
			},
			client:      suite.client.api,
			expectedLen: 1,
			ctx:         adminUser.UserCtx,
		},
		{
			name: "happy path group using api token with same members",
			group: testclient.CreateGroupInput{
				Name: gofakeit.Name(),
				CreateGroupSettings: &testclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*testclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client:      suite.client.apiWithToken,
			expectedLen: 2,
			ctx:         context.Background(),
		},
		{
			name: "happy path group using personal access token with same members",
			group: testclient.CreateGroupInput{
				Name:    ulids.New().String(),
				OwnerID: &testUser1.OrganizationID,
				CreateGroupSettings: &testclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*testclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client:      suite.client.apiWithPAT,
			expectedLen: 2,
			ctx:         context.Background(),
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateGroupWithMembers(tc.ctx, tc.group, tc.members)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.CreateGroupWithMembers.Group.ID != "")

			// Make sure provided values match
			assert.Check(t, is.Equal(tc.group.Name, resp.CreateGroupWithMembers.Group.Name))

			// ensure we can still set the visibility on the group when creating it
			assert.Check(t, is.Equal(resp.CreateGroupWithMembers.Group.Setting.Visibility, enums.VisibilityPrivate))

			assert.Assert(t, is.Len(resp.CreateGroupWithMembers.Group.Members.Edges, tc.expectedLen))

			// make sure we get the member data back
			for _, member := range tc.members {
				members := resp.CreateGroupWithMembers.Group.Members.Edges
				found := false
				for _, m := range members {
					assert.Assert(t, m.Node.User.ID != "")

					if m.Node.User.ID == member.UserID {
						found = true
						assert.Check(t, is.Equal(*member.Role, m.Node.Role))

						assert.Check(t, m.Node.User.FirstName != nil)
						assert.Check(t, m.Node.User.LastName != nil)
					}
				}

				assert.Check(t, found, "member %s not found", member.UserID)
			}

			// cleanup using the api
			_, err = tc.client.DeleteGroup(tc.ctx, resp.CreateGroupWithMembers.Group.ID)
			assert.NilError(t, err)
		})
	}
}

func TestMutationCreateGroupByClone(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group := (&GroupBuilder{client: suite.client, ProgramEditorsIDs: []string{program.ID}, ControlEditorsIDs: []string{control.ID}}).MustNew(testUser1.UserCtx, t)

	groupAnotherUser := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// add a group member to the group
	(&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                  string
		group                 testclient.CreateGroupInput
		groupPermissionsClone *string
		groupMembersClone     *string
		members               []*testclient.GroupMembersInput
		client                *testclient.TestClient
		ctx                   context.Context
		errorMsg              string
	}{
		{
			name: "happy path, clone group everything",
			group: testclient.CreateGroupInput{
				Name: gofakeit.Name(),
			},
			groupPermissionsClone: &group.ID,
			groupMembersClone:     &group.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
		},
		{
			name: "happy path, clone group members, use personal access token",
			group: testclient.CreateGroupInput{
				Name:    gofakeit.Name(),
				OwnerID: &testUser1.OrganizationID,
			},
			groupPermissionsClone: &group.ID,
			client:                suite.client.apiWithPAT,
			ctx:                   context.Background(),
		},
		{
			name: "happy path, clone group permissions, use api token",
			group: testclient.CreateGroupInput{
				Name: gofakeit.Name(),
			},
			groupMembersClone: &group.ID,
			client:            suite.client.apiWithToken,
			ctx:               context.Background(),
		},
		{
			name: "clone group everything, but view only user",
			group: testclient.CreateGroupInput{
				Name: gofakeit.Name(),
			},
			groupPermissionsClone: &group.ID,
			groupMembersClone:     &group.ID,
			client:                suite.client.api,
			ctx:                   viewOnlyUser.UserCtx,
			errorMsg:              notAuthorizedErrorMsg,
		},
		{
			name: "clone group everything, no access to clone group",
			group: testclient.CreateGroupInput{
				Name: gofakeit.Name(),
			},
			groupPermissionsClone: &groupAnotherUser.ID,
			groupMembersClone:     &groupAnotherUser.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			errorMsg:              notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateGroupByClone(tc.ctx, tc.group, tc.groupPermissionsClone, tc.groupMembersClone)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.CreateGroupByClone.Group.ID != "")

			// the display id should be different
			assert.Assert(t, group.DisplayID != resp.CreateGroupByClone.Group.DisplayID)

			// make sure there the is the member from the group we cloned
			expectedLen := 0
			if tc.groupMembersClone != nil {
				expectedLen += 1
			}

			assert.Check(t, is.Len(resp.CreateGroupByClone.Group.Members.Edges, expectedLen))

			// added a control and a program to the group we cloned, make sure they are there
			expectedLenPerms := 0
			if tc.groupPermissionsClone != nil {
				expectedLenPerms = 2
			}

			assert.Check(t, is.Len(resp.CreateGroupByClone.Group.Permissions.Edges, expectedLenPerms))

			// delete group via the api
			_, err = tc.client.DeleteGroup(tc.ctx, resp.CreateGroupByClone.Group.ID)
			assert.NilError(t, err)
		})
	}

	// cleanup
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: groupAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateGroup(t *testing.T) {
	nameUpdate := gofakeit.Name()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence()
	gravatarURLUpdate := gofakeit.URL()

	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	gm := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	// create a second group member to test removing and re-adding
	group2 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	gm2 := (&GroupMemberBuilder{client: suite.client, GroupID: group2.ID}).MustNew(testUser1.UserCtx, t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	programClone := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlClone := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add additional permissions as well as the same we will be updating to the group (control ID)
	groupClone := (&GroupBuilder{client: suite.client, ProgramEditorsIDs: []string{programClone.ID}, ControlEditorsIDs: []string{controlClone.ID, control.ID}}).MustNew(testUser1.UserCtx, t)

	gmCtx := auth.NewTestContextWithOrgID(gm.UserID, testUser1.OrganizationID)

	// ensure user cannot get access to the program
	_, err := suite.client.api.GetProgramByID(gmCtx, program.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	// access to procedures is granted by default in the org
	procedureResp, err := suite.client.api.GetProcedureByID(gmCtx, procedure.ID)
	assert.NilError(t, err)
	assert.Assert(t, procedureResp != nil)

	testCases := []struct {
		name        string
		groupID     string
		updateInput testclient.UpdateGroupInput
		expectedRes testclient.UpdateGroup_UpdateGroup_Group
		client      *testclient.TestClient
		ctx         context.Context
		errorMsg    string
	}{
		{
			name:    "add permissions to object, happy path",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				AddProgramViewerIDs:         []string{program.ID},
				AddProcedureBlockedGroupIDs: []string{procedure.ID},
				AddControlEditorIDs:         []string{control.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        group.Name,
				DisplayName: group.DisplayName,
				Description: &group.Description,
				Setting: &testclient.UpdateGroup_UpdateGroup_Group_Setting{
					JoinPolicy: enums.JoinPolicyOpen,
				},
				Permissions: testclient.UpdateGroup_UpdateGroup_Group_Permissions{
					Edges: []*testclient.UpdateGroup_UpdateGroup_Group_Permissions_Edges{
						{
							Node: &testclient.UpdateGroup_UpdateGroup_Group_Permissions_Edges_Node{
								ObjectType:  "Program",
								ID:          program.ID,
								Permissions: enums.Viewer,
								DisplayID:   &program.DisplayID,
								Name:        &program.Name,
							},
						},
						{
							Node: &testclient.UpdateGroup_UpdateGroup_Group_Permissions_Edges_Node{
								ObjectType:  "Procedure",
								ID:          procedure.ID,
								Permissions: enums.Blocked,
								DisplayID:   &procedure.DisplayID,
								Name:        &procedure.Name,
							},
						},
						{
							Node: &testclient.UpdateGroup_UpdateGroup_Group_Permissions_Edges_Node{
								ObjectType:  "Control",
								ID:          control.ID,
								Permissions: enums.Editor,
								DisplayID:   &control.DisplayID,
								Name:        &control.RefCode,
							},
						},
					},
				},
			},
		},
		{
			name:    "add permissions to object, no access to program",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				AddProgramEditorIDs: []string{program.ID},
			},
			client:   suite.client.api,
			ctx:      adminUser.UserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:    "update name and clone permissions, happy path - this will add two permissions to the group",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				Name:                    &nameUpdate,
				DisplayName:             &displayNameUpdate,
				Description:             &descriptionUpdate,
				InheritGroupPermissions: &groupClone.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
			},
		},
		{
			name:    "add user as admin using api token",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				AddGroupMembers: []*testclient.CreateGroupMembershipInput{
					{
						UserID: om.UserID,
						Role:   &enums.RoleAdmin,
					},
				},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				Members: testclient.UpdateGroup_UpdateGroup_Group_Members{
					Edges: []*testclient.UpdateGroup_UpdateGroup_Group_Members_Edges{
						{
							Node: &testclient.UpdateGroup_UpdateGroup_Group_Members_Edges_Node{
								Role: enums.RoleAdmin,
								User: testclient.UpdateGroup_UpdateGroup_Group_Members_Edges_Node_User{
									ID: om.UserID,
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "remove group member",
			groupID: group2.ID,
			updateInput: testclient.UpdateGroupInput{
				RemoveGroupMembers: []string{gm2.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group2.ID,
				DisplayID:   group2.DisplayID,
				Name:        group2.Name,
				DisplayName: group2.DisplayName,
				Description: &group2.Description,
			},
		},
		{
			name:    "re-add group member",
			groupID: group2.ID,
			updateInput: testclient.UpdateGroupInput{
				AddGroupMembers: []*testclient.CreateGroupMembershipInput{
					{
						UserID: gm2.UserID,
						Role:   &gm2.Role,
					},
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group2.ID,
				DisplayID:   group2.DisplayID,
				Name:        group2.Name,
				DisplayName: group2.DisplayName,
				Description: &group2.Description,
			},
		},
		{
			name:    "update gravatar, happy path using personal access token",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				LogoURL: &gravatarURLUpdate,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
			},
		},
		{
			name:    "update visibility",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				UpdateGroupSettings: &testclient.UpdateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
				Setting: &testclient.UpdateGroup_UpdateGroup_Group_Setting{
					Visibility: enums.VisibilityPrivate,
				},
			},
		},
		{
			name:    "update visibility, same setting",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				UpdateGroupSettings: &testclient.UpdateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
				Setting: &testclient.UpdateGroup_UpdateGroup_Group_Setting{
					Visibility: enums.VisibilityPrivate,
				},
			},
		},
		{
			name:    "update visibility, back to public",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				UpdateGroupSettings: &testclient.UpdateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
				Setting: &testclient.UpdateGroup_UpdateGroup_Group_Setting{
					Visibility: enums.VisibilityPublic,
				},
			},
		},
		{
			name:    "update settings, happy path",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				UpdateGroupSettings: &testclient.UpdateGroupSettingInput{
					JoinPolicy: &enums.JoinPolicyOpen,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				DisplayID:   group.DisplayID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				Setting: &testclient.UpdateGroup_UpdateGroup_Group_Setting{
					JoinPolicy: enums.JoinPolicyOpen,
				},
			},
		},
		{
			name:    "no access",
			groupID: group.ID,
			updateInput: testclient.UpdateGroupInput{
				Name:        &nameUpdate,
				DisplayName: &displayNameUpdate,
				Description: &descriptionUpdate,
			},
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: "not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateGroup(tc.ctx, tc.groupID, tc.updateInput)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.UpdateGroup.Group.ID != "")

			// Make sure provided values match
			updatedGroup := resp.GetUpdateGroup().Group
			assert.Check(t, is.Equal(tc.expectedRes.Name, updatedGroup.Name))
			assert.Check(t, is.Equal(tc.expectedRes.DisplayName, updatedGroup.DisplayName))
			assert.Check(t, is.DeepEqual(tc.expectedRes.Description, updatedGroup.Description))

			// ensure the displayID is not updated
			assert.Check(t, is.Equal(tc.expectedRes.DisplayID, updatedGroup.DisplayID))
			if tc.updateInput.LogoURL != nil {
				assert.Check(t, is.Equal(*tc.expectedRes.LogoURL, *updatedGroup.LogoURL))
			}

			if tc.updateInput.AddGroupMembers != nil && tc.groupID == group.ID {
				assert.Check(t, is.Len(updatedGroup.Members.Edges, 2))
				assert.Check(t, is.Equal(tc.expectedRes.Members.Edges[0].Node.Role, updatedGroup.Members.Edges[1].Node.Role))
				assert.Check(t, is.Equal(tc.expectedRes.Members.Edges[0].Node.User.ID, updatedGroup.Members.Edges[1].Node.User.ID))
			}

			if tc.updateInput.UpdateGroupSettings != nil {
				if tc.updateInput.UpdateGroupSettings.JoinPolicy != nil {
					assert.Check(t, is.Equal(updatedGroup.GetSetting().JoinPolicy, enums.JoinPolicyOpen))
				}

				if tc.updateInput.UpdateGroupSettings.Visibility != nil {
					assert.Check(t, is.Equal(updatedGroup.GetSetting().Visibility, *tc.updateInput.UpdateGroupSettings.Visibility))
				}
			}

			if tc.updateInput.AddProgramViewerIDs != nil || tc.updateInput.AddProcedureEditorIDs != nil || tc.updateInput.AddControlBlockedGroupIDs != nil {
				assert.Check(t, is.Equal(len(tc.expectedRes.Permissions.Edges), len(updatedGroup.Permissions.Edges)))

				for _, permission := range updatedGroup.Permissions.Edges {
					found := false
					for _, expectedPermission := range tc.expectedRes.Permissions.Edges {
						if permission.Node.ID == expectedPermission.Node.ID {
							found = true
							assert.Check(t, is.DeepEqual(permission, expectedPermission))
						}
					}

					assert.Check(t, found, "permission %s not found", permission.Node.ObjectType)
				}

				// ensure user can now get access to the program
				programResp, err := suite.client.api.GetProgramByID(gmCtx, program.ID)
				assert.NilError(t, err)
				assert.Assert(t, programResp != nil)

				// ensure user can now access the control (they have editor access and should be able to make changes)
				description := gofakeit.HipsterSentence()
				controlResp, err := suite.client.api.UpdateControl(gmCtx, control.ID, testclient.UpdateControlInput{
					Description: &description,
				})
				assert.NilError(t, err)
				assert.Assert(t, controlResp != nil)
				assert.Check(t, is.Equal(description, *controlResp.UpdateControl.Control.Description))

				// access to procedures is granted by default in the org, it should be blocked now
				_, err = suite.client.api.GetProcedureByID(gmCtx, procedure.ID)
				assert.ErrorContains(t, err, notFoundErrorMsg)
			}

			if tc.updateInput.InheritGroupPermissions != nil {
				// ensure the group has the additional permissions as the group we cloned, there is one overlap with the group we cloned
				assert.Check(t, is.Len(updatedGroup.Permissions.Edges, 5))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{om.ID, gm.Edges.OrgMembership.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{group.ID, groupClone.ID, group2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program.ID, programClone.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control.ID, controlClone.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedure.ID}).MustDelete(testUser1.UserCtx, t)

}

func TestMutationDeleteGroup(t *testing.T) {
	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group2 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group3 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	privateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	privateGroupWithSetting, err := suite.client.api.GetGroupByID(testUser1.UserCtx, privateGroup.ID)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateGroupSetting(testUser1.UserCtx, privateGroupWithSetting.Group.Setting.ID, testclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name     string
		groupID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "delete private group, happy path",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			groupID: privateGroup.ID,
		},
		{
			name:    "delete group, happy path using api token",
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			groupID: group2.ID,
		},
		{
			name:    "delete group, happy path using personal access token",
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			groupID: group3.ID,
		},
		{
			name:     "delete group, no access",
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			groupID:  group1.ID,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:    "delete group, happy path",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			groupID: group1.ID,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteGroup(tc.ctx, tc.groupID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.DeleteGroup.DeletedID != "")

			// make sure the deletedID matches the ID we wanted to delete
			assert.Check(t, is.Equal(tc.groupID, resp.DeleteGroup.DeletedID))
		})
	}
}

func TestManagedGroups(t *testing.T) {
	whereInput := &testclient.GroupWhereInput{
		IsManaged: lo.ToPtr(true),
	}

	testUser := suite.userBuilder(context.Background(), t)

	resp, err := suite.client.api.GetGroupInfo(testUser.UserCtx, whereInput)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// there should be 4 managed groups created by the system on org creation
	// one for the user
	assert.Check(t, is.Len(resp.Groups.Edges, 4))

	// you should not be able to update a managed group
	groupID := resp.Groups.Edges[0].Node.ID
	input := testclient.UpdateGroupInput{
		Tags: []string{"test"},
	}

	_, err = suite.client.api.UpdateGroup(testUser.UserCtx, groupID, input)
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should not be able to add group members to a managed group
	_, err = suite.client.api.AddUserToGroupWithRole(testUser.UserCtx, testclient.CreateGroupMembershipInput{
		GroupID: groupID,
		UserID:  testUser2.ID,
	})
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should not be able to delete a managed group
	_, err = suite.client.api.DeleteGroup(testUser.UserCtx, groupID)
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should, however, be able to update permissions edges on a managed group
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	policy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	input = testclient.UpdateGroupInput{
		AddProgramViewerIDs:              []string{program.ID},
		AddControlEditorIDs:              []string{control.ID},
		AddInternalPolicyBlockedGroupIDs: []string{policy.ID},
	}

	updateResp, err := suite.client.api.UpdateGroup(testUser.UserCtx, groupID, input)
	assert.NilError(t, err)

	perms := updateResp.UpdateGroup.Group.GetPermissions()
	assert.Check(t, is.Len(perms.Edges, 3))

	// make sure I can also remove them
	input = testclient.UpdateGroupInput{
		RemoveProgramViewerIDs:              []string{program.ID},
		RemoveControlEditorIDs:              []string{control.ID},
		RemoveInternalPolicyBlockedGroupIDs: []string{policy.ID},
	}

	updateResp, err = suite.client.api.UpdateGroup(testUser.UserCtx, groupID, input)
	assert.NilError(t, err)

	perms = updateResp.UpdateGroup.Group.GetPermissions()
	assert.Check(t, is.Len(perms.Edges, 0))

	// cleanup objects created
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: policy.ID}).MustDelete(testUser.UserCtx, t)
}
