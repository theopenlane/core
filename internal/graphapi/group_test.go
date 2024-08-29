package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
)

func (suite *GraphTestSuite) TestQueryGroup() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	group1 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	listGroups := []string{fmt.Sprintf("group:%s", group1.ID)}

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		allowed  bool
		expected *ent.Group
		errorMsg string
	}{
		{
			name:     "happy path group",
			client:   suite.client.api,
			ctx:      reqCtx,
			allowed:  true,
			queryID:  group1.ID,
			expected: group1,
		},
		{
			name:     "happy path group, using api token",
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			allowed:  true,
			queryID:  group1.ID,
			expected: group1,
		},
		{
			name:     "happy path group, using personal access token",
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			allowed:  true,
			queryID:  group1.ID,
			expected: group1,
		},
		{
			name:     "no access",
			client:   suite.client.api,
			ctx:      reqCtx,
			allowed:  false,
			queryID:  group1.ID,
			errorMsg: "not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			// second check won't happen if org does not exist
			if tc.errorMsg == "" {
				mock_fga.ListTimes(t, suite.client.fga, listGroups, 1)
			}

			resp, err := suite.client.api.GetGroupByID(reqCtx, tc.queryID)

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

	// delete created org and group
	(&GroupCleanup{client: suite.client, ID: group1.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestQueryGroupsByOwner() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, org1.ID)
	require.NoError(t, err)

	group1 := (&GroupBuilder{client: suite.client, Owner: org1.ID}).MustNew(reqCtx, t)
	group2 := (&GroupBuilder{client: suite.client, Owner: org2.ID}).MustNew(reqCtx, t)

	t.Run("Get Groups By Owner", func(t *testing.T) {
		defer mock_fga.ClearMocks(suite.client.fga)

		// check tuple per org
		listGroups := []string{fmt.Sprintf("group:%s", group1.ID)}

		mock_fga.ListAny(t, suite.client.fga, listGroups)

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
		if !group1Found || group2Found {
			t.Fail()
		}

		whereInput = &openlaneclient.GroupWhereInput{
			HasOwnerWith: []*openlaneclient.OrganizationWhereInput{
				{
					ID: &org2.ID,
				},
			},
		}

		resp, err = suite.client.api.GetGroups(reqCtx, whereInput)

		require.NoError(t, err)
		require.Empty(t, resp.Groups.Edges)
	})

	// delete created orgs and groups
	reqCtx2, err := auth.NewTestContextWithOrgID(testUser.ID, org2.ID)
	require.NoError(t, err)

	(&OrganizationCleanup{client: suite.client, ID: org1.ID}).MustDelete(reqCtx, t)
	(&OrganizationCleanup{client: suite.client, ID: org2.ID}).MustDelete(reqCtx2, t)
}

func (suite *GraphTestSuite) TestQueryGroups() {
	t := suite.T()

	// setup user context
	reqCtx2, err := userContext()
	require.NoError(t, err)

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx2, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx2, t)

	reqCtx1, err := auth.NewTestContextWithOrgID(testUser.ID, org1.ID)
	require.NoError(t, err)

	reqCtx2, err = auth.NewTestContextWithOrgID(testUser.ID, org2.ID)
	require.NoError(t, err)

	group1 := (&GroupBuilder{client: suite.client, Owner: org1.ID}).MustNew(reqCtx1, t)
	group2 := (&GroupBuilder{client: suite.client, Owner: org2.ID}).MustNew(reqCtx2, t)
	group3 := (&GroupBuilder{client: suite.client, Owner: org2.ID}).MustNew(reqCtx2, t)

	t.Run("Get Groups", func(t *testing.T) {
		defer mock_fga.ClearMocks(suite.client.fga)

		// check org tuples
		listGroups := []string{fmt.Sprintf("group:%s", group2.ID), fmt.Sprintf("group:%s", group3.ID)}

		mock_fga.ListTimes(t, suite.client.fga, listGroups, 1)

		resp, err := suite.client.api.GetAllGroups(reqCtx2)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Groups.Edges)

		// make sure two organizations are returned (group 2 and group 3)
		assert.Equal(t, 2, len(resp.Groups.Edges))

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
		if !group2Found || !group3Found {
			t.Fail()
		}

		// if group 1 (which belongs to an unauthorized org) is found, fail the test
		if group1Found {
			t.Fail()
		}

		// Check user with no relations, gets no groups back
		mock_fga.ListAny(t, suite.client.fga, []string{})

		resp, err = suite.client.api.GetAllGroups(reqCtx2)

		require.NoError(t, err)
		require.NotNil(t, resp)

		// make sure no organizations are returned
		assert.Equal(t, 0, len(resp.Groups.Edges))
	})

	// delete created orgs and groups
	(&GroupCleanup{client: suite.client, ID: group1.ID}).MustDelete(reqCtx1, t)
	(&GroupCleanup{client: suite.client, ID: group2.ID}).MustDelete(reqCtx2, t)
	(&GroupCleanup{client: suite.client, ID: group3.ID}).MustDelete(reqCtx2, t)
	(&OrganizationCleanup{client: suite.client, ID: org1.ID}).MustDelete(reqCtx1, t)
	(&OrganizationCleanup{client: suite.client, ID: org2.ID}).MustDelete(reqCtx2, t)
}

func (suite *GraphTestSuite) TestMutationCreateGroup() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	otherOwner := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	listObjects := []string{fmt.Sprintf("organization:%s", testOrgID), fmt.Sprintf("organization:%s", testPersonalOrgID)}

	testCases := []struct {
		name        string
		groupName   string
		description string
		displayName string
		owner       string
		settings    *openlaneclient.CreateGroupSettingInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		list        bool
		check       bool
		errorMsg    string
	}{
		{
			name:        "happy path group",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			list:        true,
			check:       true,
		},
		{
			name:        "happy path group using api token",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			allowed:     true,
			// TODO (sfunk): look at the authz logic, this one looks slightly different from other objects
			// but should be the same
			list:  false, // no list objects because the api token can only be associated with a single org
			check: true,
		},
		{
			name:        "happy path group using personal access token",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			owner:       testOrgID,
			description: gofakeit.HipsterSentence(10),
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			allowed:     true,
			list:        true,
			check:       true,
		},
		{
			name:        "happy path group with settings",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			settings: &openlaneclient.CreateGroupSettingInput{
				JoinPolicy: &enums.JoinPolicyInviteOnly,
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			list:    true,
			check:   true,
		},
		{
			name:        "no access to owner",
			groupName:   gofakeit.Name(),
			displayName: gofakeit.LetterN(50),
			description: gofakeit.HipsterSentence(10),
			owner:       otherOwner.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			check:       true,
			list:        true,
			errorMsg:    "not authorized",
		},
		{
			name:      "happy path group, minimum fields",
			groupName: gofakeit.Name(),
			client:    suite.client.api,
			ctx:       reqCtx,
			allowed:   true,
			list:      true,
			check:     true,
		},
		{
			name:     "missing name",
			errorMsg: "validator failed",
			client:   suite.client.api,
			ctx:      reqCtx,
			allowed:  true,
			list:     true,
			check:    true,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// clear mocks at end of each test
			defer mock_fga.ClearMocks(suite.client.fga)

			tc := tc
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

			if tc.check {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			// When calls are expected to fail, we won't ever write tuples
			if tc.errorMsg == "" {
				mock_fga.WriteAny(t, suite.client.fga)

				if tc.list {
					mock_fga.ListAny(t, suite.client.fga, listObjects)
				}
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
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateGroup() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	nameUpdate := gofakeit.Name()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence(10)
	gravatarURLUpdate := gofakeit.URL()

	group := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	om := (&OrgMemberBuilder{client: suite.client, OrgID: testOrgID}).MustNew(reqCtx, t)

	// setup auth for the tests
	listObjects := []string{fmt.Sprintf("group:%s", group.ID)}

	testCases := []struct {
		name        string
		allowed     bool
		updateInput openlaneclient.UpdateGroupInput
		expectedRes openlaneclient.UpdateGroup_UpdateGroup_Group
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		list        bool
		errorMsg    string
	}{
		{
			name:    "update name, happy path",
			allowed: true,
			updateInput: openlaneclient.UpdateGroupInput{
				Name:        &nameUpdate,
				DisplayName: &displayNameUpdate,
				Description: &descriptionUpdate,
			},
			client: suite.client.api,
			ctx:    reqCtx,
			list:   true,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
			},
		},
		{
			name:    "add user as admin using api token",
			allowed: true,
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
			list:   true,
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
			name:    "update gravatar, happy path using personal access token",
			allowed: true,
			updateInput: openlaneclient.UpdateGroupInput{
				LogoURL: &gravatarURLUpdate,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
			list:   true,
			expectedRes: openlaneclient.UpdateGroup_UpdateGroup_Group{
				ID:          group.ID,
				Name:        nameUpdate,
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
				LogoURL:     &gravatarURLUpdate,
			},
		},
		{
			name:    "update settings, happy path",
			allowed: true,
			updateInput: openlaneclient.UpdateGroupInput{
				UpdateGroupSettings: &openlaneclient.UpdateGroupSettingInput{
					JoinPolicy: &enums.JoinPolicyOpen,
				},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			list:   true,
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
			name:    "no access",
			allowed: false,
			updateInput: openlaneclient.UpdateGroupInput{
				Name:        &nameUpdate,
				DisplayName: &displayNameUpdate,
				Description: &descriptionUpdate,
			},
			client:   suite.client.api,
			ctx:      reqCtx,
			list:     false,
			errorMsg: "not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			if tc.list {
				mock_fga.ListAny(t, suite.client.fga, listObjects)
			}

			if tc.updateInput.AddGroupMembers != nil && tc.errorMsg == "" {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			// update group
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

	(&GroupCleanup{client: suite.client, ID: group.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationDeleteGroup() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	group1 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)
	group2 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)
	group3 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	listObjects := []string{
		fmt.Sprintf("group:%s", group1.ID),
		fmt.Sprintf("group:%s", group2.ID),
		fmt.Sprintf("group:%s", group3.ID),
	}

	testCases := []struct {
		name     string
		groupID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		allowed  bool
		errorMsg string
	}{
		{
			name:    "delete group, happy path",
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
			groupID: group1.ID,
		},
		{
			name:    "delete group, happy path using api token",
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
			groupID: group2.ID,
		},
		{
			name:    "delete group, happy path using personal access token",
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
			groupID: group3.ID,
		},
		{
			name:     "delete group, no access",
			client:   suite.client.api,
			ctx:      reqCtx,
			allowed:  false,
			groupID:  group1.ID,
			errorMsg: "not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// mock read of tuple
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			if tc.allowed {
				mock_fga.ReadAny(t, suite.client.fga)
				mock_fga.ListAny(t, suite.client.fga, listObjects)
				mock_fga.WriteAny(t, suite.client.fga)
			}

			// delete group
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

			o, err := suite.client.api.GetGroupByID(reqCtx, tc.groupID)

			require.Nil(t, o)
			require.Error(t, err)
			assert.ErrorContains(t, err, "not found")
		})
	}
}
