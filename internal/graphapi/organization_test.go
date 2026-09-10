package graphapi_test

import (
	"context"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

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
	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryOrganization(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedOrgOwner(t)

	// create api token for the user
	(&th.APITokenBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)
	// create personal access token for the user
	(&th.PersonalAccessTokenBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)

	// add org members
	om := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext("abc123", localTestOrg.Owner.OrganizationID)

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
			queryID:            localTestOrg.Owner.OrganizationID,
			client:             suite.Client.API,
			ctx:                localTestOrg.Owner.UserCtx,
			orgMembersExpected: 2, // owner and 1 member
		},
		{
			name:               "happy path, get using api token",
			queryID:            localTestOrg.Owner.OrganizationID,
			client:             localTestOrg.APIClient,
			ctx:                context.Background(),
			orgMembersExpected: 2, // owner and 1 member
		},
		{
			name:               "happy path, get using personal access token",
			queryID:            localTestOrg.Owner.OrganizationID,
			client:             localTestOrg.PatClient,
			ctx:                context.Background(),
			orgMembersExpected: 2, // owner and 1 member
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			client:   suite.Client.API,
			ctx:      localTestOrg.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  localTestOrg.Owner.OrganizationID,
			errorMsg: th.NotFoundErrorMsg,
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

	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestQueryOrganizations(t *testing.T) {
	t.Parallel()

	orgUser := suite.SeedOrgOwner(t)

	org1ID := orgUser.Owner.OrganizationID
	avatarFile1 := th.UploadFile(t, th.LogoFilePath)

	input := testclient.CreateOrganizationInput{
		Name: "test-org-" + ulids.New().String(),
	}
	// mock expect upload
	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*avatarFile1})
	resp, err := suite.Client.API.CreateOrganization(orgUser.Owner.UserCtx, input, avatarFile1, nil)
	th.RequireNoError(t, err)
	org2ID := resp.CreateOrganization.Organization.ID
	assert.Assert(t, resp.CreateOrganization.Organization.AvatarFile != nil)
	avatarFileIDOrg2 := resp.CreateOrganization.Organization.AvatarFile.ID

	avatarFile2 := th.UploadFile(t, th.LogoFilePath)
	input2 := testclient.CreateOrganizationInput{
		Name: "test-org-" + ulids.New().String(),
	}
	// mock expect upload
	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*avatarFile2})
	resp2, err := suite.Client.API.CreateOrganization(orgUser.Owner.UserCtx, input2, avatarFile2, nil)
	th.RequireNoError(t, err)
	org3ID := resp2.CreateOrganization.Organization.ID
	assert.Assert(t, resp2.CreateOrganization.Organization.AvatarFile != nil)
	avatarFileIDOrg3 := resp2.CreateOrganization.Organization.AvatarFile.ID

	// ensure context only has one organization id set, this will mimic JWT authorization
	testContext := auth.NewTestContextWithOrgID(orgUser.Owner.ID, org1ID)

	t.Run("Get Organizations", func(t *testing.T) {
		resp, err := suite.Client.API.GetAllOrganizations(testContext)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Organizations.Edges != nil)

		// make sure 4 organizations are returned, the two created and
		// the personal org
		assert.Check(t, is.Equal(4, len(resp.Organizations.Edges)))

		org1Found := false
		org2Found := false
		org3Found := false

		for _, o := range resp.Organizations.Edges {
			if o.Node.ID == org1ID {
				org1Found = true
				// no avatar set
				assert.Check(t, o.Node.AvatarRemoteURL != nil)
				assert.Check(t, o.Node.AvatarFile == nil)
			} else if o.Node.ID == org2ID {
				org2Found = true
				assert.Assert(t, o.Node.AvatarFile != nil)
				assert.Check(t, is.Equal(o.Node.AvatarFile.ID, avatarFileIDOrg2))
				assert.Check(t, o.Node.AvatarFile.PresignedURL != nil)
			} else if o.Node.ID == org3ID {
				org3Found = true
				assert.Assert(t, o.Node.AvatarFile != nil)
				assert.Check(t, is.Equal(o.Node.AvatarFile.ID, avatarFileIDOrg3))
				assert.Check(t, o.Node.AvatarFile.PresignedURL != nil)
			}
		}

		assert.Check(t, org1Found)
		assert.Check(t, org2Found)
		assert.Check(t, org3Found)
	})

	t.Run("support user can read organization avatar", func(t *testing.T) {
		ctx := th.NewSupportCtx(orgUser.Owner.UserCtx, org2ID)

		resp, err := suite.Client.API.GetAllOrganizations(ctx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Organizations.Edges != nil)
		assert.Check(t, is.Len(resp.Organizations.Edges, 1))

		organization := resp.Organizations.Edges[0].Node

		assert.Check(t, is.Equal(org2ID, organization.ID))
		assert.Assert(t, organization.AvatarFile != nil)
		assert.Check(t, is.Equal(avatarFileIDOrg2, organization.AvatarFile.ID))
		assert.Check(t, organization.AvatarFile.PresignedURL != nil)
	})

	// cleanup orgs
	th.CleanupOrganizationDataWithContext(orgUser.Owner.UserCtx, t)
}

func TestMutationCreateOrganization(t *testing.T) {
	t.Parallel()

	orgUser := suite.SeedOrgOwner(t)

	parentOrg, err := suite.Client.API.GetOrganizationByID(orgUser.Owner.UserCtx, orgUser.Owner.OrganizationID)
	assert.NilError(t, err)

	// setup deleted org
	orgToDelete := (&th.OrganizationBuilder{Client: suite.Client}).MustNew(orgUser.Owner.UserCtx, t)
	// delete said org
	(&th.Cleanup[*generated.OrganizationDeleteOne]{Client: suite.Client.DB.Organization, ID: orgToDelete.ID}).MustDelete(orgUser.Owner.UserCtx, t)

	avatarFile := th.UploadFile(t, th.LogoFilePath)
	invalidAvatarFile := th.UploadFile(t, th.TxtFilePath)

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
			client:                   suite.Client.API,
			ctx:                      orgUser.Owner.UserCtx,
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
			client:      suite.Client.API,
			ctx:         orgUser.Owner.UserCtx,
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
			client:      suite.Client.API,
			ctx:         orgUser.Owner.UserCtx,
			errorMsg:    th.InvalidInputErrorMsg,
		},
		{
			name:           "happy path organization with parent org",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    orgUser.Owner.OrganizationID,
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "organization with parent org, no access",
			orgName:        gofakeit.Name(),
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    th.SharedTestUser2.OrganizationID,
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
			errorMsg:       th.NotFoundErrorMsg,
		},
		{
			name:           "organization with parent org using personal access token, not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    orgUser.Owner.OrganizationID,
			client:         orgUser.PatClient,
			ctx:            context.Background(),
			errorMsg:       common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name:           "organization with parent org using personal access token, no access to parent, not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    th.SharedTestUser2.OrganizationID,
			client:         orgUser.PatClient,
			ctx:            context.Background(),
			errorMsg:       common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name:           "organization create with api token not allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			client:         orgUser.APIClient,
			ctx:            context.Background(),
			errorMsg:       common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name:           "organization with parent personal org",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: gofakeit.HipsterSentence(),
			parentOrgID:    orgUser.Owner.PersonalOrgID,
			errorMsg:       "personal organizations are not allowed to have child organizations",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "empty organization name",
			orgName:        "",
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "value is less than the required length",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "long organization name",
			orgName:        gofakeit.LetterN(161),
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "value is greater than the required length",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "organization with no description",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			orgDescription: "",
			parentOrgID:    orgUser.Owner.OrganizationID,
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "duplicate organization name",
			orgName:        parentOrg.Organization.Name,
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "already exists",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "duplicate organization name, case insensitive",
			orgName:        strings.ToUpper(parentOrg.Organization.Name),
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "already exists",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "duplicate organization name, but other was deleted, should pass",
			orgName:        orgToDelete.Name,
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "organization name with trailing space should work with trailing space removed",
			orgName:        "orgname ",
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "invalid organization name, too short",
			orgName:        "a",
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "value is less than the required length",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{

			name:           "invalid organization name with special characters",
			orgName:        "orgn!me$",
			orgDescription: gofakeit.HipsterSentence(),
			errorMsg:       "invalid or unparsable field: name, field cannot contain special characters",
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "duplicate display name, should be allowed",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			displayName:    parentOrg.Organization.DisplayName,
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:           "display name with spaces should pass",
			orgName:        ulids.New().String(), // use ulid to ensure uniqueness
			displayName:    gofakeit.Sentence(),
			orgDescription: gofakeit.HipsterSentence(),
			client:         suite.Client.API,
			ctx:            orgUser.Owner.UserCtx,
		},
		{
			name:       "invalid avatar file",
			orgName:    ulids.New().String(), // use ulid to ensure uniqueness
			avatarFile: invalidAvatarFile,
			client:     suite.Client.API,
			ctx:        orgUser.Owner.UserCtx,
			errorMsg:   "unsupported mime type uploaded: text/plain",
		},
		{
			name:    "invalid allowed email domains ",
			orgName: ulids.New().String(), // use ulid to ensure uniqueness
			settings: &testclient.CreateOrganizationSettingInput{
				AllowedEmailDomains: []string{"theopenlane"},
			},
			client:   suite.Client.API,
			ctx:      orgUser.Owner.UserCtx,
			errorMsg: "invalid or unparsable field: domains",
		},
		{
			name:    "invalid domains",
			orgName: ulids.New().String(), // use ulid to ensure uniqueness
			settings: &testclient.CreateOrganizationSettingInput{
				Domains: []string{"theopenlane"},
			},
			client:   suite.Client.API,
			ctx:      orgUser.Owner.UserCtx,
			errorMsg: "invalid or unparsable field: domains",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.avatarFile})
				} else {
					th.ExpectUploadCheckOnly(t, suite.Client.MockProvider)
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

			resp, err := tc.client.CreateOrganization(tc.ctx, input, tc.avatarFile, nil)

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
				userResp, err := tc.client.GetUserByID(tc.ctx, orgUser.Owner.ID)
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
			newCtx := auth.NewTestContextWithOrgID(orgUser.Owner.ID, resp.CreateOrganization.Organization.ID)

			et, err := suite.Client.API.GetEntityTypes(newCtx, &testclient.EntityTypeWhereInput{
				OwnerID: &resp.CreateOrganization.Organization.ID,
			})
			assert.NilError(t, err)

			assert.Assert(t, is.Len(et.EntityTypes.Edges, 1))

			// ensure managed groups are created
			managedGroups, err := suite.Client.API.GetGroups(newCtx, &testclient.GroupWhereInput{
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
			(&th.Cleanup[*generated.OrganizationDeleteOne]{Client: suite.Client.DB.Organization, ID: resp.CreateOrganization.Organization.ID}).MustDelete(orgUser.Owner.UserCtx, t)
		})
	}
}

func TestMutationUpdateOrganization(t *testing.T) {
	t.Parallel()
	orgUser := suite.UserBuilder(context.Background(), t)

	nameUpdate := ulids.New().String()
	displayNameUpdate := gofakeit.LetterN(40)
	descriptionUpdate := gofakeit.HipsterSentence()
	nameUpdateLong := gofakeit.LetterN(200)

	org := (&th.OrganizationBuilder{Client: suite.Client}).MustNew(orgUser.UserCtx, t)
	user1 := (&th.UserBuilder{Client: suite.Client}).MustNew(orgUser.UserCtx, t)

	reqCtx := auth.NewTestContextWithOrgID(orgUser.ID, org.ID)

	// create groups for creator permissions tests and add a member
	// group created by org owner
	groupProgramCreators := (&th.GroupBuilder{Client: suite.Client}).MustNew(reqCtx, t)
	anotherGroupProgramCreators := (&th.GroupBuilder{Client: suite.Client}).MustNew(reqCtx, t)
	groupProcedureCreators := (&th.GroupBuilder{Client: suite.Client}).MustNew(reqCtx, t)

	(&th.GroupMemberBuilder{Client: suite.Client, GroupID: groupProgramCreators.ID}).MustNew(reqCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, GroupID: anotherGroupProgramCreators.ID}).MustNew(reqCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, GroupID: groupProcedureCreators.ID}).MustNew(reqCtx, t)

	// add a member to the org, to test permissions
	om := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(reqCtx, t)
	memberUserCtx := auth.NewTestContextWithOrgID(om.UserID, org.ID)

	// avatar file setup
	avatarFile := th.UploadFile(t, th.LogoFilePath)
	invalidAvatarFile := th.UploadFile(t, th.TxtFilePath)

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
			name:  "update display name, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				DisplayName: &nameUpdate,
			},
			client: suite.Client.API,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,
				DisplayName: nameUpdate,
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
			client: suite.Client.API,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,
				DisplayName: nameUpdate,
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
			client: suite.Client.API,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,
				DisplayName: nameUpdate,
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
			client: suite.Client.API,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,
				DisplayName: nameUpdate,
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
			client:   suite.Client.API,
			ctx:      memberUserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name:  "update description and avatar file, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				Description: &descriptionUpdate,
			},
			avatarFile: avatarFile,
			client:     suite.Client.API,
			ctx:        reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,
				DisplayName: nameUpdate, // this would have been updated on the prior test
				Description: &descriptionUpdate,
			},
		},
		{
			name:  "update display name, happy path",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				DisplayName: &displayNameUpdate,
			},
			client: suite.Client.API,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,
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
			client: suite.Client.API,
			ctx:    reqCtx,
			expectedRes: testclient.UpdateOrganization_UpdateOrganization_Organization{
				ID:          org.ID,
				Name:        org.Name,          // this would have been updated on the prior test
				DisplayName: displayNameUpdate, // this would have been updated on the prior test
				Description: &descriptionUpdate,
			},
		},
		{
			name:       "update avatar, invalid file",
			orgID:      org.ID,
			avatarFile: invalidAvatarFile,
			client:     suite.Client.API,
			ctx:        reqCtx,
			errorMsg:   "unsupported mime type uploaded: text/plain",
		},
		{
			name:  "update name, too long",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				DisplayName: &nameUpdateLong,
			},
			client:   suite.Client.API,
			ctx:      reqCtx,
			errorMsg: "value is greater than the required length",
		},
		{
			name:  "update name, no access",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				DisplayName: &nameUpdate,
			},
			client:   suite.Client.API,
			ctx:      memberUserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name:  "update name, not found",
			orgID: org.ID,
			updateInput: testclient.UpdateOrganizationInput{
				DisplayName: &nameUpdate,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.avatarFile})
				} else {
					th.ExpectUploadCheckOnly(t, suite.Client.MockProvider)
				}
			}

			resp, err := tc.client.UpdateOrganization(tc.ctx, tc.orgID, tc.updateInput, tc.avatarFile, nil)

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

	(&th.Cleanup[*generated.OrganizationDeleteOne]{Client: suite.Client.DB.Organization, ID: org.ID}).MustDelete(reqCtx, t)
	(&th.Cleanup[*generated.UserDeleteOne]{Client: suite.Client.DB.User, ID: user1.ID}).MustDelete(reqCtx, t)
}

func TestMutationDeleteOrganization(t *testing.T) {
	t.Parallel()
	orgUser := suite.SeedFreshMinimalOrgUsers(t, false)

	reqCtx := orgUser.Owner.UserCtx

	setting, err := suite.Client.API.UpdateUserSetting(reqCtx, orgUser.Owner.UserInfo.Edges.Setting.ID,
		testclient.UpdateUserSettingInput{
			DefaultOrgID: &orgUser.Owner.OrganizationID,
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, orgUser.Owner.OrganizationID, setting.UpdateUserSetting.UserSetting.DefaultOrg.ID)

	testCases := []struct {
		name     string
		orgID    string
		ctx      context.Context
		errorMsg string
	}{
		{
			name:     "delete org, access denied",
			orgID:    orgUser.Owner.OrganizationID,
			ctx:      orgUser.Member.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name:     "delete org, not found",
			orgID:    orgUser.Owner.OrganizationID,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:  "delete org, happy path",
			orgID: orgUser.Owner.OrganizationID,
			ctx:   orgUser.Owner.UserCtx,
		},
		{
			name:     "delete org, personal org not allowed",
			orgID:    orgUser.Owner.PersonalOrgID,
			ctx:      orgUser.Owner.UserCtx,
			errorMsg: "cannot delete personal organizations",
		},
		{
			name:     "delete org, not found",
			orgID:    "tacos-tuesday",
			ctx:      orgUser.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.DeleteOrganization(tc.ctx, tc.orgID)

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
			reqCtx := auth.NewTestContextWithOrgID(orgUser.Owner.ID, orgUser.Owner.OrganizationID)

			// make sure the default org is reset
			settingUpdated, err := suite.Client.API.GetUserSettingByID(reqCtx, orgUser.Owner.UserInfo.Edges.Setting.ID)
			assert.NilError(t, err)
			assert.Assert(t, settingUpdated.UserSetting.DefaultOrg != nil)
			assert.Check(t, orgUser.Owner.OrganizationID != settingUpdated.UserSetting.DefaultOrg.ID)

			// allow ctx to ensure the org no longer exists after deletion
			allowCtx := ent.NewContext(rule.WithInternalContext(reqCtx), suite.Client.DB)

			_, err = suite.Client.API.GetOrganizationByID(allowCtx, tc.orgID)
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)

			// tuples and entity are deleted, so we need to skip soft delete and privacy checks
			ctx := entx.SkipSoftDelete(reqCtx)
			ctx = privacy.DecisionContext(ctx, privacy.Allow)

			o, err := suite.Client.API.GetOrganizationByID(ctx, tc.orgID)
			assert.NilError(t, err)
			assert.Assert(t, o != nil)

			assert.Equal(t, o.Organization.ID, tc.orgID)
		})
	}
}
