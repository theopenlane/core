package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryOrganization(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)
	apiTokenClient := suite.setupAPITokenClient(orgUser.UserCtx, t)
	patTokenClient := suite.setupPatClient(orgUser, t)

	// create api token for the user
	(&APITokenBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)
	// create personal access token for the user
	(&PersonalAccessTokenBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

	// add org members
	om := (&OrgMemberBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext("abc123", orgUser.OrganizationID)

	testCases := []struct {
		name               string
		queryID            string
		client             *testclient.TestClient
		ctx                context.Context
		orgMembersExpected int
		errorMsg           string
	}{
		{
			name:               "happy path, get organization",
			queryID:            orgUser.OrganizationID,
			client:             suite.client.api,
			ctx:                orgUser.UserCtx,
			orgMembersExpected: 2, // owner and 1 member
		},
		{
			name:               "happy path, get using api token",
			queryID:            orgUser.OrganizationID,
			client:             apiTokenClient,
			ctx:                context.Background(),
			orgMembersExpected: 2, // owner and 1 member
		},
		{
			name:               "happy path, get using personal access token",
			queryID:            orgUser.OrganizationID,
			client:             patTokenClient,
			ctx:                context.Background(),
			orgMembersExpected: 2, // owner and 1 member
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			client:   suite.client.api,
			ctx:      orgUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  orgUser.OrganizationID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetOrganizationByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.Organization.Members.Edges, tc.orgMembersExpected), "expected %d org members, got %d", tc.orgMembersExpected, len(resp.Organization.Members.Edges))

			if tc.orgMembersExpected > 1 {
				orgMemberFound := false

				for _, m := range resp.Organization.Members.Edges {
					if m.Node.User.ID == om.UserID {
						orgMemberFound = true
					}
				}

				assert.Check(t, orgMemberFound)
			}
		})
	}
}

func TestQueryOrganizations(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)

	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

	t.Run("Get Organizations", func(t *testing.T) {
		resp, err := suite.client.api.GetAllOrganizations(orgUser.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Organizations.Edges != nil)

		// make sure two organizations are returned, the two created and
		// the personal org and test org created at suite setup
		assert.Check(t, is.Equal(4, len(resp.Organizations.Edges)))

		org1Found := false
		org2Found := false

		for _, o := range resp.Organizations.Edges {
			if o.Node.ID == org1.ID {
				org1Found = true
			} else if o.Node.ID == org2.ID {
				org2Found = true
			}
		}

		assert.Check(t, org1Found)
		assert.Check(t, org2Found)
	})

	// cleanup orgs
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, IDs: []string{org1.ID, org2.ID}}).MustDelete(orgUser.UserCtx, t)
}

func TestMutationCreateOrganization(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(orgUser, t)
	tokenClient := suite.setupAPITokenClient(orgUser.UserCtx, t)

	parentOrg, err := suite.client.api.GetOrganizationByID(orgUser.UserCtx, orgUser.OrganizationID)
	assert.NilError(t, err)

	// setup deleted org
	orgToDelete := (&OrganizationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)
	// delete said org
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, ID: orgToDelete.ID}).MustDelete(orgUser.UserCtx, t)

	// avatar file setup
	avatarFile := uploadFile(t, logoFilePath)
	invalidAvatarFile := uploadFile(t, txtFilePath)

	testCases := []struct {
		name                     string
		orgName                  string
		displayName              string
		orgDescription           string
		parentOrgID              string
		avatarFile               *graphql.Upload
		settings                 *testclient.CreateOrganizationSettingInput
		client                   *testclient.TestClient
		ctx                      context.Context
		expectedDefaultOrgUpdate bool
		errorMsg                 string
	}{
		{
			name:                     "happy path organization",
			orgName:                  ulids.New().String(), // use ulid to ensure uniqueness
			displayName:              gofakeit.LetterN(50),
			orgDescription:           gofakeit.HipsterSentence(),
			expectedDefaultOrgUpdate: true, // only the first org created should update the default org
			parentOrgID:              "",   // root org
			client:                   suite.client.api,
			ctx:                      orgUser.UserCtx,
		},
		{
			name:           "happy path organization with settings and avatar",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			displayName:    gofakeit.LetterN(50),
			orgDescription: gofakeit.HipsterSentence(),
			avatarFile:     avatarFile,
			settings: &testclient.CreateOrganizationSettingInput{
				Domains:                      []string{"meow.theopenlane.io"},
				AllowedEmailDomains:          []string{"theopenlane.io"},
				AllowMatchingDomainsAutojoin: lo.ToPtr(true),
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
			ctx:         orgUser.UserCtx,
		},
		{
			name:           "organization settings with free email domain not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			displayName:    gofakeit.LetterN(50),
			orgDescription: gofakeit.HipsterSentence(),
			settings: &testclient.CreateOrganizationSettingInput{
				AllowedEmailDomains: []string{"gmail.com"},
			},
			parentOrgID: "", // root org
			client:      suite.client.api,
			ctx:         orgUser.UserCtx,
			errorMsg:    invalidInputErrorMsg,
		},
		{
			name:           "happy path organization with parent org",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    orgUser.OrganizationID,
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "organization with parent org, no access",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    testUser2.OrganizationID,
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
			errorMsg:       notAuthorizedErrorMsg,
		},
		{
			name:           "organization with parent org using personal access token, not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    orgUser.OrganizationID,
			client:         patClient,
			ctx:            context.Background(),
			errorMsg:       common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name:           "organization with parent org using personal access token, no access to parent, not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    testUser2.OrganizationID,
			client:         patClient,
			ctx:            context.Background(),
			errorMsg:       common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name:           "organization create with api token not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			client:         tokenClient,
			ctx:            context.Background(),
			errorMsg:       common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name:           "organization with parent personal org",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    orgUser.PersonalOrgID,
			errorMsg:       "personal organizations are not allowed to have child organizations",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "empty organization name",
			orgName:        "",
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "value is less than the required length",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "long organization name",
			orgName:        gofakeit.LetterN(161),
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "value is greater than the required length",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "organization with no description",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: "",
			parentOrgID:    orgUser.OrganizationID,
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "duplicate organization name",
			orgName:        parentOrg.Organization.Name,
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "already exists",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "duplicate organization name, case insensitive",
			orgName:        strings.ToUpper(parentOrg.Organization.Name),
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "already exists",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "duplicate organization name, but other was deleted, should pass",
			orgName:        orgToDelete.Name,
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "organization name with trailing space should work with trailing space removed",
			orgName:        "orgname ",
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "invalid organization name, too short",
			orgName:        "ab",
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "value is less than the required length",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{

			name:           "invalid organization name with special characters",
			orgName:        "orgn!me$",
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "invalid or unparsable field: name, field cannot contain special characters",
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "duplicate display name, should be allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			displayName:    parentOrg.Organization.DisplayName,
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:           "display name with spaces should pass",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			displayName:    gofakeit.Sentence(),
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.client.api,
			ctx:            orgUser.UserCtx,
		},
		{
			name:       "invalid avatar file",
			orgName:    ulids.New().String(), // use ulid to ensure uniqueness
			avatarFile: invalidAvatarFile,
			client:     suite.client.api,
			ctx:        orgUser.UserCtx,
			errorMsg:   "unsupported mime type uploaded: text/plain",
		},
		{
			name:    "invalid allowed email domains ",
			orgName: ulids.New().String(), // use ulid to ensure uniqueness
			settings: &testclient.CreateOrganizationSettingInput{
				AllowedEmailDomains: []string{"theopenlane"},
			},
			client:   suite.client.api,
			ctx:      orgUser.UserCtx,
			errorMsg: "invalid or unparsable field: domains",
		},
		{
			name:    "invalid domains",
			orgName: ulids.New().String(), // use ulid to ensure uniqueness
			settings: &testclient.CreateOrganizationSettingInput{
				Domains: []string{"theopenlane"},
			},
			client:   suite.client.api,
			ctx:      orgUser.UserCtx,
			errorMsg: "invalid or unparsable field: domains",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.avatarFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}

			input := testclient.CreateOrganizationInput{
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
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Make sure provided values match
			assert.Check(t, is.Equal(strings.TrimSpace(tc.orgName), resp.CreateOrganization.Organization.Name))
			assert.Check(t, is.Equal(tc.orgDescription, *resp.CreateOrganization.Organization.Description))

			if tc.parentOrgID == "" {
				assert.Check(t, is.Nil(resp.CreateOrganization.Organization.Parent))
			} else {
				parent := resp.CreateOrganization.Organization.GetParent()
				assert.Check(t, is.Equal(tc.parentOrgID, parent.ID))
			}

			// Ensure org settings is not null
			assert.Check(t, resp.CreateOrganization.Organization.Setting.ID != "")

			// Ensure display name is not empty
			assert.Check(t, len(resp.CreateOrganization.Organization.DisplayName) != 0)

			// Ensure avatar file is not empty
			if tc.avatarFile != nil {
				assert.Check(t, resp.CreateOrganization.Organization.AvatarLocalFileID != nil)
				assert.Check(t, resp.CreateOrganization.Organization.AvatarFile.PresignedURL != nil)
			}

			if tc.settings != nil {
				assert.Check(t, is.Len(resp.CreateOrganization.Organization.Setting.Domains, 1))

				// make sure default org is updated if it's the first org created
				userResp, err := tc.client.GetUserByID(tc.ctx, orgUser.ID)
				assert.NilError(t, err)

				if tc.expectedDefaultOrgUpdate {
					assert.Check(t, is.Equal(resp.CreateOrganization.Organization.ID, userResp.User.Setting.DefaultOrg.ID))
				} else {
					assert.Check(t, resp.CreateOrganization.Organization.ID != userResp.User.Setting.DefaultOrg.ID)
				}

				if tc.settings.BillingAddress != nil {
					assert.Check(t, is.Equal(tc.settings.BillingAddress.Line1, resp.CreateOrganization.Organization.Setting.BillingAddress.Line1))
					assert.Check(t, is.Equal(tc.settings.BillingAddress.City, resp.CreateOrganization.Organization.Setting.BillingAddress.City))
					assert.Check(t, is.Equal(tc.settings.BillingAddress.State, resp.CreateOrganization.Organization.Setting.BillingAddress.State))
					assert.Check(t, is.Equal(tc.settings.BillingAddress.PostalCode, resp.CreateOrganization.Organization.Setting.BillingAddress.PostalCode))
					assert.Check(t, is.Equal(tc.settings.BillingAddress.Country, resp.CreateOrganization.Organization.Setting.BillingAddress.Country))
				}

				// ensure the allowed email domains is set properly
				assert.Check(t, is.Contains(resp.CreateOrganization.Organization.Setting.AllowedEmailDomains, userResp.User.Email[strings.Index(userResp.User.Email, "@")+1:]))
				assert.Check(t, is.Equal(true, *resp.CreateOrganization.Organization.Setting.AllowMatchingDomainsAutojoin))
			}

			// ensure entity types are created
			newCtx := auth.NewTestContextWithOrgID(orgUser.ID, resp.CreateOrganization.Organization.ID)

			et, err := suite.client.api.GetEntityTypes(newCtx, &testclient.EntityTypeWhereInput{
				OwnerID: &resp.CreateOrganization.Organization.ID,
			})
			assert.NilError(t, err)

			assert.Assert(t, is.Len(et.EntityTypes.Edges, 1))

			// ensure managed groups are created
			managedGroups, err := suite.client.api.GetGroups(newCtx, &testclient.GroupWhereInput{
				IsManaged: lo.ToPtr(true),
			})
			assert.NilError(t, err)

			// ensure owner is in the managed group
			for _, g := range managedGroups.Groups.Edges {
				if g.Node.Name == "Viewers" {
					assert.Check(t, is.Len(g.Node.Members.Edges, 0))
				} else {
					assert.Check(t, is.Len(g.Node.Members.Edges, 1))
				}
			}

			// while group is in the base module, this query includes programs and others
			// which are in other modules
			//
			// 4 groups because a system managed group is now created for each user
			// in the organization
			num := 4
			if tc.parentOrgID != "" {
				num = 3
			}

			assert.Check(t, is.Len(managedGroups.Groups.Edges, num))

			// cleanup org
			(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, ID: resp.CreateOrganization.Organization.ID}).MustDelete(orgUser.UserCtx, t)
		})
	}
}

func TestMutationUpdateOrganization(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)

	nameUpdate := ulids.New().String()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence()
	nameUpdateLong := gofakeit.LetterN(200)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)
	user1 := (&UserBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

	reqCtx := auth.NewTestContextWithOrgID(orgUser.ID, org.ID)

	// create groups for creator permissions tests and add a member
	// group created by org owner
	groupProgramCreators := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)
	anotherGroupProgramCreators := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)
	groupProcedureCreators := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	(&GroupMemberBuilder{client: suite.client, GroupID: groupProgramCreators.ID}).MustNew(reqCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: anotherGroupProgramCreators.ID}).MustNew(reqCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: groupProcedureCreators.ID}).MustNew(reqCtx, t)

	// add a member to the org, to test permissions
	om := (&OrgMemberBuilder{client: suite.client}).MustNew(reqCtx, t)
	memberUserCtx := auth.NewTestContextWithOrgID(om.UserID, org.ID)

	// avatar file setup
	avatarFile := uploadFile(t, logoFilePath)
	invalidAvatarFile := uploadFile(t, txtFilePath)

	testCases := []struct {
		name        string
		orgID       string
		updateInput testclient.UpdateOrganizationInput
		avatarFile  *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedRes testclient.UpdateOrganization_UpdateOrganization_Organization
		errorMsg    string
	}{
		{
			name:  "update name, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				Name: &nameUpdate,
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
			},
		},
		{
			name:  "add member as admin",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				AddOrgMembers: []*testclient.CreateOrgMembershipInput{
					{
						UserID: user1.ID,
						Role:   &enums.RoleAdmin,
					},
				},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
				Members: testclient.UpdateOrganization_UpdateOrganization_Organization_Members{
					Edges: []*testclient.UpdateOrganization_UpdateOrganization_Organization_Members_Edges{
						{
							Node: &testclient.UpdateOrganization_UpdateOrganization_Organization_Members_Edges_Node{
								Role:   enums.RoleAdmin,
								UserID: user1.ID,
							},
						},
					},
				},
			},
		},
		{
			name:  "add two program creators group",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				AddProgramCreatorIDs: []string{groupProgramCreators.ID, anotherGroupProgramCreators.ID},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
				ProgramCreators: testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators{
					Edges: []*testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators_Edges{
						{
							Node: &testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators_Edges_Node{
								ID:          groupProgramCreators.ID,
								DisplayID:   groupProgramCreators.DisplayID,
								Name:        groupProgramCreators.Name,
								DisplayName: groupProgramCreators.DisplayName,
							},
						},
						{
							Node: &testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators_Edges_Node{
								ID:          anotherGroupProgramCreators.ID,
								DisplayID:   anotherGroupProgramCreators.DisplayID,
								Name:        anotherGroupProgramCreators.Name,
								DisplayName: anotherGroupProgramCreators.DisplayName,
							},
						},
					},
				},
			},
		},
		{
			name:  "remove one program creator group, add procedure creator group",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				RemoveProgramCreatorIDs: []string{groupProgramCreators.ID},
				AddProcedureCreatorIDs:  []string{groupProcedureCreators.ID},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,
				DisplayName: org.DisplayName,
				Description: &org.Description,
				ProcedureCreators: testclient.UpdateOrganization_UpdateOrganization_Organization_ProcedureCreators{
					Edges: []*testclient.UpdateOrganization_UpdateOrganization_Organization_ProcedureCreators_Edges{
						{
							Node: &testclient.UpdateOrganization_UpdateOrganization_Organization_ProcedureCreators_Edges_Node{
								ID:          groupProcedureCreators.ID,
								DisplayID:   groupProcedureCreators.DisplayID,
								Name:        groupProcedureCreators.Name,
								DisplayName: groupProcedureCreators.DisplayName,
							},
						},
					},
				},
				ProgramCreators: testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators{
					Edges: []*testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators_Edges{
						{
							Node: &testclient.UpdateOrganization_UpdateOrganization_Organization_ProgramCreators_Edges_Node{
								ID:          anotherGroupProgramCreators.ID,
								DisplayID:   anotherGroupProgramCreators.DisplayID,
								Name:        anotherGroupProgramCreators.Name,
								DisplayName: anotherGroupProgramCreators.DisplayName,
							},
						},
					},
				},
			},
		},
		{
			name:  "add program creator group, not allowed",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				AddProgramCreatorIDs: []string{groupProgramCreators.ID},
			},
			client:   suite.client.api,
			ctx:      memberUserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:  "update description and avatar file, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				Description: &descriptionUpdate,
			},
			avatarFile: avatarFile,
			client:     suite.client.api,
			ctx:        reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate, // this would have been updated on the prior test
				DisplayName: org.DisplayName,
				Description: &descriptionUpdate,
			},
		},
		{
			name:  "update display name, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				DisplayName: &displayNameUpdate,
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate, // this would have been updated on the prior test
				DisplayName: displayNameUpdate,
				Description: &descriptionUpdate,
			},
		},
		{
			name:  "update settings, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				Description: &descriptionUpdate,
				UpdateOrgSettings: &testclient.UpdateOrganizationSettingInput{
					Domains: []string{"meow.theopenlane.io", "woof.theopenlane.io"},
				},
			},
			client: suite.client.api,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        nameUpdate,        // this would have been updated on the prior test
				DisplayName: displayNameUpdate, // this would have been updated on the prior test
				Description: &descriptionUpdate,
			},
		},
		{
			name:       "update avatar, invalid file",
			orgID:      org.ID,
			avatarFile: invalidAvatarFile,
			client:     suite.client.api,
			ctx:        reqCtx,
			errorMsg:   "unsupported mime type uploaded: text/plain",
		},
		{
			name:  "update name, too long",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				Name: &nameUpdateLong,
			},
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "value is greater than the required length",
		},
		{
			name:  "update name, no access",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				Name: &nameUpdate,
			},
			client:   suite.client.api,
			ctx:      memberUserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:  "update name, not found",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
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
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.avatarFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}

			resp, err := tc.client.UpdateOrganization(tc.ctx, tc.orgID, tc.updateInput, tc.avatarFile)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Make sure provided values match
			updatedOrg := resp.GetUpdateOrganization().Organization
			assert.Check(t, is.Equal(tc.expectedRes.Name, updatedOrg.Name))
			assert.Check(t, is.Equal(tc.expectedRes.DisplayName, updatedOrg.DisplayName))
			assert.Check(t, is.DeepEqual(tc.expectedRes.Description, updatedOrg.Description))

			if tc.updateInput.AddOrgMembers != nil {
				// Adding a member to an org will make it 3 users, there is an owner
				// assigned to the org automatically and an another member added in the test and
				// 3 created as part of the group member logic
				assert.Check(t, is.Len(updatedOrg.Members.Edges, 6))
				assert.Check(t, is.Equal(tc.expectedRes.Members.Edges[0].Node.Role, updatedOrg.Members.Edges[5].Node.Role))
				assert.Check(t, is.Equal(tc.expectedRes.Members.Edges[0].Node.UserID, updatedOrg.Members.Edges[5].Node.UserID))
			}

			if tc.updateInput.UpdateOrgSettings != nil {
				assert.Check(t, is.Len(updatedOrg.GetSetting().Domains, 2))
			}

			if tc.avatarFile != nil {
				assert.Check(t, updatedOrg.AvatarLocalFileID != nil)
				assert.Check(t, updatedOrg.AvatarFile.PresignedURL != nil)
			}
		})
	}

	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, ID: org.ID}).MustDelete(reqCtx, t)
	(&Cleanup[*generated.UserDeleteOne]{client: suite.client.db.User, ID: user1.ID}).MustDelete(reqCtx, t)
}

func TestMutationDeleteOrganization(t *testing.T) {
	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)

	reqCtx := orgUser.UserCtx

	setting, err := suite.client.api.UpdateUserSetting(reqCtx, orgUser.UserInfo.Edges.Setting.ID,
		testclient.UpdateUserSettingInput{
			DefaultOrgID: &orgUser.OrganizationID,
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, orgUser.OrganizationID, setting.UpdateUserSetting.UserSetting.DefaultOrg.ID)

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
			orgID:    orgUser.OrganizationID,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:  "delete org, happy path",
			orgID: orgUser.OrganizationID,
			ctx:   orgUser.UserCtx,
		},
		{
			name:     "delete org, personal org not allowed",
			orgID:    orgUser.PersonalOrgID,
			ctx:      orgUser.UserCtx,
			errorMsg: "cannot delete personal organizations",
		},
		{
			name:     "delete org, not found",
			orgID:    "tacos-tuesday",
			ctx:      orgUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.DeleteOrganization(tc.ctx, tc.orgID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.DeleteOrganization.DeletedID != "")

			// make sure the deletedID matches the ID we wanted to delete
			assert.Check(t, is.Equal(tc.orgID, resp.DeleteOrganization.DeletedID))

			// update the context to have the correct org after the org is deleted
			reqCtx := auth.NewTestContextWithOrgID(orgUser.ID, orgUser.OrganizationID)

			// make sure the default org is reset
			settingUpdated, err := suite.client.api.GetUserSettingByID(reqCtx, orgUser.UserInfo.Edges.Setting.ID)
			assert.NilError(t, err)
			assert.Assert(t, settingUpdated.UserSetting.DefaultOrg != nil)
			assert.Check(t, orgUser.OrganizationID != settingUpdated.UserSetting.DefaultOrg.ID)

			// allow ctx to ensure the org no longer exists after deletion
			allowCtx := ent.NewContext(rule.WithInternalContext(reqCtx), suite.client.db)

			_, err = suite.client.api.GetOrganizationByID(allowCtx, tc.orgID)
			assert.ErrorContains(t, err, notFoundErrorMsg)

			// tuples and entity are deleted, so we need to skip soft delete and privacy checks
			ctx := entx.SkipSoftDelete(reqCtx)
			ctx = privacy.DecisionContext(ctx, privacy.Allow)

			o, err := suite.client.api.GetOrganizationByID(ctx, tc.orgID)
			assert.NilError(t, err)
			assert.Assert(t, o != nil)

			assert.Equal(t, o.Organization.ID, tc.orgID)
		})
	}
}

func TestMutationOrganizationCascadeDelete(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

	reqCtx := auth.NewTestContextWithOrgID(orgUser.ID, org.ID)
	group1 := (&GroupBuilder{client: suite.client}).MustNew(reqCtx, t)

	// add child org
	childOrg := (&OrganizationBuilder{client: suite.client, ParentOrgID: org.ID}).MustNew(reqCtx, t)

	// delete org
	resp, err := suite.client.api.DeleteOrganization(reqCtx, org.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.DeleteOrganization.DeletedID != "")

	// make sure the deletedID matches the ID we wanted to delete
	assert.Check(t, is.Equal(org.ID, resp.DeleteOrganization.DeletedID))

	_, err = suite.client.api.GetOrganizationByID(reqCtx, org.ID)

	assert.ErrorContains(t, err, notFoundErrorMsg)

	_, err = suite.client.api.GetOrganizationByID(reqCtx, childOrg.ID)

	assert.ErrorContains(t, err, notFoundErrorMsg)

	_, err = suite.client.api.GetGroupByID(reqCtx, group1.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	// allow after tuples have been deleted
	ctx := privacy.DecisionContext(reqCtx, privacy.Allow)
	ctx = entx.SkipSoftDelete(ctx)

	o, err := suite.client.api.GetOrganizationByID(ctx, org.ID)

	assert.NilError(t, err)
	assert.Equal(t, o.Organization.ID, org.ID)

	// allow after tuples have been deleted
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	ctx = entx.SkipSoftDelete(ctx)

	co, err := suite.client.api.GetOrganizationByID(ctx, childOrg.ID)
	assert.NilError(t, err)

	assert.Equal(t, co.Organization.ID, childOrg.ID)
}
