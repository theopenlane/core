package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryGroup() {
	t := suite.T()

	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Group)
		})
	}
}

func (suite *GraphTestSuite) TestQueryGroupsByOwner() {
	t := suite.T()

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	reqCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, org1.ID)
	require.NoError(t, err)

	group1 := (&GroupBuilder{client: suite.client, Owner: org1.ID}).MustNew(reqCtx, t)

	reqCtx2, err := auth.NewTestContextWithOrgID(testUser1.ID, org2.ID)
	require.NoError(t, err)

	group2 := (&GroupBuilder{client: suite.client, Owner: org2.ID}).MustNew(reqCtx2, t)

	t.Run("Get Groups By Owner", func(t *testing.T) {
		whereInput := &openlaneclient.GroupWhereInput{
			HasOwnerWith: []*openlaneclient.OrganizationWhereInput{
				{
					ID: &org1.ID,
				},
			},
			IsManaged: lo.ToPtr(false),
		}

		resp, err := suite.client.api.GetGroups(reqCtx, whereInput)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Groups.Edges)

		// make sure 1 group is returned
		assert.Equal(t, 1, len(resp.Groups.Edges))

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
		assert.True(t, group1Found)
		assert.False(t, group2Found)

		whereInput = &openlaneclient.GroupWhereInput{
			HasOwnerWith: []*openlaneclient.OrganizationWhereInput{
				{
					ID: &org2.ID,
				},
			},
			IsManaged: lo.ToPtr(false),
		}

		resp, err = suite.client.api.GetGroups(reqCtx2, whereInput)

		require.NoError(t, err)
		require.Len(t, resp.Groups.Edges, 1)
	})

	// delete created orgs
	(&OrganizationCleanup{client: suite.client, ID: org1.ID}).MustDelete(reqCtx, t)
	(&OrganizationCleanup{client: suite.client, ID: org2.ID}).MustDelete(reqCtx2, t)
}

func (suite *GraphTestSuite) TestQueryGroups() {
	t := suite.T()

	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group2 := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	group3 := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	t.Run("Get Groups", func(t *testing.T) {
		resp, err := suite.client.api.GetAllGroups(testUser2.UserCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Groups.Edges)

		// make sure two organizations are returned (group 2 and group 3), the seeded group, and the 3 managed groups
		assert.Equal(t, 6, len(resp.Groups.Edges))

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
		assert.True(t, group2Found)
		assert.True(t, group3Found)

		// if group 1 (which belongs to an unauthorized org) is found, fail the test
		require.False(t, group1Found)

		resp, err = suite.client.api.GetAllGroups(testUser1.UserCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)

		// make sure only two groups are returned, group 1 and the seeded group, and the 3 managed groups
		assert.Equal(t, 5, len(resp.Groups.Edges))
	})
}

func (suite *GraphTestSuite) TestMutationCreateGroup() {
	t := suite.T()

	name := gofakeit.Name()

	// group for the view only user
	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

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
			owner:       testUser1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
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
						AddGroupCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				require.NoError(t, err)
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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateGroup.Group)

			// Make sure provided values match
			assert.Equal(t, tc.groupName, resp.CreateGroup.Group.Name)
			assert.Equal(t, tc.description, *resp.CreateGroup.Group.Description)

			if tc.displayName != "" {
				assert.Equal(t, tc.displayName, resp.CreateGroup.Group.DisplayName)
			} else {
				// display name defaults to the name if not set
				assert.Equal(t, tc.groupName, resp.CreateGroup.Group.DisplayName)
			}

			if tc.settings != nil {
				assert.Equal(t, resp.CreateGroup.Group.Setting.JoinPolicy, enums.JoinPolicyInviteOnly)
			}

			if tc.owner != "" && tc.ctx == testUser2.UserCtx {
				// make sure the owner is ignored if the user doesn't have access
				assert.NotEqual(t, tc.owner, resp.CreateGroup.Group.Owner.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateGroupWithMembers() {
	t := suite.T()

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateGroupWithMembers.Group)

			// Make sure provided values match
			assert.Equal(t, tc.group.Name, resp.CreateGroupWithMembers.Group.Name)

			// ensure we can still set the visibility on the group when creating it
			assert.Equal(t, resp.CreateGroupWithMembers.Group.Setting.Visibility, enums.VisibilityPrivate)

			// make sure there are three members, user who created the group, admin, and member
			// except when using an api token
			expectedLen := 3
			if tc.client == suite.client.apiWithToken {
				expectedLen = 2
			}

			require.Len(t, resp.CreateGroupWithMembers.Group.Members, expectedLen)

			// make sure we get the member data back
			for _, member := range tc.members {
				found := false
				for _, m := range resp.CreateGroupWithMembers.Group.Members {
					require.NotNil(t, m.User)

					if m.User.ID == member.UserID {
						found = true
						assert.Equal(t, *member.Role, m.Role)

						assert.NotEmpty(t, m.User.FirstName)
						assert.NotEmpty(t, m.User.LastName)
					}
				}

				assert.Truef(t, found, "member %s not found", member.UserID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateGroupByClone() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group := (&GroupBuilder{client: suite.client, ProgramEditorsIDs: []string{program.ID}, ControlEditorsIDs: []string{control.ID}}).MustNew(testUser1.UserCtx, t)

	// add a group member to the group
	(&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                  string
		group                 openlaneclient.CreateGroupInput
		groupPermissionsClone string
		groupMembersClone     string
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
			groupPermissionsClone: group.ID,
			groupMembersClone:     group.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
		},
		{
			name: "happy path, clone group members, use personal access token",
			group: openlaneclient.CreateGroupInput{
				Name:    gofakeit.Name(),
				OwnerID: &testUser1.OrganizationID,
			},
			groupMembersClone: group.ID,
			client:            suite.client.apiWithPAT,
			ctx:               context.Background(),
		},
		// {
		// 	name: "happy path, clone group permissions, use api token",
		// 	group: openlaneclient.CreateGroupInput{
		// 		Name: gofakeit.Name(),
		// 	},
		// 	groupPermissionsClone: group.ID,
		// 	client:                suite.client.apiWithToken,
		// 	ctx:                   context.Background(),
		// },
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateGroupByClone(tc.ctx, tc.group, &tc.groupPermissionsClone, &tc.groupMembersClone)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateGroupByClone.Group)

			// the display id should be different
			require.NotEqual(t, group.DisplayID, resp.CreateGroupByClone.Group.DisplayID)

			// make sure there are two members, user who created the group and the cloned member
			// even when an api token is used, there will still be the original user (testUser1)
			expectedLen := 1
			if tc.groupMembersClone != "" {
				expectedLen += 1
			}

			assert.Len(t, resp.CreateGroupByClone.Group.Members, expectedLen)

			// added a control and a program to the group we cloned, make sure they are there
			expectedLenPerms := 0
			if tc.groupPermissionsClone != "" {
				expectedLenPerms = 2
			}

			assert.Len(t, resp.CreateGroupByClone.Group.Permissions, expectedLenPerms)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateGroup() {
	t := suite.T()

	nameUpdate := gofakeit.Name()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence(10)
	gravatarURLUpdate := gofakeit.URL()

	group := (&GroupBuilder{client: suite.client, Owner: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	gm := (&GroupMemberBuilder{client: suite.client, GroupID: group.ID}).MustNew(testUser1.UserCtx, t)

	om := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	gmCtx, err := auth.NewTestContextWithOrgID(gm.UserID, testUser1.OrganizationID)
	require.NoError(t, err)

	// ensure user cannot get access to the program
	programResp, err := suite.client.api.GetProgramByID(gmCtx, program.ID)
	require.Error(t, err)
	require.Nil(t, programResp)

	// ensure user cannot get access to the control
	controlResp, err := suite.client.api.GetControlByID(gmCtx, control.ID)
	require.Error(t, err)
	require.Nil(t, controlResp)

	// access to procedures is granted by default in the org
	procedureResp, err := suite.client.api.GetProcedureByID(gmCtx, procedure.ID)
	require.NoError(t, err)
	require.NotNil(t, procedureResp)

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
						Name:        &control.Name,
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
			name: "update name, happy path",
			updateInput: openlaneclient.UpdateGroupInput{
				Name:        &nameUpdate,
				DisplayName: &displayNameUpdate,
				Description: &descriptionUpdate,
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
				Members: []*openlaneclient.UpdateGroup_UpdateGroup_Group_Members{
					{
						Role: enums.RoleAdmin,
						User: openlaneclient.UpdateGroup_UpdateGroup_Group_Members_User{
							ID: om.UserID,
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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateGroup.Group)

			// Make sure provided values match
			updatedGroup := resp.GetUpdateGroup().Group
			assert.Equal(t, tc.expectedRes.Name, updatedGroup.Name)
			assert.Equal(t, tc.expectedRes.DisplayName, updatedGroup.DisplayName)
			assert.Equal(t, tc.expectedRes.Description, updatedGroup.Description)

			if tc.updateInput.GravatarLogoURL != nil {
				assert.Equal(t, *tc.expectedRes.LogoURL, *updatedGroup.LogoURL)
			}

			if tc.updateInput.AddGroupMembers != nil {
				// Adding a member to an group will make it 2 users, there is an admin
				// assigned to the group automatically and a member added in the test case
				assert.Len(t, updatedGroup.Members, 3)
				assert.Equal(t, tc.expectedRes.Members[0].Role, updatedGroup.Members[2].Role)
				assert.Equal(t, tc.expectedRes.Members[0].User.ID, updatedGroup.Members[2].User.ID)
			}

			if tc.updateInput.UpdateGroupSettings != nil {
				assert.Equal(t, updatedGroup.GetSetting().JoinPolicy, enums.JoinPolicyOpen)
			}

			if tc.updateInput.AddProgramViewerIDs != nil || tc.updateInput.AddProcedureEditorIDs != nil || tc.updateInput.AddControlBlockedGroupIDs != nil {
				assert.Equal(t, len(tc.expectedRes.Permissions), len(updatedGroup.Permissions))

				// ensure user can now get access to the program
				programResp, err := suite.client.api.GetProgramByID(gmCtx, program.ID)
				require.NoError(t, err)
				require.NotNil(t, programResp)

				// ensure user can now access the control (they have editor access and should be able to make changes)
				description := gofakeit.HipsterSentence(10)
				controlResp, err := suite.client.api.UpdateControl(gmCtx, control.ID, openlaneclient.UpdateControlInput{
					Description: &description,
				})
				require.NoError(t, err)
				require.NotNil(t, controlResp)
				assert.Equal(t, description, *controlResp.UpdateControl.Control.Description)

				// access to procedures is granted by default in the org
				procedureResp, err := suite.client.api.GetProcedureByID(gmCtx, procedure.ID)
				require.Error(t, err)
				require.Nil(t, procedureResp)

			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteGroup() {
	t := suite.T()

	group1 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group2 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group3 := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		groupID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "delete group, happy path",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			groupID: group1.ID,
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
			errorMsg: notFoundErrorMsg, // user was not added to the group, so they can't delete it
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteGroup(tc.ctx, tc.groupID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.DeleteGroup.DeletedID)

			// make sure the deletedID matches the ID we wanted to delete
			assert.Equal(t, tc.groupID, resp.DeleteGroup.DeletedID)
		})
	}
}

func (suite *GraphTestSuite) TestManagedGroups() {
	t := suite.T()
	whereInput := &openlaneclient.GroupWhereInput{
		IsManaged: lo.ToPtr(true),
	}

	resp, err := suite.client.api.GetGroups(testUser1.UserCtx, whereInput)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// there should be 3 managed groups created by the system on org creation
	assert.Len(t, resp.Groups.Edges, 3)

	// you should not be able to update a managed group
	groupID := resp.Groups.Edges[0].Node.ID
	input := openlaneclient.UpdateGroupInput{
		Tags: []string{"test"},
	}

	_, err = suite.client.api.UpdateGroup(testUser1.UserCtx, groupID, input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should not be able to add group members to a managed group
	_, err = suite.client.api.AddUserToGroupWithRole(testUser1.UserCtx, openlaneclient.CreateGroupMembershipInput{
		GroupID: groupID,
		UserID:  testUser2.ID,
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "managed groups cannot be modified")

	// you should not be able to delete a managed group
	_, err = suite.client.api.DeleteGroup(testUser1.UserCtx, groupID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "managed groups cannot be modified")
}
