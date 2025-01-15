package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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

		// make sure two organizations are returned (group 2 and group 3) and the seeded group
		assert.Equal(t, 3, len(resp.Groups.Edges))

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

		// make sure only two groups are returned, group 1 and the seeded group
		assert.Equal(t, 2, len(resp.Groups.Edges))
	})
}

func (suite *GraphTestSuite) TestMutationCreateGroup() {
	t := suite.T()

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
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
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
						AddGroupCreatorIDs: []string{viewOnlyUser.GroupID},
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

func (suite *GraphTestSuite) TestMutationUpdateGroup() {
	t := suite.T()

	nameUpdate := gofakeit.Name()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence(10)
	gravatarURLUpdate := gofakeit.URL()

	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	om := (&OrgMemberBuilder{client: suite.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		updateInput openlaneclient.UpdateGroupInput
		expectedRes openlaneclient.UpdateGroup_UpdateGroup_Group
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		errorMsg    string
	}{
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
				Setting: openlaneclient.UpdateGroup_UpdateGroup_Group_Setting{
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
				// assigned to the group automatically
				assert.Len(t, updatedGroup.Members, 2)
				assert.Equal(t, tc.expectedRes.Members[0].Role, updatedGroup.Members[1].Role)
				assert.Equal(t, tc.expectedRes.Members[0].User.ID, updatedGroup.Members[1].User.ID)
			}

			if tc.updateInput.UpdateGroupSettings != nil {
				assert.Equal(t, updatedGroup.GetSetting().JoinPolicy, enums.JoinPolicyOpen)
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
