package graphapi_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryOrganization() {
	t := suite.T()

	testCases := []struct {
		name               string
		queryID            string
		client             *openlaneclient.OpenlaneClient
		ctx                context.Context
		orgMembersExpected int
		errorMsg           string
	}{
		{
			name:               "happy path, get organization",
			queryID:            testUser1.OrganizationID,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
			orgMembersExpected: 3, // owner, admin, and view only user
		},
		{
			name:               "happy path, get using api token",
			queryID:            testUser1.OrganizationID,
			client:             suite.client.apiWithToken,
			ctx:                context.Background(),
			orgMembersExpected: 3, // owner, admin, and view only user
		},
		{
			name:               "happy path, get using personal access token",
			queryID:            testUser1.OrganizationID,
			client:             suite.client.apiWithPAT,
			ctx:                context.Background(),
			orgMembersExpected: 3, // owner, admin, and view only user
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
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
					if m.User.ID == viewOnlyUser.ID {
						orgMemberFound = true
					}
				}

				assert.True(t, orgMemberFound)
			}
		})
	}
}

func (suite *GraphTestSuite) TestQueryOrganizations() {
	t := suite.T()

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	t.Run("Get Organizations", func(t *testing.T) {
		resp, err := suite.client.api.GetAllOrganizations(testUser1.UserCtx)

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

		assert.True(t, org1Found)
		assert.True(t, org2Found)
	})
}

func (suite *GraphTestSuite) TestMutationCreateOrganization() {
	t := suite.T()

	parentOrg, err := suite.client.api.GetOrganizationByID(testUser1.UserCtx, testUser1.OrganizationID)
	require.NoError(t, err)

	// setup deleted org
	orgToDelete := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	// delete said org
	(&OrganizationCleanup{client: suite.client, ID: orgToDelete.ID}).MustDelete(testUser1.UserCtx, t)

	// avatar file setup
	avatarFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	require.NoError(t, err)

	invalidAvatarFile, err := objects.NewUploadFile("testdata/uploads/hello.txt")
	require.NoError(t, err)

	testCases := []struct {
		name                     string
		orgName                  string
		displayName              string
		orgDescription           string
		parentOrgID              string
		avatarFile               *graphql.Upload
		settings                 *openlaneclient.CreateOrganizationSettingInput
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
			expectedDefaultOrgUpdate: true, // only the first org created should update the default org
			parentOrgID:              "",   // root org
			client:                   suite.client.api,
			ctx:                      testUser1.UserCtx,
		},
		{
			name:           "happy path organization with settings and avatar",
			orgName:        gofakeit.Name(),
			displayName:    gofakeit.LetterN(50),
			orgDescription: gofakeit.HipsterSentence(10),
			avatarFile: &graphql.Upload{
				File:        avatarFile.File,
				Filename:    avatarFile.Filename,
				Size:        avatarFile.Size,
				ContentType: avatarFile.ContentType,
			},
			settings: &openlaneclient.CreateOrganizationSettingInput{
				Domains: []string{"meow.theopenlane.io"},
				BillingAddress: &models.Address{
					Line1:      gofakeit.StreetNumber() + " " + gofakeit.Street(),
					City:       gofakeit.City(),
					State:      gofakeit.State(),
					PostalCode: gofakeit.Zip(),
					Country:    gofakeit.Country(),
				},
			},
			parentOrgID: "", // root org
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name:           "happy path organization with parent org",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(10),
			parentOrgID:    testUser1.OrganizationID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "happy path organization with parent org using personal access token",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(10),
			parentOrgID:    testUser1.OrganizationID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
		},
		{
			name:           "organization with parent personal org",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(10),
			parentOrgID:    testUser1.PersonalOrgID,
			errorMsg:       "personal organizations are not allowed to have child organizations",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "empty organization name",
			orgName:        "",
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "value is less than the required length",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "long organization name",
			orgName:        gofakeit.LetterN(161),
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "value is greater than the required length",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "organization with no description",
			orgName:        gofakeit.Name(),
			orgDescription: "",
			parentOrgID:    testUser1.OrganizationID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "duplicate organization name",
			orgName:        parentOrg.Organization.Name,
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "already exists",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "duplicate organization name, case insensitive",
			orgName:        strings.ToUpper(parentOrg.Organization.Name),
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "already exists",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "duplicate organization name, but other was deleted, should pass",
			orgName:        orgToDelete.Name,
			orgDescription: gofakeit.HipsterSentence(10),
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "organization name with trailing space should work with trailing space removed",
			orgName:        "orgname ",
			orgDescription: gofakeit.HipsterSentence(10),
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "invalid organization name, too short",
			orgName:        "ab",
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "value is less than the required length",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{

			name:           "invalid organization name with special characters",
			orgName:        "orgn!me$",
			orgDescription: gofakeit.HipsterSentence(10),
			errorMsg:       "invalid or unparsable field: name, field cannot contain special characters",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "duplicate display name, should be allowed",
			orgName:        gofakeit.LetterN(80),
			displayName:    parentOrg.Organization.DisplayName,
			orgDescription: gofakeit.HipsterSentence(10),
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "display name with spaces should pass",
			orgName:        gofakeit.Name(),
			displayName:    gofakeit.Sentence(3),
			orgDescription: gofakeit.HipsterSentence(10),
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:    "invalid avatar file",
			orgName: gofakeit.Name(),
			avatarFile: &graphql.Upload{
				File:        invalidAvatarFile.File,
				Filename:    invalidAvatarFile.Filename,
				Size:        invalidAvatarFile.Size,
				ContentType: invalidAvatarFile.ContentType,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "unsupported mime type uploaded: text/plain",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*tc.avatarFile})
				} else {
					expectUploadCheckOnly(t, suite.client.objectStore.Storage)
				}
			}

			input := openlaneclient.CreateOrganizationInput{
				Name:        tc.orgName,
				Description: &tc.orgDescription,
			}

			if tc.displayName != "" {
				input.DisplayName = &tc.displayName
			}

			if tc.parentOrgID != "" {
				input.ParentID = &tc.parentOrgID
			}

			if tc.settings != nil {
				input.CreateOrgSettings = tc.settings
			}

			resp, err := tc.client.CreateOrganization(tc.ctx, input, tc.avatarFile)

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
			assert.Equal(t, strings.TrimSpace(tc.orgName), resp.CreateOrganization.Organization.Name)
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

			// Ensure avatar file is not empty
			if tc.avatarFile != nil {
				assert.NotNil(t, resp.CreateOrganization.Organization.AvatarLocalFileID)
				assert.NotNil(t, resp.CreateOrganization.Organization.AvatarFile.PresignedURL)
			}

			if tc.settings != nil {
				assert.Len(t, resp.CreateOrganization.Organization.Setting.Domains, 1)

				// make sure default org is updated if it's the first org created
				userResp, err := tc.client.GetUserByID(tc.ctx, testUser1.ID)
				require.NoError(t, err)

				if tc.expectedDefaultOrgUpdate {
					assert.Equal(t, resp.CreateOrganization.Organization.ID, userResp.User.Setting.DefaultOrg.ID)
				} else {
					assert.NotEqual(t, resp.CreateOrganization.Organization.ID, userResp.User.Setting.DefaultOrg.ID)
				}

				if tc.settings.BillingAddress != nil {
					assert.Equal(t, tc.settings.BillingAddress.Line1, resp.CreateOrganization.Organization.Setting.BillingAddress.Line1)
					assert.Equal(t, tc.settings.BillingAddress.City, resp.CreateOrganization.Organization.Setting.BillingAddress.City)
					assert.Equal(t, tc.settings.BillingAddress.State, resp.CreateOrganization.Organization.Setting.BillingAddress.State)
					assert.Equal(t, tc.settings.BillingAddress.PostalCode, resp.CreateOrganization.Organization.Setting.BillingAddress.PostalCode)
					assert.Equal(t, tc.settings.BillingAddress.Country, resp.CreateOrganization.Organization.Setting.BillingAddress.Country)
				}
			}

			// ensure entity types are created
			newCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, resp.CreateOrganization.Organization.ID)
			require.NoError(t, err)

			et, err := suite.client.api.GetEntityTypes(newCtx, &openlaneclient.EntityTypeWhereInput{
				OwnerID: &resp.CreateOrganization.Organization.ID,
			})
			require.NoError(t, err)

			require.Len(t, et.EntityTypes.Edges, 1)
			assert.Equal(t, "vendor", et.EntityTypes.Edges[0].Node.Name)
			assert.Equal(t, resp.CreateOrganization.Organization.ID, *et.EntityTypes.Edges[0].Node.OwnerID)

			// ensure managed groups are created
			managedGroups, err := suite.client.api.GetGroups(newCtx, &openlaneclient.GroupWhereInput{
				IsManaged: lo.ToPtr(true),
			})

			// admins, viewers, all users should be created
			require.Len(t, managedGroups.Groups.Edges, 3)

			// cleanup org
			(&OrganizationCleanup{client: suite.client, ID: resp.CreateOrganization.Organization.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func (suite *GraphTestSuite) TestMutationBulkCSVUploadOrganization() {
	t := suite.T()

	fileName := "testdata/uploads/orgs.csv"

	input, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	require.NoError(t, err)

	file := graphql.Upload{
		File:        input,
		Filename:    fileName,
		ContentType: "text/csv",
	}

	resp, err := suite.client.api.CreateBulkCSVOrganization(testUser1.UserCtx, file)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// make sure the orgs were created
	assert.Len(t, resp.CreateBulkCSVOrganization.Organizations, 2)
}

func (suite *GraphTestSuite) TestMutationUpdateOrganization() {
	t := suite.T()

	nameUpdate := gofakeit.Name()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence(10)
	nameUpdateLong := gofakeit.LetterN(200)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	user1 := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create groups for creator permissions tests and add a member
	// group created by org owner
	groupProgramCreators := (&GroupBuilder{client: suite.client, Owner: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: groupProgramCreators.ID}).MustNew(testUser1.UserCtx, t)

	anotherGroupProgramCreators := (&GroupBuilder{client: suite.client, Owner: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: anotherGroupProgramCreators.ID}).MustNew(testUser1.UserCtx, t)

	// group created by admin
	groupProcedureCreators := (&GroupBuilder{client: suite.client, Owner: testUser1.OrganizationID}).MustNew(adminUser.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: groupProcedureCreators.ID}).MustNew(adminUser.UserCtx, t)

	reqCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, org.ID)
	require.NoError(t, err)

	// add a member to the org, to test permissions
	om := (&OrgMemberBuilder{client: suite.client, OrgID: org.ID, Role: string(enums.RoleMember)}).MustNew(testUser1.UserCtx, t)
	memberUserCtx, err := auth.NewTestContextWithOrgID(om.UserID, org.ID)
	require.NoError(t, err)

	// avatar file setup
	avatarFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	require.NoError(t, err)

	invalidAvatarFile, err := objects.NewUploadFile("testdata/uploads/hello.txt")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		orgID       string
		updateInput openlaneclient.UpdateOrganizationInput
		avatarFile  *graphql.Upload
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedRes openlaneclient.UpdateOrganization_UpdateOrganization_Organization
		errorMsg    string
	}{
		{
			name:  "update name, happy path",
			orgID: org.ID,
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
			name:  "add member as admin",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				AddOrgMembers: []*openlaneclient.CreateOrgMembershipInput{
					{
						UserID: user1.ID,
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
						UserID: user1.ID,
					},
				},
			},
		},
		{
			name:  "add two program creators group",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				AddProgramCreatorIDs: []string{groupProgramCreators.ID, anotherGroupProgramCreators.ID},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
				ProgramCreators: []*openlaneclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators{
					{
						ID:          groupProgramCreators.ID,
						DisplayID:   groupProgramCreators.DisplayID,
						Name:        groupProgramCreators.Name,
						DisplayName: groupProgramCreators.DisplayName,
					},
					{
						ID:          anotherGroupProgramCreators.ID,
						DisplayID:   anotherGroupProgramCreators.DisplayID,
						Name:        anotherGroupProgramCreators.Name,
						DisplayName: anotherGroupProgramCreators.DisplayName,
					},
				},
			},
		},
		{
			name:  "remove one program creator group, add procedure creator group",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				RemoveProgramCreatorIDs: []string{groupProgramCreators.ID},
				AddProcedureCreatorIDs:  []string{groupProcedureCreators.ID},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: openlaneclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
				ProcedureCreators: []*openlaneclient.UpdateOrganization_UpdateOrganization_Organization_ProcedureCreators{
					{
						ID:          groupProcedureCreators.ID,
						DisplayID:   groupProcedureCreators.DisplayID,
						Name:        groupProcedureCreators.Name,
						DisplayName: groupProcedureCreators.DisplayName,
					},
				},
				ProgramCreators: []*openlaneclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators{
					{
						ID:          anotherGroupProgramCreators.ID,
						DisplayID:   anotherGroupProgramCreators.DisplayID,
						Name:        anotherGroupProgramCreators.Name,
						DisplayName: anotherGroupProgramCreators.DisplayName,
					},
				},
			},
		},
		{
			name:  "add program creator group, not allowed",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				AddProgramCreatorIDs: []string{groupProgramCreators.ID},
			},
			client:   suite.client.api,
			ctx:      memberUserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:  "update description and avatar file, happy path",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				Description: &descriptionUpdate,
			},
			avatarFile: &graphql.Upload{
				File:        avatarFile.File,
				Filename:    avatarFile.Filename,
				Size:        avatarFile.Size,
				ContentType: avatarFile.ContentType,
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
			name:  "update display name, happy path",
			orgID: org.ID,
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
			name:  "update settings, happy path",
			orgID: org.ID,
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
			name:  "update avatar, invalid file",
			orgID: org.ID,
			avatarFile: &graphql.Upload{
				File:        invalidAvatarFile.File,
				Filename:    invalidAvatarFile.Filename,
				Size:        invalidAvatarFile.Size,
				ContentType: invalidAvatarFile.ContentType,
			},
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "unsupported mime type uploaded: text/plain",
		},
		{
			name:  "update name, too long",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				Name: &nameUpdateLong,
			},
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "value is greater than the required length",
		},
		{
			name:  "update name, no access",
			orgID: viewOnlyUser.OrganizationID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				Name: &nameUpdate,
			},
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:  "update name, not found",
			orgID: org.ID,
			updateInput: openlaneclient.UpdateOrganizationInput{
				Name: &nameUpdate,
			},
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*tc.avatarFile})
				} else {
					expectUploadCheckOnly(t, suite.client.objectStore.Storage)
				}
			}

			resp, err := tc.client.UpdateOrganization(tc.ctx, tc.orgID, tc.updateInput, tc.avatarFile)

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
				// Adding a member to an org will make it 3 users, there is an owner
				// assigned to the org automatically and an another member added in the test
				assert.Len(t, updatedOrg.Members, 3)
				assert.Equal(t, tc.expectedRes.Members[0].Role, updatedOrg.Members[2].Role)
				assert.Equal(t, tc.expectedRes.Members[0].UserID, updatedOrg.Members[2].UserID)
			}

			if tc.updateInput.UpdateOrgSettings != nil {
				assert.Len(t, updatedOrg.GetSetting().Domains, 2)
			}

			if tc.avatarFile != nil {
				assert.NotNil(t, updatedOrg.AvatarLocalFileID)
				assert.NotNil(t, updatedOrg.AvatarFile.PresignedURL)
			}
		})
	}

	(&OrganizationCleanup{client: suite.client, ID: org.ID}).MustDelete(reqCtx, t)
	(&UserCleanup{client: suite.client, ID: testUser1.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationDeleteOrganization() {
	t := suite.T()

	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	reqCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, org.ID)
	require.NoError(t, err)

	setting, err := suite.client.api.UpdateUserSetting(reqCtx, testUser1.UserInfo.Edges.Setting.ID,
		openlaneclient.UpdateUserSettingInput{
			DefaultOrgID: &org.ID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, org.ID, setting.UpdateUserSetting.UserSetting.DefaultOrg.ID)

	testCases := []struct {
		name     string
		orgID    string
		ctx      context.Context
		errorMsg string
	}{
		{
			name:     "delete org, access denied",
			orgID:    viewOnlyUser.OrganizationID,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:     "delete org, not found",
			orgID:    org.ID,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:  "delete org, happy path",
			orgID: org.ID,
			ctx:   testUser1.UserCtx,
		},
		{
			name:     "delete org, personal org not allowed",
			orgID:    testUser1.PersonalOrgID,
			ctx:      testUser1.UserCtx,
			errorMsg: "cannot delete personal organizations",
		},
		{
			name:     "delete org, not found",
			orgID:    "tacos-tuesday",
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.DeleteOrganization(tc.ctx, tc.orgID)

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
			settingUpdated, err := suite.client.api.GetUserSettingByID(reqCtx, testUser1.UserInfo.Edges.Setting.ID)
			require.NoError(t, err)
			require.NotNil(t, settingUpdated.UserSetting.DefaultOrg)
			assert.NotEqual(t, org.ID, settingUpdated.UserSetting.DefaultOrg.ID)

			o, err := suite.client.api.GetOrganizationByID(reqCtx, tc.orgID)

			require.Nil(t, o)
			require.Error(t, err)
			assert.ErrorContains(t, err, notFoundErrorMsg)

			// tuples and entity are deleted, so we need to skip soft delete and privacy checks
			ctx := entx.SkipSoftDelete(reqCtx)
			ctx = privacy.DecisionContext(ctx, privacy.Allow)

			o, err = suite.client.api.GetOrganizationByID(ctx, tc.orgID)
			require.NoError(t, err)
			require.NotNil(t, o)

			require.Equal(t, o.Organization.ID, tc.orgID)
		})
	}
}

func (suite *GraphTestSuite) TestMutationOrganizationCascadeDelete() {
	t := suite.T()

	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	reqCtx, err := auth.NewTestContextWithOrgID(testUser1.ID, org.ID)
	require.NoError(t, err)

	// add child org
	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: org.ID}).MustNew(reqCtx, t)

	group1 := (&GroupBuilder{client: suite.client, Owner: org.ID}).MustNew(reqCtx, t)

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
	assert.ErrorContains(t, err, notFoundErrorMsg)

	co, err := suite.client.api.GetOrganizationByID(reqCtx, childOrg.ID)

	require.Nil(t, co)
	require.Error(t, err)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	g, err := suite.client.api.GetGroupByID(reqCtx, group1.ID)

	require.Nil(t, g)
	require.Error(t, err)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	// allow after tuples have been deleted
	ctx := privacy.DecisionContext(reqCtx, privacy.Allow)
	ctx = entx.SkipSoftDelete(ctx)

	o, err = suite.client.api.GetOrganizationByID(ctx, org.ID)

	require.NoError(t, err)
	require.Equal(t, o.Organization.ID, org.ID)

	// allow after tuples have been deleted
	ctx = privacy.DecisionContext(reqCtx, privacy.Allow)
	ctx = entx.SkipSoftDelete(ctx)

	g, err = suite.client.api.GetGroupByID(ctx, group1.ID)
	require.NoError(t, err)
	require.Equal(t, g.Group.ID, group1.ID)

	// allow after tuples have been deleted
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	ctx = entx.SkipSoftDelete(ctx)

	co, err = suite.client.api.GetOrganizationByID(ctx, childOrg.ID)
	require.NoError(t, err)

	require.Equal(t, co.Organization.ID, childOrg.ID)
}
