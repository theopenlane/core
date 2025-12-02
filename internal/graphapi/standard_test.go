package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/testutils"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryStandard(t *testing.T) {
	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	numControls := 20
	controlIDs := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID}).MustNew(systemAdminUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	notPublicStandard := (&StandardBuilder{client: suite.client, IsPublic: false}).MustNew(systemAdminUser.UserCtx, t)

	orgStandardName := "org-owned-standard"
	orgOwnedStandard := (&StandardBuilder{client: suite.client, Name: orgStandardName}).MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	// add test cases for querying the Standard
	testCases := []struct {
		name                 string
		queryID              string
		expectedControlCount int64
		client               *testclient.TestClient
		ctx                  context.Context
		errorMsg             string
	}{
		{
			name:    "happy path, org owned standard",
			queryID: orgOwnedStandard.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: orgOwnedStandard.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: orgOwnedStandard.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:    "happy path using api token",
			queryID: orgOwnedStandard.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:                 "happy path using api token for public standard",
			queryID:              publicStandard.ID,
			client:               suite.client.apiWithToken,
			ctx:                  context.Background(),
			expectedControlCount: int64(numControls),
		},
		{
			name:     "standard not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "standard not found, using not authorized user",
			queryID:  orgOwnedStandard.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:                 "public standard, other org user",
			queryID:              publicStandard.ID,
			client:               suite.client.api,
			ctx:                  testUser2.UserCtx,
			expectedControlCount: int64(numControls),
		},
		{
			name:                 "public standard, view only user",
			queryID:              publicStandard.ID,
			client:               suite.client.api,
			ctx:                  viewOnlyUser.UserCtx,
			expectedControlCount: int64(numControls),
		},
		{
			name:     "org owned, but not public standard, not found",
			queryID:  notPublicStandard.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "org owned, but not public standard, get by system admin",
			queryID: notPublicStandard.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  orgOwnedStandard.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetStandardByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Standard.ID))
			assert.Check(t, resp.Standard.Name != "")

			if tc.queryID == orgOwnedStandard.ID {
				assert.Check(t, is.Equal(orgStandardName, resp.Standard.Name))
				assert.Check(t, !*resp.Standard.SystemOwned)
			} else {
				assert.Check(t, *resp.Standard.SystemOwned)
			}

			assert.Check(t, resp.Standard.Framework != nil)

			if tc.ctx == systemAdminUser.UserCtx {
				assert.Check(t, resp.Standard.IsPublic != nil)
			} else {
				assert.Check(t, resp.Standard.IsPublic == nil)
			}

			assert.Check(t, is.Equal(tc.expectedControlCount, resp.Standard.Controls.TotalCount))

			// only check edges if we expect them
			if tc.expectedControlCount > 0 {
				assert.Check(t, is.Equal(testutils.MaxResultLimit, len(resp.Standard.Controls.Edges)))
			}

		})
	}

	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{publicStandard.ID, notPublicStandard.ID}}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: orgOwnedStandard.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryStandards(t *testing.T) {
	// create multiple org owned standards
	countOrgOwned := 2
	orgOwnedStandardIDs := []string{}
	for range countOrgOwned {
		standard := (&StandardBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		orgOwnedStandardIDs = append(orgOwnedStandardIDs, standard.ID)
	}

	countPublic := 4
	publicStandardIDs := []string{}
	for range countPublic {
		standard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)
		publicStandardIDs = append(publicStandardIDs, standard.ID)
	}

	countNotPublic := 1
	notPublicStandardIDs := []string{}
	for range countNotPublic {
		standard := (&StandardBuilder{client: suite.client, IsPublic: false}).MustNew(systemAdminUser.UserCtx, t)
		notPublicStandardIDs = append(notPublicStandardIDs, standard.ID)
	}

	// reset count
	countPublic = 0
	countNotPublic = 0

	standards, err := suite.client.api.GetAllStandards(systemAdminUser.UserCtx)
	assert.NilError(t, err)

	for _, standard := range standards.Standards.Edges {
		if *standard.Node.IsPublic {
			countPublic++
			continue
		}

		countNotPublic++
	}

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path, org using should get all org owned + public standards",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "another user, only public should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: countPublic,
		},
		{
			name:            "happy path, system admin user",
			client:          suite.client.api,
			ctx:             systemAdminUser.UserCtx,
			expectedResults: countNotPublic + countPublic,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllStandards(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Standards.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(int64(tc.expectedResults), resp.Standards.TotalCount))

			// under the max results in tests (10), has next should be false
			assert.Check(t, !resp.Standards.PageInfo.HasNextPage)
		})
	}

	systemOwnedIDs := notPublicStandardIDs

	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: systemOwnedIDs}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: orgOwnedStandardIDs}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateStandard(t *testing.T) {
	patClientSystemAdmin := suite.setupPatClient(systemAdminUser, t)

	numControls := 20
	controlIDs := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	numAdminControls := 32
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	createImageUpload := func() *graphql.Upload {
		pdfFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pdfFile.RawFile,
			Filename:    pdfFile.OriginalName,
			Size:        pdfFile.Size,
			ContentType: pdfFile.ContentType,
		}
	}

	testCases := []struct {
		name        string
		request     testclient.CreateStandardInput
		upload      *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateStandardInput{
				Name: "Super Awesome Standard",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, minimal input with logo upload",
			request: testclient.CreateStandardInput{
				Name: "Super Awesome Standard",
			},
			upload: createImageUpload(),
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, system admin - system owned with controls",
			request: testclient.CreateStandardInput{
				Name:       "Super Awesome Standard",
				IsPublic:   lo.ToPtr(true),
				ControlIDs: adminControlIDs,
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "happy path, system admin - system owned using pat",
			request: testclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client: patClientSystemAdmin,
			ctx:    context.Background(),
		},
		{
			name: "happy path, system admin - system owned and public",
			request: testclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "happy path, all input by org admin",
			request: testclient.CreateStandardInput{
				Name:                 "Super Awesome Standard With Everything But Edges",
				Tags:                 []string{"tag1", "tag2"},
				Framework:            lo.ToPtr("Awesome Framework"),
				ShortName:            lo.ToPtr("super-great"),
				Description:          lo.ToPtr("This is a super awesome standard with everything!"),
				GoverningBodyLogoURL: lo.ToPtr("https://example.com/logo.png"),
				GoverningBody:        lo.ToPtr("Super Awesome Governing Body"),
				Domains:              []string{"availability", "meows"},
				Link:                 lo.ToPtr("https://example.com/super-awesome-standard"),
				Status:               &enums.StandardDraft,
				StandardType:         lo.ToPtr("cybersecurity"),
				Version:              lo.ToPtr("2025 - ship latest"),
				Revision:             lo.ToPtr("v1.0.0"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateStandardInput{
				Name:      "Greatness, Kitties, and Rainbows",
				Tags:      []string{"uffo", "brax"},
				Framework: lo.ToPtr("Meows Framework"),
				OwnerID:   &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateStandardInput{
				Name:      "Greatness, Kitties, and Sherbet",
				Tags:      []string{"kc", "eddy"},
				Framework: lo.ToPtr("Meows Framework")},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized to make a public standard",
			request: testclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "user not authorized to make public standard",
			request: testclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "user not authorized to free to use standard",
			request: testclient.CreateStandardInput{
				Name:      "Super Awesome Standard",
				FreeToUse: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateStandardInput{
				Name: "Oh noes",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required field",
			request:     testclient.CreateStandardInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.upload != nil {
				expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.upload})
			}

			resp, err := tc.client.CreateStandard(tc.ctx, tc.request, tc.upload)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, len(resp.CreateStandard.Standard.Name) != 0)

			expectedRevision := "v0.0.1" //default
			if tc.request.Revision != nil {
				expectedRevision = *tc.request.Revision
			}

			assert.Check(t, is.Equal(expectedRevision, *resp.CreateStandard.Standard.Revision))

			expectedStatus := enums.StandardActive
			if tc.request.Status != nil {
				expectedStatus = *tc.request.Status
			}
			assert.Check(t, is.Equal(expectedStatus, *resp.CreateStandard.Standard.Status))

			expectedSystemOwned := false
			if tc.ctx == systemAdminUser.UserCtx || tc.client == patClientSystemAdmin {
				expectedSystemOwned = true
			}
			assert.Check(t, is.Equal(expectedSystemOwned, *resp.CreateStandard.Standard.SystemOwned))

			if tc.ctx == systemAdminUser.UserCtx || tc.client == patClientSystemAdmin {
				isPublic := false
				if tc.request.IsPublic != nil {
					isPublic = *tc.request.IsPublic
				}
				assert.Check(t, is.Equal(isPublic, *resp.CreateStandard.Standard.IsPublic))

				expectedFreeToUse := false
				if tc.request.FreeToUse != nil {
					expectedFreeToUse = *tc.request.FreeToUse
				}
				assert.Check(t, is.Equal(expectedFreeToUse, *resp.CreateStandard.Standard.FreeToUse))
			} else {
				// these are private fields, so they should not be set or returned except to system admins
				assert.Check(t, resp.CreateStandard.Standard.IsPublic == nil)
			}

			expectedTags := []string{}
			if tc.request.Tags != nil {
				expectedTags = tc.request.Tags
			}
			assert.Check(t, is.DeepEqual(expectedTags, resp.CreateStandard.Standard.Tags))

			expectedFramework := ""
			if tc.request.Framework != nil {
				expectedFramework = *tc.request.Framework
			}
			assert.Check(t, is.Equal(expectedFramework, *resp.CreateStandard.Standard.Framework))

			// short name defaults to the name
			expectedShortName := tc.request.Name
			if tc.request.ShortName != nil {
				expectedShortName = *tc.request.ShortName
			}
			assert.Check(t, is.Equal(expectedShortName, *resp.CreateStandard.Standard.ShortName))

			expectedDescription := ""
			if tc.request.Description != nil {
				expectedDescription = *tc.request.Description
			}
			assert.Check(t, is.Equal(expectedDescription, *resp.CreateStandard.Standard.Description))

			expectedGoverningBodyLogoURL := ""
			if tc.request.GoverningBodyLogoURL != nil {
				expectedGoverningBodyLogoURL = *tc.request.GoverningBodyLogoURL
			}
			assert.Check(t, is.Equal(expectedGoverningBodyLogoURL, *resp.CreateStandard.Standard.GoverningBodyLogoURL))

			expectedGoverningBody := ""
			if tc.request.GoverningBody != nil {
				expectedGoverningBody = *tc.request.GoverningBody
			}
			assert.Check(t, is.Equal(expectedGoverningBody, *resp.CreateStandard.Standard.GoverningBody))

			assert.Check(t, is.DeepEqual(tc.request.Domains, resp.CreateStandard.Standard.Domains))

			expectedLink := ""
			if tc.request.Link != nil {
				expectedLink = *tc.request.Link
			}
			assert.Check(t, is.Equal(expectedLink, *resp.CreateStandard.Standard.Link))

			expectedStandardType := ""
			if tc.request.StandardType != nil {
				expectedStandardType = *tc.request.StandardType
			}
			assert.Check(t, is.Equal(expectedStandardType, *resp.CreateStandard.Standard.StandardType))

			expectedVersion := ""
			if tc.request.Version != nil {
				expectedVersion = *tc.request.Version
			}
			assert.Check(t, is.Equal(expectedVersion, *resp.CreateStandard.Standard.Version))

			if tc.upload != nil {
				assert.Assert(t, resp.CreateStandard.Standard.LogoFile != nil)
				assert.Check(t, resp.CreateStandard.Standard.LogoFile.ID != "")
			}

			// cleanup the created standard
			ctx := tc.ctx
			if tc.ctx != systemAdminUser.UserCtx && tc.client != suite.client.api {
				ctx = testUser1.UserCtx
			}

			if tc.client == patClientSystemAdmin {
				ctx = systemAdminUser.UserCtx
			}

			(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: resp.CreateStandard.Standard.ID}).MustDelete(ctx, t)
		})
	}
}

func TestMutationUpdateStandard(t *testing.T) {
	standardOrgOwned := (&StandardBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	standardSystemOwned := (&StandardBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	// users should not be able to get the system owned standard because its not public
	_, err := suite.client.api.GetStandardByID(testUser1.UserCtx, standardSystemOwned.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	createImageUpload := func() *graphql.Upload {
		pdfFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pdfFile.RawFile,
			Filename:    pdfFile.OriginalName,
			Size:        pdfFile.Size,
			ContentType: pdfFile.ContentType,
		}
	}

	testCases := []struct {
		name        string
		id          string
		request     testclient.UpdateStandardInput
		upload      *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field, org owned standard",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				Tags: []string{"new-tag-1", "new-tag-2"},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update field, org owned standard with upload",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				Tags: []string{"new-tag-1", "new-tag-2"},
			},
			upload: createImageUpload(),
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update multiple fields, org owned standard",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				AppendTags:           []string{"new-tag"},
				GoverningBodyLogoURL: lo.ToPtr("https://example.com/new-logo.png"),
				GoverningBody:        lo.ToPtr("Cat Association"),
				ShortName:            lo.ToPtr("super-great"),
				Description:          lo.ToPtr("This is a super awesome standard with everything!"),
				Link:                 lo.ToPtr("https://example.com/super-awesome-standard"),
				Status:               lo.ToPtr(enums.StandardArchived),
				StandardType:         lo.ToPtr("cats"),
				AppendDomains:        []string{"availability", "meows"},
				RevisionBump:         &models.Major,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update not allowed, not enough permissions",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, cannot update public field",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "update not allowed, cannot update public field",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearIsPublic: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "bad request, invalid link",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				Link: lo.ToPtr("not a link"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "update not allowed, no permissions",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "happy path, update field, system owned standard",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				IsPublic: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields, org owned standard",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags:     lo.ToPtr(true),
				AppendDomains: []string{"mice", "meows"},
				Status:        lo.ToPtr(enums.StandardDraft),
				RevisionBump:  &models.Minor,
				FreeToUse:     lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			tc.ctx = resetContext(tc.ctx, t)

			if tc.upload != nil {
				expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.upload})
			}

			resp, err := tc.client.UpdateStandard(tc.ctx, tc.id, tc.request, tc.upload)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.GoverningBodyLogoURL != nil {
				assert.Check(t, is.Equal(*tc.request.GoverningBodyLogoURL, *resp.UpdateStandard.Standard.GoverningBodyLogoURL))
			}

			if tc.request.AppendTags != nil {
				for _, tag := range tc.request.AppendTags {
					assert.Check(t, is.Contains(resp.UpdateStandard.Standard.Tags, tag))
				}

				tagDefs, err := tc.client.GetTagDefinitions(tc.ctx, nil, nil, &testclient.TagDefinitionWhereInput{
					NameIn: tc.request.AppendTags,
				})

				assert.NilError(t, err)
				assert.Check(t, is.Len(tagDefs.TagDefinitions.Edges, len(tc.request.AppendTags)))
			}

			if tc.request.GoverningBody != nil {
				assert.Check(t, is.Equal(*tc.request.GoverningBody, *resp.UpdateStandard.Standard.GoverningBody))
			}

			if tc.request.ShortName != nil {
				assert.Check(t, is.Equal(*tc.request.ShortName, *resp.UpdateStandard.Standard.ShortName))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateStandard.Standard.Description))
			}

			if tc.request.Link != nil {
				assert.Check(t, is.Equal(*tc.request.Link, *resp.UpdateStandard.Standard.Link))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateStandard.Standard.Status))
			}

			if tc.request.StandardType != nil {
				assert.Check(t, is.Equal(*tc.request.StandardType, *resp.UpdateStandard.Standard.StandardType))
			}

			if tc.request.RevisionBump == &models.Major {
				assert.Check(t, is.Equal("v1.0.0", *resp.UpdateStandard.Standard.Revision))
			}

			if tc.request.RevisionBump == &models.Minor {
				assert.Check(t, is.Equal("v0.1.0", *resp.UpdateStandard.Standard.Revision))
			}

			if tc.request.IsPublic != nil && *tc.request.IsPublic {
				assert.Check(t, *resp.UpdateStandard.Standard.IsPublic)

				// users should now be be able to get the system owned standard because its not public
				std, err := suite.client.api.GetStandardByID(testUser1.UserCtx, standardSystemOwned.ID)
				assert.NilError(t, err)
				assert.Assert(t, std != nil)
				assert.Equal(t, standardSystemOwned.ID, std.Standard.ID)
			}

			if tc.upload != nil {
				assert.Assert(t, resp.UpdateStandard.Standard.LogoFile != nil)
				assert.Check(t, resp.UpdateStandard.Standard.LogoFile.ID != "")
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateStandard.Standard.Tags))
			}
		})
	}

	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standardOrgOwned.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standardSystemOwned.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteStandard(t *testing.T) {
	newAdminUser := suite.systemAdminBuilder(context.Background(), t)

	newTestUser1 := suite.userBuilder(context.Background(), t)
	apiClient := suite.setupAPITokenClient(newTestUser1.UserCtx, t)
	patClient := suite.setupPatClient(newTestUser1, t)

	// we need to create the standards each time because the cascade delete of the standard
	standardOrgOwned1 := (&StandardBuilder{client: suite.client}).MustNew(newTestUser1.UserCtx, t)
	standardOrgOwned2 := (&StandardBuilder{client: suite.client}).MustNew(newTestUser1.UserCtx, t)
	standardOrgOwned3 := (&StandardBuilder{client: suite.client}).MustNew(newTestUser1.UserCtx, t)

	standardSystemOwned := (&StandardBuilder{client: suite.client}).MustNew(newAdminUser.UserCtx, t)

	const numberOfControls = 4

	for range numberOfControls {
		(&ControlBuilder{client: suite.client, StandardID: standardSystemOwned.ID}).MustNew(newAdminUser.UserCtx, t)
	}

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(newAdminUser.UserCtx, t)

	for range numberOfControls {
		(&ControlBuilder{client: suite.client, StandardID: publicStandard.ID}).MustNew(newAdminUser.UserCtx, t)
	}

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  standardOrgOwned1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete system owned",
			idToDelete:  standardSystemOwned.ID,
			client:      suite.client.api,
			ctx:         newTestUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete standard",
			idToDelete: standardOrgOwned1.ID,
			client:     suite.client.api,
			ctx:        newTestUser1.UserCtx,
		},
		{
			name:       "happy path, delete system owned",
			idToDelete: standardSystemOwned.ID,
			client:     suite.client.api,
			ctx:        newAdminUser.UserCtx,
		},
		{
			name:        "delete public standard not allowed",
			idToDelete:  publicStandard.ID,
			client:      suite.client.api,
			ctx:         newAdminUser.UserCtx,
			expectedErr: hooks.ErrPublicStandardCannotBeDeleted.Error(),
		},
		{
			name:        "already deleted, not found",
			idToDelete:  standardOrgOwned1.ID,
			client:      suite.client.api,
			ctx:         newTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: standardOrgOwned2.ID,
			client:     patClient,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: standardOrgOwned3.ID,
			client:     apiClient,
			ctx:        context.Background(),
		},
		{
			name:        "already deleted system owned, not found",
			idToDelete:  standardSystemOwned.ID,
			client:      suite.client.api,
			ctx:         newAdminUser.UserCtx,
			expectedErr: "not found",
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         newTestUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteStandard(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteStandard.DeletedID))
		})
	}

	// delete the public standard and the controls linked to it
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: publicStandard.ID}).MustDelete(systemAdminUser.UserCtx, t)
}
