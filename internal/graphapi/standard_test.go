package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/testutils"
)

func TestQueryStandard(t *testing.T) {
	publicStandard := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	numControls := 20
	controlIDs := []string{}
	for range numControls {
		control := (&th.ControlBuilder{Client: suite.Client, StandardID: publicStandard.ID}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	notPublicStandard := (&th.StandardBuilder{Client: suite.Client, IsPublic: false}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	orgStandardName := "org-owned-standard"
	orgOwnedStandard := (&th.StandardBuilder{Client: suite.Client, Name: orgStandardName}).MustNew(th.SharedTestUser1.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

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
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: orgOwnedStandard.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: orgOwnedStandard.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:    "happy path using api token",
			queryID: orgOwnedStandard.ID,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
		},
		{
			name:                 "happy path using api token for public standard",
			queryID:              publicStandard.ID,
			client:               suite.Client.APIWithToken,
			ctx:                  context.Background(),
			expectedControlCount: int64(numControls),
		},
		{
			name:     "standard not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "standard not found, using not authorized user",
			queryID:  orgOwnedStandard.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:                 "public standard, other org user",
			queryID:              publicStandard.ID,
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser2.UserCtx,
			expectedControlCount: int64(numControls),
		},
		{
			name:                 "public standard, view only user",
			queryID:              publicStandard.ID,
			client:               suite.Client.API,
			ctx:                  th.SharedViewOnlyUser.UserCtx,
			expectedControlCount: int64(numControls),
		},
		{
			name:     "org owned, but not public standard, not found",
			queryID:  notPublicStandard.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:    "org owned, but not public standard, get by system admin",
			queryID: notPublicStandard.ID,
			client:  suite.Client.API,
			ctx:     th.SharedSystemAdminUser.UserCtx,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  orgOwnedStandard.ID,
			errorMsg: th.NotFoundErrorMsg,
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

			if tc.ctx == th.SharedSystemAdminUser.UserCtx {
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

	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: controlIDs}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, IDs: []string{publicStandard.ID, notPublicStandard.ID}}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, ID: orgOwnedStandard.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryStandards(t *testing.T) {
	// create multiple org owned standards
	countOrgOwned := 2
	orgOwnedStandardIDs := []string{}
	for range countOrgOwned {
		standard := (&th.StandardBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
		orgOwnedStandardIDs = append(orgOwnedStandardIDs, standard.ID)
	}

	countPublic := 4
	publicStandardIDs := []string{}
	for range countPublic {
		standard := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
		publicStandardIDs = append(publicStandardIDs, standard.ID)
	}

	countNotPublic := 1
	notPublicStandardIDs := []string{}
	for range countNotPublic {
		standard := (&th.StandardBuilder{Client: suite.Client, IsPublic: false}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
		notPublicStandardIDs = append(notPublicStandardIDs, standard.ID)
	}

	// reset count
	countPublic = 0
	countNotPublic = 0

	standards, err := suite.Client.API.GetAllStandards(th.SharedSystemAdminUser.UserCtx)
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
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: countOrgOwned + countPublic,
		},
		{
			name:            "another user, only public should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: countPublic,
		},
		{
			name:            "happy path, system admin user",
			client:          suite.Client.API,
			ctx:             th.SharedSystemAdminUser.UserCtx,
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

	systemOwnedIDs := append(notPublicStandardIDs, publicStandardIDs...)

	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, IDs: systemOwnedIDs}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, IDs: orgOwnedStandardIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryStandardsWithDeletedControls(t *testing.T) {
	standard1 := (&th.StandardBuilder{Client: suite.Client, IsPublic: true, Name: "Standard With Active Controls"}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	standard2 := (&th.StandardBuilder{Client: suite.Client, IsPublic: true, Name: "Standard With Deleted Controls"}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	standard3 := (&th.StandardBuilder{Client: suite.Client, IsPublic: true, Name: "Standard With No Controls"}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	control1 := (&th.ControlBuilder{Client: suite.Client, StandardID: standard1.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	control2 := (&th.ControlBuilder{Client: suite.Client, StandardID: standard1.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	controlToDelete1 := (&th.ControlBuilder{Client: suite.Client, StandardID: standard2.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	controlToDelete2 := (&th.ControlBuilder{Client: suite.Client, StandardID: standard2.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	whereFilter := &testclient.StandardWhereInput{
		HasControlsWith: []*testclient.ControlWhereInput{
			{
				HasOwnerWith: []*testclient.OrganizationWhereInput{
					{
						ID: &th.SharedTestUser1.OrganizationID,
					},
				},
			},
		},
	}

	// check to make sure there are 2 standards since we only linked to two standards
	resp, err := suite.Client.API.GetStandards(th.SharedTestUser1.UserCtx, nil, nil, whereFilter)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	assert.Check(t, is.Len(resp.Standards.Edges, 2))

	// delete the controls linked to standard2
	for _, id := range []string{controlToDelete1.ID, controlToDelete2.ID} {
		_, err := suite.Client.API.DeleteControl(th.SharedTestUser1.UserCtx, id)
		assert.NilError(t, err)
	}

	resp, err = suite.Client.API.GetStandards(th.SharedTestUser1.UserCtx, nil, nil, whereFilter)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	assert.Check(t, is.Len(resp.Standards.Edges, 1))
	assert.Check(t, is.Equal(standard1.ID, resp.Standards.Edges[0].Node.ID), "expected standard1 only")

	// cleanup
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control1.ID, control2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, IDs: []string{standard1.ID, standard2.ID, standard3.ID}}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestMutationCreateStandard(t *testing.T) {
	patClientSystemAdmin := suite.SetupPatClient(th.SharedSystemAdminUser, t)

	numControls := 20
	controlIDs := []string{}
	for range numControls {
		control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	numAdminControls := 32
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	createImageUpload := th.LogoFileFunc(t)

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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, minimal input with logo upload",
			request: testclient.CreateStandardInput{
				Name: "Super Awesome Standard",
			},
			upload: createImageUpload(),
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, system admin - system owned with controls",
			request: testclient.CreateStandardInput{
				Name:       "Super Awesome Standard",
				IsPublic:   lo.ToPtr(true),
				ControlIDs: adminControlIDs,
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
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
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
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
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateStandardInput{
				Name:      "Greatness, Kitties, and Rainbows",
				Tags:      []string{"uffo", "brax"},
				Framework: lo.ToPtr("Meows Framework"),
				OwnerID:   &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateStandardInput{
				Name:      "Greatness, Kitties, and Sherbet",
				Tags:      []string{"kc", "eddy"},
				Framework: lo.ToPtr("Meows Framework")},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized to make a public standard",
			request: testclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
		{
			name: "user not authorized to make public standard",
			request: testclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
		{
			name: "user not authorized to free to use standard",
			request: testclient.CreateStandardInput{
				Name:      "Super Awesome Standard",
				FreeToUse: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateStandardInput{
				Name: "Oh noes",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "missing required field",
			request:     testclient.CreateStandardInput{},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.upload != nil {
				th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.upload})
			}

			resp, err := tc.client.CreateStandard(tc.ctx, tc.request, tc.upload, nil)
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
			if tc.ctx == th.SharedSystemAdminUser.UserCtx || tc.client == patClientSystemAdmin {
				expectedSystemOwned = true
			}
			assert.Check(t, is.Equal(expectedSystemOwned, *resp.CreateStandard.Standard.SystemOwned))

			if tc.ctx == th.SharedSystemAdminUser.UserCtx || tc.client == patClientSystemAdmin {
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
			if tc.ctx != th.SharedSystemAdminUser.UserCtx && tc.client != suite.Client.API {
				ctx = th.SharedTestUser1.UserCtx
			}

			if tc.client == patClientSystemAdmin {
				ctx = th.SharedSystemAdminUser.UserCtx
			}

			(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, ID: resp.CreateStandard.Standard.ID}).MustDelete(ctx, t)
		})
	}
}

func TestMutationUpdateStandard(t *testing.T) {
	standardOrgOwned := (&th.StandardBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	standardSystemOwned := (&th.StandardBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	// users should not be able to get the system owned standard because its not public
	_, err := suite.Client.API.GetStandardByID(th.SharedTestUser1.UserCtx, standardSystemOwned.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	createImageUpload := th.LogoFileFunc(t)

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
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update field, org owned standard with upload",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				Tags: []string{"new-tag-1", "new-tag-2"},
			},
			upload: createImageUpload(),
			client: suite.Client.APIWithPAT,
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
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "update not allowed, not enough permissions",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, cannot update public field",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
		{
			name: "update not allowed, cannot update public field",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearIsPublic: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
		{
			name: "bad request, invalid link",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				Link: lo.ToPtr("not a link"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "update not allowed, no permissions",
			id:   standardOrgOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "happy path, update field, system owned standard",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				IsPublic: lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
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
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			id:   standardSystemOwned.ID,
			request: testclient.UpdateStandardInput{
				ClearTags: lo.ToPtr(true),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			tc.ctx = th.ResetContext(tc.ctx, t)

			if tc.upload != nil {
				th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.upload})
			}

			resp, err := tc.client.UpdateStandard(tc.ctx, tc.id, tc.request, tc.upload, nil)
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
				std, err := suite.Client.API.GetStandardByID(th.SharedTestUser1.UserCtx, standardSystemOwned.ID)
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

	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, ID: standardOrgOwned.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, ID: standardSystemOwned.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestMutationDeleteStandard(t *testing.T) {
	t.Parallel()

	newAdminUser := suite.SystemAdminBuilder(context.Background(), t)

	localTestOrg := suite.SeedOrgOwner(t)

	newTestUser1 := localTestOrg.Owner
	apiClient := localTestOrg.APIClient
	patClient := localTestOrg.PatClient

	// we need to create the standards each time because the cascade delete of the standard
	standardOrgOwned1 := (&th.StandardBuilder{Client: suite.Client}).MustNew(newTestUser1.UserCtx, t)
	standardOrgOwned2 := (&th.StandardBuilder{Client: suite.Client}).MustNew(newTestUser1.UserCtx, t)
	standardOrgOwned3 := (&th.StandardBuilder{Client: suite.Client}).MustNew(newTestUser1.UserCtx, t)

	standardSystemOwned := (&th.StandardBuilder{Client: suite.Client}).MustNew(newAdminUser.UserCtx, t)

	const numberOfControls = 4

	for range numberOfControls {
		(&th.ControlBuilder{Client: suite.Client, StandardID: standardSystemOwned.ID}).MustNew(newAdminUser.UserCtx, t)
	}

	publicStandard := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(newAdminUser.UserCtx, t)

	for range numberOfControls {
		(&th.ControlBuilder{Client: suite.Client, StandardID: publicStandard.ID}).MustNew(newAdminUser.UserCtx, t)
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
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, delete system owned",
			idToDelete:  standardSystemOwned.ID,
			client:      suite.Client.API,
			ctx:         newTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete standard",
			idToDelete: standardOrgOwned1.ID,
			client:     suite.Client.API,
			ctx:        newTestUser1.UserCtx,
		},
		{
			name:       "happy path, delete system owned",
			idToDelete: standardSystemOwned.ID,
			client:     suite.Client.API,
			ctx:        newAdminUser.UserCtx,
		},
		{
			name:        "delete public standard not allowed",
			idToDelete:  publicStandard.ID,
			client:      suite.Client.API,
			ctx:         newAdminUser.UserCtx,
			expectedErr: hooks.ErrPublicStandardCannotBeDeleted.Error(),
		},
		{
			name:        "already deleted, not found",
			idToDelete:  standardOrgOwned1.ID,
			client:      suite.Client.API,
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
			client:      suite.Client.API,
			ctx:         newAdminUser.UserCtx,
			expectedErr: "not found",
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         newTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, ID: publicStandard.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}
