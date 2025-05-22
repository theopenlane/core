package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
)

func TestQueryGroup(t *testing.T) {
	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	privateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	privateGroupWithSetting, err := suite.client.api.GetGroupByID(testUser1.UserCtx, privateGroup.ID)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateGroupSetting(testUser1.UserCtx, privateGroupWithSetting.Group.Setting.ID, openlaneclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
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
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetGroupByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {

				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

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
		whereInput := &openlaneclient.GroupWhereInput{
			HasOwnerWith: []*openlaneclient.OrganizationWhereInput{
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

		whereInput = &openlaneclient.GroupWhereInput{
			HasOwnerWith: []*openlaneclient.OrganizationWhereInput{
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
	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group2 := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	group3 := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	privateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	privateGroupWithSetting, err := suite.client.api.GetGroupByID(testUser1.UserCtx, privateGroup.ID)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateGroupSetting(testUser1.UserCtx, privateGroupWithSetting.Group.Setting.ID, openlaneclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)

	t.Run("Get Groups", func(t *testing.T) {
		resp, err := suite.client.api.GetAllGroups(testUser2.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Groups.Edges != nil)

		// make sure two organizations are returned (group 2 and group 3), the seeded group, and the 3 managed groups
		assert.Check(t, is.Equal(6, len(resp.Groups.Edges)))

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

		// make sure only 5 groups are returned, group 1 and the seeded group, and the 3 managed groups
		assert.Check(t, is.Equal(5, len(resp.Groups.Edges)))
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
		settings      *openlaneclient.CreateGroupSettingInput
		addGroupToOrg bool
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
	}{
		{
			name:        "happy path group",
			groupName:   name,
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name:        "duplicate group name, case insensitive",
			groupName:   strings.ToUpper(name),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			errorMsg:    "group already exists",
		},
		{
			name:        "happy path group using api token",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
		},
		{
			name:        "happy path group using personal access token",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			owner:       testUser1.OrganizationID,
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
		},
		{
			name:        "happy path group with settings",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			settings: &openlaneclient.CreateGroupSettingInput{
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
			description: gofakeit.HipsterSentence(10),
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
					openlaneclient.UpdateOrganizationInput{
						AddGroupCreatorIDs: []string{group.ID},
					}, nil)
				assert.NilError(t, err)
			}

			input := openlaneclient.CreateGroupInput{
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
				assert.Check(t, is.Nil(resp))

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
		openlaneclient.UpdateOrganizationInput{
			RemoveGroupCreatorIDs: []string{group.ID},
		}, nil)
	assert.NilError(t, err)

	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: createdGroups}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateGroupWithMembers(t *testing.T) {
	testCases := []struct {
		name     string
		group    openlaneclient.CreateGroupInput
		members  []*openlaneclient.GroupMembersInput
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name: "happy path group",
			group: openlaneclient.CreateGroupInput{
				Name: gofakeit.Name(),
				CreateGroupSettings: &openlaneclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*openlaneclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path group using api token with same members",
			group: openlaneclient.CreateGroupInput{
				Name: gofakeit.Name(),
				CreateGroupSettings: &openlaneclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*openlaneclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path group using personal access token with same members",
			group: openlaneclient.CreateGroupInput{
				Name:    gofakeit.Name(),
				OwnerID: &testUser1.OrganizationID,
				CreateGroupSettings: &openlaneclient.CreateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			members: []*openlaneclient.GroupMembersInput{
				{
					UserID: adminUser.ID,
					Role:   &enums.RoleAdmin,
				},
				{
					UserID: viewOnlyUser.ID,
					Role:   &enums.RoleMember,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateGroupWithMembers(tc.ctx, tc.group, tc.members)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.CreateGroupWithMembers.Group.ID != "")

			// Make sure provided values match
			assert.Check(t, is.Equal(tc.group.Name, resp.CreateGroupWithMembers.Group.Name))

			// ensure we can still set the visibility on the group when creating it
			assert.Check(t, is.Equal(resp.CreateGroupWithMembers.Group.Setting.Visibility, enums.VisibilityPrivate))

			// make sure there are three members, user who created the group, admin, and member
			// except when using an api token
			expectedLen := 3
			if tc.client == suite.client.apiWithToken {
				expectedLen = 2
			}

			assert.Assert(t, is.Len(resp.CreateGroupWithMembers.Group.Members, expectedLen))

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
		group                 openlaneclient.CreateGroupInput
		groupPermissionsClone *string
		groupMembersClone     *string
		members               []*openlaneclient.GroupMembersInput
		client                *openlaneclient.OpenlaneClient
		ctx                   context.Context
		errorMsg              string
	}{
		{
			name: "happy path, clone group everything",
			group: openlaneclient.CreateGroupInput{
				Name: gofakeit.Name(),
			},
			groupPermissionsClone: &group.ID,
			groupMembersClone:     &group.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
		},
		{
			name: "happy path, clone group members, use personal access token",
			group: openlaneclient.CreateGroupInput{
				Name:    gofakeit.Name(),
				OwnerID: &testUser1.OrganizationID,
			},
			groupPermissionsClone: &group.ID,
			client:                suite.client.apiWithPAT,
			ctx:                   context.Background(),
		},
		{
			name: "happy path, clone group permissions, use api token",
			group: openlaneclient.CreateGroupInput{
				Name: gofakeit.Name(),
			},
			groupMembersClone: &group.ID,
			client:            suite.client.apiWithToken,
			ctx:               context.Background(),
		},
		{
			name: "clone group everything, but view only user",
			group: openlaneclient.CreateGroupInput{
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
			group: openlaneclient.CreateGroupInput{
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
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.CreateGroupByClone.Group.ID != "")

			// the display id should be different
			assert.Assert(t, group.DisplayID != resp.CreateGroupByClone.Group.DisplayID)

			// make sure there are two members, user who created the group and the cloned member
			// even when an api token is used, there will still be the original user (testUser1)
			expectedLen := 1
			if tc.groupMembersClone != nil {
				expectedLen += 1
			}

			assert.Check(t, is.Len(resp.CreateGroupByClone.Group.Members, expectedLen))

			// added a control and a program to the group we cloned, make sure they are there
			expectedLenPerms := 0
			if tc.groupPermissionsClone != nil {
				expectedLenPerms = 2
			}

			assert.Check(t, is.Len(resp.CreateGroupByClone.Group.Permissions, expectedLenPerms))

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
	descriptionUpdate := gofakeit.HipsterSentence(10)
	gravatarURLUpdate := gofakeit.URL()

	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	gm := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

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
	programResp, err := suite.client.api.GetProgramByID(gmCtx, program.ID)

	assert.Assert(t, is.Nil(programResp))

	// ensure user cannot get access to the control
	controlResp, err := suite.client.api.GetControlByID(gmCtx, control.ID)

	assert.Assert(t, is.Nil(controlResp))

	// access to procedures is granted by default in the org
	procedureResp, err := suite.client.api.GetProcedureByID(gmCtx, procedure.ID)
	assert.NilError(t, err)
	assert.Assert(t, procedureResp != nil)

	testCases := []struct {
		name        string
		updateInput openlaneclient.UpdateGroupInput
		expectedRes openlaneclient.UpdateGroup_UpdateGroup_Group
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		errorMsg    string
	}{
		{
			name: "add permissions to object, happy path",
			updateInput: openlaneclient.UpdateGroupInput{
				AddProgramViewerIDs:         []string{program.ID},
				AddProcedureBlockedGroupIDs: []string{procedure.ID},
				AddControlEditorIDs:         []string{control.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        group.Name,
				DisplayName: group.DisplayName,
				Description: &group.Description,
				Setting: &openlaneclient.UpdateGroup_UpdateGroup_Group_Setting{
					JoinPolicy: enums.JoinPolicyOpen,
				},
				Permissions: []*openlaneclient.UpdateGroup_UpdateGroup_Group_Permissions{
					{
						ObjectType:  "Program",
						ID:          &program.ID,
						Permissions: enums.Viewer,
						DisplayID:   &program.DisplayID,
						Name:        &program.Name,
					},
					{
						ObjectType:  "Procedure",
						ID:          &procedure.ID,
						Permissions: enums.Blocked,
						DisplayID:   &procedure.DisplayID,
						Name:        &procedure.Name,
					},
					{
						ObjectType:  "Control",
						ID:          &control.ID,
						Permissions: enums.Editor,
						DisplayID:   &control.DisplayID,
						Name:        &control.RefCode,
					},
				},
			},
		},
		{
			name: "add permissions to object, no access to program",
			updateInput: openlaneclient.UpdateGroupInput{
				AddProgramEditorIDs: []string{program.ID},
			},
			client:   suite.client.api,
			ctx:      adminUser.UserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name: "update name and clone permissions, happy path - this will add two permissions to the group",
			updateInput: openlaneclient.UpdateGroupInput{
				Name:                    &nameUpdate,
				DisplayName:             &displayNameUpdate,
				Description:             &descriptionUpdate,
				InheritGroupPermissions: &groupClone.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
			},
		},
		{
			name: "add user as admin using api token",
			updateInput: openlaneclient.UpdateGroupInput{
				AddGroupMembers: []*openlaneclient.CreateGroupMembershipInput{
					{
						UserID: om.UserID,
						Role:   &enums.RoleAdmin,
					},
				},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				Members: openlaneclient.UpdateGroup_UpdateGroup_Group_Members{
					Edges: []*openlaneclient.UpdateGroup_UpdateGroup_Group_Members_Edges{
						{
							Node: &openlaneclient.UpdateGroup_UpdateGroup_Group_Members_Edges_Node{
								Role: enums.RoleAdmin,
								User: openlaneclient.UpdateGroup_UpdateGroup_Group_Members_Edges_Node_User{
									ID: om.UserID,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "update gravatar, happy path using personal access token",
			updateInput: openlaneclient.UpdateGroupInput{
				LogoURL: &gravatarURLUpdate,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
			},
		},
		{
			name: "update visibility",
			updateInput: openlaneclient.UpdateGroupInput{
				UpdateGroupSettings: &openlaneclient.UpdateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
				Setting: &openlaneclient.UpdateGroup_UpdateGroup_Group_Setting{
					Visibility: enums.VisibilityPrivate,
				},
			},
		},
		{
			name: "update visibility, same setting",
			updateInput: openlaneclient.UpdateGroupInput{
				UpdateGroupSettings: &openlaneclient.UpdateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
				Setting: &openlaneclient.UpdateGroup_UpdateGroup_Group_Setting{
					Visibility: enums.VisibilityPrivate,
				},
			},
		},
		{
			name: "update visibility, back to public",
			updateInput: openlaneclient.UpdateGroupInput{
				UpdateGroupSettings: &openlaneclient.UpdateGroupSettingInput{
					Visibility: &enums.VisibilityPrivate,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
				Setting: &openlaneclient.UpdateGroup_UpdateGroup_Group_Setting{
					Visibility: enums.VisibilityPublic,
				},
			},
		},
		{
			name: "update settings, happy path",
			updateInput: openlaneclient.UpdateGroupInput{
				UpdateGroupSettings: &openlaneclient.UpdateGroupSettingInput{
					JoinPolicy: &enums.JoinPolicyOpen,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				Setting: &openlaneclient.UpdateGroup_UpdateGroup_Group_Setting{
					JoinPolicy: enums.JoinPolicyOpen,
				},
			},
		},
		{
			name: "no access",
			updateInput: openlaneclient.UpdateGroupInput{
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
			resp, err := tc.client.UpdateGroup(tc.ctx, group.ID, tc.updateInput)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

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
			assert.Check(t, is.Equal(group.DisplayID, updatedGroup.DisplayID))
			if tc.updateInput.LogoURL != nil {
				assert.Check(t, is.Equal(*tc.expectedRes.LogoURL, *updatedGroup.LogoURL))
			}

			if tc.updateInput.AddGroupMembers != nil {
				// Adding a member to an group will make it 2 users, there is an admin
				// assigned to the group automatically and a member added in the test case
				assert.Check(t, is.Len(updatedGroup.Members, 3))
				assert.Check(t, is.Equal(tc.expectedRes.Members.Edges[0].Node.Role, updatedGroup.Members.Edges[2].Node.Role))
				assert.Check(t, is.Equal(tc.expectedRes.Members.Edges[0].Node.User.ID, updatedGroup.Members.Edges[2].Node.User.ID))
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
				assert.Check(t, is.Equal(len(tc.expectedRes.Permissions), len(updatedGroup.Permissions)))

				for _, permission := range updatedGroup.Permissions {
					found := false
					for _, expectedPermission := range tc.expectedRes.Permissions {
						if *permission.ID == *expectedPermission.ID {
							found = true
							assert.Check(t, is.DeepEqual(permission, expectedPermission))
						}
					}

					assert.Check(t, found, "permission %s not found", permission.ObjectType)
				}

				// ensure user can now get access to the program
				programResp, err := suite.client.api.GetProgramByID(gmCtx, program.ID)
				assert.NilError(t, err)
				assert.Assert(t, programResp != nil)

				// ensure user can now access the control (they have editor access and should be able to make changes)
				description := gofakeit.HipsterSentence(10)
				controlResp, err := suite.client.api.UpdateControl(gmCtx, control.ID, openlaneclient.UpdateControlInput{
					Description: &description,
				})
				assert.NilError(t, err)
				assert.Assert(t, controlResp != nil)
				assert.Check(t, is.Equal(description, *controlResp.UpdateControl.Control.Description))

				// access to procedures is granted by default in the org, it should be blocked now
				procedureResp, err := suite.client.api.GetProcedureByID(gmCtx, procedure.ID)

				assert.Assert(t, is.Nil(procedureResp))
			}

			if tc.updateInput.InheritGroupPermissions != nil {
				// ensure the group has the additional permissions as the group we cloned, there is one overlap with the group we cloned
				assert.Check(t, is.Len(updatedGroup.Permissions, 5))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{om.ID, gm.Edges.Orgmembership.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{group.ID, groupClone.ID}}).MustDelete(testUser1.UserCtx, t)
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

	_, err = suite.client.api.UpdateGroupSetting(testUser1.UserCtx, privateGroupWithSetting.Group.Setting.ID, openlaneclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name     string
		groupID  string
		client   *openlaneclient.OpenlaneClient
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
				assert.Check(t, is.Nil(resp))

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
	whereInput := &openlaneclient.GroupWhereInput{
		IsManaged: lo.ToPtr(true),
	}

	resp, err := suite.client.api.GetGroupInfo(testUser1.UserCtx, whereInput)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// there should be 3 managed groups created by the system on org creation
	assert.Check(t, is.Len(resp.Groups.Edges, 3))

	// you should not be able to update a managed group
	groupID := resp.Groups.Edges[0].Node.ID
	input := openlaneclient.UpdateGroupInput{
		Tags: []string{"test"},
	}

	_, err = suite.client.api.UpdateGroup(testUser1.UserCtx, groupID, input)
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should not be able to add group members to a managed group
	_, err = suite.client.api.AddUserToGroupWithRole(testUser1.UserCtx, openlaneclient.CreateGroupMembershipInput{
		GroupID: groupID,
		UserID:  testUser2.ID,
	})
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should not be able to delete a managed group
	_, err = suite.client.api.DeleteGroup(testUser1.UserCtx, groupID)
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should, however, be able to update permissions edges on a managed group
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	input = openlaneclient.UpdateGroupInput{
		AddProgramEditorIDs:              []string{program.ID},
		AddControlViewerIDs:              []string{control.ID},
		AddInternalPolicyBlockedGroupIDs: []string{policy.ID},
	}

	updateResp, err := suite.client.api.UpdateGroup(testUser1.UserCtx, groupID, input)
	assert.NilError(t, err)

	perms := updateResp.UpdateGroup.Group.GetPermissions()
	assert.Check(t, is.Len(perms, 3))

	// make sure I can also remove them
	input = openlaneclient.UpdateGroupInput{
		RemoveProgramEditorIDs:              []string{program.ID},
		RemoveControlViewerIDs:              []string{control.ID},
		RemoveInternalPolicyBlockedGroupIDs: []string{policy.ID},
	}

	updateResp, err = suite.client.api.UpdateGroup(testUser1.UserCtx, groupID, input)
	assert.NilError(t, err)

	perms = updateResp.UpdateGroup.Group.GetPermissions()
	assert.Check(t, is.Len(perms, 0))

	// cleanup objects created
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: policy.ID}).MustDelete(testUser1.UserCtx, t)
}
