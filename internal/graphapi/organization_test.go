package graphapi_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/entx"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryOrganization() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	orgMember := (&OrgMemberBuilder{client: suite.client, OrgID: org1.ID}).MustNew(reqCtx, t)

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, org1.ID)
	require.NoError(t, err)

	testCases := []struct {
		name               string
		queryID            string
		client             *openlaneclient.OpenlaneClient
		ctx                context.Context
		expected           *ent.Organization
		orgMembersExpected int
		errorMsg           string
	}{
		{
			name:               "happy path, get organization",
			queryID:            org1.ID,
			client:             suite.client.api,
			ctx:                reqCtx,
			orgMembersExpected: 2,
			expected:           org1,
		},
		{
			name:               "happy path, get using api token",
			queryID:            testOrgID,
			client:             suite.client.apiWithToken,
			ctx:                context.Background(),
			orgMembersExpected: 1,
			expected:           org1,
		},
		{
			name:               "happy path, get using personal access token",
			queryID:            testOrgID,
			client:             suite.client.apiWithPAT,
			ctx:                context.Background(),
			orgMembersExpected: 1,
			expected:           org1,
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "organization not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, true)

			if tc.client != suite.client.apiWithToken {
				mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID, "organization:" + org1.ID})
			}

			if tc.errorMsg == "" {
				mock_fga.ListUsersAny(t, suite.client.fga, []string{testUser.ID, orgMember.UserID}, nil)
			}

			resp, err := tc.client.GetOrganizationByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Organization)
			require.NotNil(t, resp.Organization.Members)
			assert.Len(t, resp.Organization.Members, tc.orgMembersExpected)

			if tc.orgMembersExpected > 1 {
				orgMemberFound := false

				for _, m := range resp.Organization.Members {
					if m.User.ID == orgMember.UserID {
						orgMemberFound = true
					}
				}

				assert.True(t, orgMemberFound)
			}
		})
	}

	// delete created org
	(&OrganizationCleanup{client: suite.client, ID: org1.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestQueryOrganizations() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	t.Run("Get Organizations", func(t *testing.T) {
		defer mock_fga.ClearMocks(suite.client.fga)

		mock_fga.CheckAny(t, suite.client.fga, true)
		mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID,
			"organization:" + testPersonalOrgID,
			"organization:" + org1.ID,
			"organization:" + org2.ID,
		})

		resp, err := suite.client.api.GetAllOrganizations(reqCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Organizations.Edges)

		// make sure two organizations are returned, the two created and
		// the personal org and test org created at suite setup
		assert.Equal(t, 4, len(resp.Organizations.Edges))

		org1Found := false
		org2Found := false

		for _, o := range resp.Organizations.Edges {
			if o.Node.ID == org1.ID {
				org1Found = true
			} else if o.Node.ID == org2.ID {
				org2Found = true
			}
		}

		// if one of the orgs isn't found, fail the test
		if !org1Found || !org2Found {
			t.Fail()
		}
	})
}

func (suite *GraphTestSuite) TestMutationCreateOrganization() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	mock_fga.CheckAny(t, suite.client.fga, true)
	mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID})
	mock_fga.ListUsersAny(t, suite.client.fga, []string{testUser.ID}, nil)

	parentOrg, err := suite.client.api.GetOrganizationByID(reqCtx, testOrgID)
	require.NoError(t, err)

	mock_fga.ClearMocks(suite.client.fga)

	// setup deleted org
	orgToDelete := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	// delete said org
	(&OrganizationCleanup{client: suite.client, ID: orgToDelete.ID}).MustDelete(reqCtx, t)

	testCases := []struct {
		name                     string
		orgName                  string
		displayName              string
		orgDescription           string
		parentOrgID              string
		settings                 *openlaneclient.CreateOrganizationSettingInput
		listOrgs                 bool
		client                   *openlaneclient.OpenlaneClient
		ctx                      context.Context
		expectedDefaultOrgUpdate bool
		errorMsg                 string
	}{
		{
			name:                     "happy path organization",
			orgName:                  gofakeit.Name(),
			displayName:              gofakeit.LetterN(50),
			orgDescription:           gofakeit.HipsterSentence(10),
			listOrgs:                 true,
			expectedDefaultOrgUpdate: true, // only the first org created should update the default org
			parentOrgID:              "",   // root org
			client:                   suite.client.api,
			ctx:                      reqCtx,
		},
		{
			name:           "happy path organization with settings",
			orgName:        gofakeit.Name(),
			displayName:    gofakeit.LetterN(50),
			orgDescription: gofakeit.HipsterSentence(10),
			listOrgs:       true,
			settings: &openlaneclient.CreateOrganizationSettingInput{
				Domains: []string{"meow.theopenlane.io"},
			},
			parentOrgID: "", // root org
			client:      suite.client.api,
			ctx:         reqCtx,
		},
		{
			name:           "happy path organization with parent org",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(10),
			listOrgs:       true,
			parentOrgID:    testOrgID,
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "happy path organization with parent org using personal access token",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(10),
			listOrgs:       true,
			parentOrgID:    testOrgID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
		},
		{
			name:           "organization with parent personal org",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(10),
			listOrgs:       true,
			parentOrgID:    testPersonalOrgID,
			errorMsg:       "personal organizations are not allowed to have child organizations",
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "empty organization name",
			orgName:        "",
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "value is less than the required length",
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "long organization name",
			orgName:        gofakeit.LetterN(161),
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "value is greater than the required length",
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "organization with no description",
			orgName:        gofakeit.Name(),
			orgDescription: "",
			listOrgs:       true,
			parentOrgID:    testOrgID,
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "duplicate organization name",
			orgName:        parentOrg.Organization.Name,
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "already exists",
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "duplicate organization name, but other was deleted, should pass",
			orgName:        orgToDelete.Name,
			orgDescription: gofakeit.HipsterSentence(10),
			listOrgs:       true,
			errorMsg:       "",
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "duplicate display name, should be allowed",
			orgName:        gofakeit.LetterN(80),
			displayName:    parentOrg.Organization.DisplayName,
			listOrgs:       true,
			orgDescription: gofakeit.HipsterSentence(10),
			client:         suite.client.api,
			ctx:            reqCtx,
		},
		{
			name:           "display name with spaces should pass",
			orgName:        gofakeit.Name(),
			displayName:    gofakeit.Sentence(3),
			orgDescription: gofakeit.HipsterSentence(10),
			listOrgs:       true,
			client:         suite.client.api,
			ctx:            reqCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			tc := tc
			input := openlaneclient.CreateOrganizationInput{
				Name:        tc.orgName,
				Description: &tc.orgDescription,
			}

			if tc.displayName != "" {
				input.DisplayName = &tc.displayName
			}

			if tc.parentOrgID != "" {
				input.ParentID = &tc.parentOrgID

				if tc.errorMsg != "" {
					mock_fga.CheckAny(t, suite.client.fga, true)
				}
			}

			if tc.settings != nil {
				input.CreateOrgSettings = tc.settings
			}

			// When calls are expected to fail, we won't ever write tuples
			if tc.errorMsg == "" {
				mock_fga.CheckAny(t, suite.client.fga, true)
				mock_fga.WriteAny(t, suite.client.fga)
			}

			if tc.listOrgs {
				mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID, "organization:" + testPersonalOrgID})
			}

			resp, err := tc.client.CreateOrganization(tc.ctx, input)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateOrganization.Organization)

			// Make sure provided values match
			assert.Equal(t, tc.orgName, resp.CreateOrganization.Organization.Name)
			assert.Equal(t, tc.orgDescription, *resp.CreateOrganization.Organization.Description)

			if tc.parentOrgID == "" {
				assert.Nil(t, resp.CreateOrganization.Organization.Parent)
			} else {
				parent := resp.CreateOrganization.Organization.GetParent()
				assert.Equal(t, tc.parentOrgID, parent.ID)
			}

			// Ensure org settings is not null
			assert.NotNil(t, resp.CreateOrganization.Organization.Setting.ID)

			// Ensure display name is not empty
			assert.NotEmpty(t, resp.CreateOrganization.Organization.DisplayName)

			if tc.settings != nil {
				assert.Len(t, resp.CreateOrganization.Organization.Setting.Domains, 1)

				// make sure default org is updated if it's the first org created
				userResp, err := tc.client.GetUserByID(tc.ctx, testUser.ID)
				require.NoError(t, err)

				if tc.expectedDefaultOrgUpdate {
					assert.Equal(t, resp.CreateOrganization.Organization.ID, userResp.User.Setting.DefaultOrg.ID)
				} else {
					assert.NotEqual(t, resp.CreateOrganization.Organization.ID, userResp.User.Setting.DefaultOrg.ID)
				}
			}

			// ensure entity types are created
			newCtx, err := auth.NewTestContextWithOrgID(testUser.ID, resp.CreateOrganization.Organization.ID)
			require.NoError(t, err)

			et, err := suite.client.api.GetEntityTypes(newCtx, &openlaneclient.EntityTypeWhereInput{
				OwnerID: &resp.CreateOrganization.Organization.ID,
			})
			require.NoError(t, err)

			require.Len(t, et.EntityTypes.Edges, 1)
			assert.Equal(t, "vendor", et.EntityTypes.Edges[0].Node.Name)
			assert.Equal(t, resp.CreateOrganization.Organization.ID, *et.EntityTypes.Edges[0].Node.OwnerID)

			// cleanup org
			(&OrganizationCleanup{client: suite.client, ID: resp.CreateOrganization.Organization.ID}).MustDelete(reqCtx, t)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateOrganization() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	nameUpdate := gofakeit.Name()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence(10)
	nameUpdateLong := gofakeit.LetterN(200)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	testUser1 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		updateInput openlaneclient.UpdateOrganizationInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedRes openlaneclient.UpdateOrganization_UpdateOrganization_Organization
		errorMsg    string
	}{
		{
			name: "update name, happy path",
			updateInput: openlaneclient.UpdateOrganizationInput{
				Name: &nameUpdate,
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
			},
		},
		{
			name: "add member as admin",
			updateInput: openlaneclient.UpdateOrganizationInput{
				AddOrgMembers: []*openlaneclient.CreateOrgMembershipInput{
					{
						UserID: testUser1.ID,
						Role:   &enums.RoleAdmin,
					},
				},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
				Members: []*openlaneclient.UpdateOrganization_UpdateOrganization_Organization_Members{
					{
						Role:   enums.RoleAdmin,
						UserID: testUser1.ID,
					},
				},
			},
		},
		{
			name: "update description, happy path",
			updateInput: openlaneclient.UpdateOrganizationInput{
				Description: &descriptionUpdate,
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate, // this would have been updated on the prior test
				DisplayName: org.DisplayName,
				Description: &descriptionUpdate,
			},
		},
		{
			name: "update display name, happy path",
			updateInput: openlaneclient.UpdateOrganizationInput{
				DisplayName: &displayNameUpdate,
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate, // this would have been updated on the prior test
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
			},
		},
		{
			name: "update settings, happy path",
			updateInput: openlaneclient.UpdateOrganizationInput{
				Description: &descriptionUpdate,
				UpdateOrgSettings: &openlaneclient.UpdateOrganizationSettingInput{
					Domains: []string{"meow.theopenlane.io", "woof.theopenlane.io"},
				},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,        // this would have been updated on the prior test
				DisplayName: displayNameUpdate, // this would have been updated on the prior test
				Description: &descriptionUpdate,
			},
		},
		{
			name: "update name, too long",
			updateInput: openlaneclient.UpdateOrganizationInput{
				Name: &nameUpdateLong,
			},
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "value is greater than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// mock checks of tuple
			defer mock_fga.ClearMocks(suite.client.fga)
			// get and update organization
			mock_fga.CheckAny(t, suite.client.fga, true)

			if tc.updateInput.AddOrgMembers != nil {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + org.ID})

			// update org
			resp, err := tc.client.UpdateOrganization(tc.ctx, org.ID, tc.updateInput)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateOrganization.Organization)

			// Make sure provided values match
			updatedOrg := resp.GetUpdateOrganization().Organization
			assert.Equal(t, tc.expectedRes.Name, updatedOrg.Name)
			assert.Equal(t, tc.expectedRes.DisplayName, updatedOrg.DisplayName)
			assert.Equal(t, tc.expectedRes.Description, updatedOrg.Description)

			if tc.updateInput.AddOrgMembers != nil {
				// Adding a member to an org will make it 2 users, there is an owner
				// assigned to the org automatically
				assert.Len(t, updatedOrg.Members, 2)
				assert.Equal(t, tc.expectedRes.Members[0].Role, updatedOrg.Members[1].Role)
				assert.Equal(t, tc.expectedRes.Members[0].UserID, updatedOrg.Members[1].UserID)
			}

			if tc.updateInput.UpdateOrgSettings != nil {
				assert.Len(t, updatedOrg.GetSetting().Domains, 2)
			}
		})
	}

	(&OrganizationCleanup{client: suite.client, ID: org.ID}).MustDelete(reqCtx, t)
	(&UserCleanup{client: suite.client, ID: testUser1.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationDeleteOrganization() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	// setup auth
	listObjects := []string{"organization:" + org.ID, "organization:" + testPersonalOrgID}

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	// update default org for user
	// setup mocks for update user setting
	mock_fga.CheckAny(t, suite.client.fga, true)
	mock_fga.ListAny(t, suite.client.fga, listObjects)
	setting, err := suite.client.api.UpdateUserSetting(reqCtx, testUser.Edges.Setting.ID,
		openlaneclient.UpdateUserSettingInput{
			DefaultOrgID: &org.ID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, org.ID, setting.UpdateUserSetting.UserSetting.DefaultOrg.ID)
	// clear mocks
	mock_fga.ClearMocks(suite.client.fga)

	testCases := []struct {
		name          string
		orgID         string
		accessAllowed bool
		errorMsg      string
	}{
		{
			name:          "delete org, access denied",
			orgID:         org.ID,
			accessAllowed: false,
			errorMsg:      "you are not authorized to perform this action",
		},
		{
			name:          "delete org, happy path",
			orgID:         org.ID,
			accessAllowed: true,
		},
		{
			name:          "delete org, personal org not allowed",
			orgID:         testPersonalOrgID,
			accessAllowed: true,
			errorMsg:      "cannot delete personal organizations",
		},
		{
			name:          "delete org, not found",
			orgID:         "tacos-tuesday",
			accessAllowed: false,
			errorMsg:      "you are not authorized to perform this action",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// mock read of tuple
			mock_fga.CheckAny(t, suite.client.fga, tc.accessAllowed)
			// org checks
			mock_fga.ListAny(t, suite.client.fga, listObjects)

			if tc.errorMsg == "" {
				mock_fga.ListUsersAny(t, suite.client.fga, []string{testUser.ID}, nil)
			}

			// additional check happens when the resource is found
			if tc.errorMsg == "" {
				mock_fga.WriteAny(t, suite.client.fga)
			}

			// delete org
			resp, err := suite.client.api.DeleteOrganization(reqCtx, tc.orgID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.DeleteOrganization.DeletedID)

			// make sure the deletedID matches the ID we wanted to delete
			assert.Equal(t, tc.orgID, resp.DeleteOrganization.DeletedID)

			// make sure the default org is reset
			settingUpdated, err := suite.client.api.GetUserSettingByID(reqCtx, testUser.Edges.Setting.ID)
			require.NoError(t, err)
			require.NotNil(t, settingUpdated.UserSetting.DefaultOrg)
			assert.NotEqual(t, org.ID, settingUpdated.UserSetting.DefaultOrg.ID)

			o, err := suite.client.api.GetOrganizationByID(reqCtx, tc.orgID)

			require.Nil(t, o)
			require.Error(t, err)
			assert.ErrorContains(t, err, "not found")

			ctx := entx.SkipSoftDelete(reqCtx)

			o, err = suite.client.api.GetOrganizationByID(ctx, tc.orgID)

			require.Equal(t, o.Organization.ID, tc.orgID)
			require.NoError(t, err)
		})
	}
}

func (suite *GraphTestSuite) TestMutationOrganizationCascadeDelete() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx, err = auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	// add child org
	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: org.ID}).MustNew(reqCtx, t)

	group1 := (&GroupBuilder{client: suite.client, Owner: org.ID}).MustNew(reqCtx, t)

	listGroups := []string{fmt.Sprintf("group:%s", group1.ID)}

	// mocks checks for all calls
	mock_fga.CheckAny(t, suite.client.fga, true)

	// mock writes to delete member of org
	mock_fga.WriteAny(t, suite.client.fga)

	mock_fga.ListTimes(t, suite.client.fga, []string{"organization:" + org.ID, "organization:" + childOrg.ID}, 4)
	mock_fga.ListAny(t, suite.client.fga, listGroups)

	// delete org
	resp, err := suite.client.api.DeleteOrganization(reqCtx, org.ID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.DeleteOrganization.DeletedID)

	// make sure the deletedID matches the ID we wanted to delete
	assert.Equal(t, org.ID, resp.DeleteOrganization.DeletedID)

	o, err := suite.client.api.GetOrganizationByID(reqCtx, org.ID)

	require.Nil(t, o)
	require.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	co, err := suite.client.api.GetOrganizationByID(reqCtx, childOrg.ID)

	require.Nil(t, co)
	require.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	g, err := suite.client.api.GetGroupByID(reqCtx, group1.ID)

	require.Nil(t, g)
	require.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	// allow after tuples have been deleted
	ctx := privacy.DecisionContext(reqCtx, privacy.Allow)

	ctx = entx.SkipSoftDelete(ctx)

	o, err = suite.client.api.GetOrganizationByID(ctx, org.ID)

	require.NoError(t, err)
	require.Equal(t, o.Organization.ID, org.ID)

	// allow after tuples have been deleted
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	g, err = suite.client.api.GetGroupByID(ctx, group1.ID)
	require.NoError(t, err)

	require.Equal(t, g.Group.ID, group1.ID)

	// allow after tuples have been deleted
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	co, err = suite.client.api.GetOrganizationByID(ctx, childOrg.ID)
	require.NoError(t, err)

	require.Equal(t, co.Organization.ID, childOrg.ID)
}

func (suite *GraphTestSuite) TestMutationCreateOrganizationTransaction() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	t.Run("Create should not write if FGA transaction fails", func(t *testing.T) {
		input := openlaneclient.CreateOrganizationInput{
			Name: gofakeit.Name(),
		}

		fgaErr := errors.New("unable to create relationship") //nolint:err113
		mock_fga.WriteError(t, suite.client.fga, fgaErr)

		resp, err := suite.client.api.CreateOrganization(reqCtx, input)

		require.Error(t, err)
		require.Empty(t, resp)

		ctx := privacy.DecisionContext(reqCtx, privacy.Allow)

		orgs, err := suite.client.api.GetAllOrganizations(ctx)
		require.NoError(t, err)

		for _, o := range orgs.Organizations.Edges {
			if o.Node.Name == input.Name {
				t.Errorf("org found that should not have been created due to FGA error")
			}
		}
	})
}
