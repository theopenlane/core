package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryStandard() {
	t := suite.T()

	// create standards to be queried
	publicStandard := (&StandardBuilder{client: suite.client, SystemOwned: true, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)
	notPublicStandard := (&StandardBuilder{client: suite.client, SystemOwned: true, IsPublic: false}).MustNew(systemAdminUser.UserCtx, t)

	orgStandardName := "org-owned-standard"
	orgOwnedStandard := (&StandardBuilder{client: suite.client, Name: orgStandardName}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Standard
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
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
			name:    "happy path using api token for public standard",
			queryID: publicStandard.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
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
			name:    "org owned public standard",
			queryID: publicStandard.ID,
			client:  suite.client.api,
			ctx:     testUser2.UserCtx,
		},
		{
			name:    "org owned public standard",
			queryID: publicStandard.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
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
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetStandardByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Standard)

			require.Equal(t, tc.queryID, resp.Standard.ID)
			require.NotEmpty(t, resp.Standard.Name)

			if tc.queryID == orgOwnedStandard.ID {
				assert.Equal(t, orgStandardName, resp.Standard.Name)
			}

			assert.NotEmpty(t, resp.Standard.Framework)

			if tc.queryID == publicStandard.ID {
				assert.True(t, *resp.Standard.IsPublic)
			} else {
				assert.False(t, *resp.Standard.IsPublic)
			}

			if tc.queryID != orgOwnedStandard.ID {
				assert.True(t, *resp.Standard.SystemOwned)
			} else {
				assert.False(t, *resp.Standard.SystemOwned)
			}
		})
	}

	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{publicStandard.ID, notPublicStandard.ID}}).MustDelete(systemAdminUser.UserCtx, suite)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: orgOwnedStandard.ID}).MustDelete(testUser1.UserCtx, suite)
}

func (suite *GraphTestSuite) TestQueryStandards() {
	t := suite.T()

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
		standard := (&StandardBuilder{client: suite.client, SystemOwned: true, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)
		publicStandardIDs = append(publicStandardIDs, standard.ID)
	}

	countNotPublic := 1
	notPublicStandardIDs := []string{}
	for range countNotPublic {
		standard := (&StandardBuilder{client: suite.client, SystemOwned: true, IsPublic: false}).MustNew(systemAdminUser.UserCtx, t)
		notPublicStandardIDs = append(notPublicStandardIDs, standard.ID)
	}

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
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
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Standards.Edges, tc.expectedResults)
			assert.Equal(t, int64(tc.expectedResults), resp.Standards.TotalCount)

			// under the max results in tests (10), has next should be false
			assert.False(t, resp.Standards.PageInfo.HasNextPage)
		})
	}

	systemOwnedIDs := append(notPublicStandardIDs, publicStandardIDs...)

	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: systemOwnedIDs}).MustDelete(systemAdminUser.UserCtx, suite)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: orgOwnedStandardIDs}).MustDelete(testUser1.UserCtx, suite)
}

func (suite *GraphTestSuite) TestMutationCreateStandard() {
	t := suite.T()

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

	testCases := []struct {
		name        string
		request     openlaneclient.CreateStandardInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateStandardInput{
				Name: "Super Awesome Standard",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, system admin - system owned with controls",
			request: openlaneclient.CreateStandardInput{
				Name:        "Super Awesome Standard",
				SystemOwned: lo.ToPtr(true),
				ControlIDs:  adminControlIDs,
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "happy path, system admin - system owned and public",
			request: openlaneclient.CreateStandardInput{
				Name:        "Super Awesome Standard",
				SystemOwned: lo.ToPtr(true),
				IsPublic:    lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "happy path, all input by org admin",
			request: openlaneclient.CreateStandardInput{
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
				ControlIDs:           controlIDs,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateStandardInput{
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
			request: openlaneclient.CreateStandardInput{
				Name:      "Greatness, Kitties, and Sherbet",
				Tags:      []string{"kc", "eddy"},
				Framework: lo.ToPtr("Meows Framework")},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized to make system owned standard",
			request: openlaneclient.CreateStandardInput{
				Name:        "Super Awesome Standard",
				SystemOwned: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized to make public standard",
			request: openlaneclient.CreateStandardInput{
				Name:     "Super Awesome Standard",
				IsPublic: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized to free to use standard",
			request: openlaneclient.CreateStandardInput{
				Name:      "Super Awesome Standard",
				FreeToUse: lo.ToPtr(true),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateStandardInput{
				Name: "Oh noes",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required field",
			request:     openlaneclient.CreateStandardInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateStandard(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.NotEmpty(t, resp.CreateStandard.Standard.Name)

			expectedRevision := "v0.0.1" //default
			if tc.request.Revision != nil {
				expectedRevision = *tc.request.Revision
			}

			assert.Equal(t, expectedRevision, *resp.CreateStandard.Standard.Revision)

			expectedStatus := enums.StandardActive
			if tc.request.Status != nil {
				expectedStatus = *tc.request.Status
			}
			assert.Equal(t, expectedStatus, *resp.CreateStandard.Standard.Status)

			expectedSystemOwned := false
			if tc.request.SystemOwned != nil {
				expectedSystemOwned = *tc.request.SystemOwned
			}
			assert.Equal(t, expectedSystemOwned, *resp.CreateStandard.Standard.SystemOwned)

			expectedIsPublic := false
			if tc.request.IsPublic != nil {
				expectedIsPublic = *tc.request.IsPublic
			}
			assert.Equal(t, expectedIsPublic, *resp.CreateStandard.Standard.IsPublic)

			// this field isn't currently used to enforce anything, it may change to restrict
			// usage on tiers + features
			expectedFreeToUse := false
			if tc.request.FreeToUse != nil {
				expectedFreeToUse = *tc.request.FreeToUse
			}
			assert.Equal(t, expectedFreeToUse, *resp.CreateStandard.Standard.FreeToUse)

			expectedTags := []string{}
			if tc.request.Tags != nil {
				expectedTags = tc.request.Tags
			}
			assert.Equal(t, expectedTags, resp.CreateStandard.Standard.Tags)

			expectedFramework := ""
			if tc.request.Framework != nil {
				expectedFramework = *tc.request.Framework
			}
			assert.Equal(t, expectedFramework, *resp.CreateStandard.Standard.Framework)

			expectedShortName := ""
			if tc.request.ShortName != nil {
				expectedShortName = *tc.request.ShortName
			}
			assert.Equal(t, expectedShortName, *resp.CreateStandard.Standard.ShortName)

			expectedDescription := ""
			if tc.request.Description != nil {
				expectedDescription = *tc.request.Description
			}
			assert.Equal(t, expectedDescription, *resp.CreateStandard.Standard.Description)

			expectedGoverningBodyLogoURL := ""
			if tc.request.GoverningBodyLogoURL != nil {
				expectedGoverningBodyLogoURL = *tc.request.GoverningBodyLogoURL
			}
			assert.Equal(t, expectedGoverningBodyLogoURL, *resp.CreateStandard.Standard.GoverningBodyLogoURL)

			expectedGoverningBody := ""
			if tc.request.GoverningBody != nil {
				expectedGoverningBody = *tc.request.GoverningBody
			}
			assert.Equal(t, expectedGoverningBody, *resp.CreateStandard.Standard.GoverningBody)

			assert.Equal(t, tc.request.Domains, resp.CreateStandard.Standard.Domains)

			expectedLink := ""
			if tc.request.Link != nil {
				expectedLink = *tc.request.Link
			}
			assert.Equal(t, expectedLink, *resp.CreateStandard.Standard.Link)

			expectedStandardType := ""
			if tc.request.StandardType != nil {
				expectedStandardType = *tc.request.StandardType
			}
			assert.Equal(t, expectedStandardType, *resp.CreateStandard.Standard.StandardType)

			expectedVersion := ""
			if tc.request.Version != nil {
				expectedVersion = *tc.request.Version
			}
			assert.Equal(t, expectedVersion, *resp.CreateStandard.Standard.Version)

			if len(tc.request.ControlIDs) > 0 {
				assert.NotEmpty(t, resp.CreateStandard.Standard.Controls.Edges)
				assert.Equal(t, int64(len(tc.request.ControlIDs)), resp.CreateStandard.Standard.Controls.TotalCount)
				// created more than the max limit for tests (10)
				assert.True(t, resp.CreateStandard.Standard.Controls.PageInfo.HasNextPage)
			} else {
				assert.Empty(t, resp.CreateStandard.Standard.Controls.Edges)
			}

			// cleanup the created standard
			ctx := tc.ctx
			if tc.client != suite.client.api {
				ctx = testUser1.UserCtx
			}

			(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: resp.CreateStandard.Standard.ID}).MustDelete(ctx, suite)
		})
	}

	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser1.UserCtx, suite)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: adminControlIDs}).MustDelete(systemAdminUser.UserCtx, suite)
}

func (suite *GraphTestSuite) TestMutationUpdateStandard() {
	t := suite.T()

	numControls := 8
	controlIDs := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	additionalControls := 4
	additionalControlIDs := []string{}
	for range additionalControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
		additionalControlIDs = append(additionalControlIDs, control.ID)
	}

	// standardOrgOwned := (&StandardBuilder{client: suite.client, ControlIDs: controlIDs}).MustNew(testUser1.UserCtx, t)

	numAdminControls := 32
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	standardSystemOwned := (&StandardBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		id          string
		request     openlaneclient.UpdateStandardInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		// {
		// 	name: "happy path, update field, org owned standard",
		// 	id:   standardOrgOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		AddControlIDs: additionalControlIDs,
		// 	},
		// 	client: suite.client.api,
		// 	ctx:    adminUser.UserCtx,
		// },
		// {
		// 	name: "happy path, update multiple fields, org owned standard",
		// 	id:   standardOrgOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		AppendTags:           []string{"new-tag"},
		// 		GoverningBodyLogoURL: lo.ToPtr("https://example.com/new-logo.png"),
		// 		GoverningBody:        lo.ToPtr("Cat Association"),
		// 		ShortName:            lo.ToPtr("super-great"),
		// 		Description:          lo.ToPtr("This is a super awesome standard with everything!"),
		// 		Link:                 lo.ToPtr("https://example.com/super-awesome-standard"),
		// 		Status:               lo.ToPtr(enums.StandardArchived),
		// 		StandardType:         lo.ToPtr("cats"),
		// 		AppendDomains:        []string{"availability", "meows"},
		// 		RevisionBump:         &models.Major,
		// 	},
		// 	client: suite.client.apiWithPAT,
		// 	ctx:    context.Background(),
		// },
		// {
		// 	name: "update not allowed, not enough permissions",
		// 	id:   standardOrgOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		ClearTags: lo.ToPtr(true),
		// 	},
		// 	client:      suite.client.api,
		// 	ctx:         viewOnlyUser.UserCtx,
		// 	expectedErr: notAuthorizedErrorMsg,
		// },
		// {
		// 	name: "update not allowed, no permissions",
		// 	id:   standardOrgOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		ClearTags: lo.ToPtr(true),
		// 	},
		// 	client:      suite.client.api,
		// 	ctx:         testUser2.UserCtx,
		// 	expectedErr: notFoundErrorMsg,
		// },
		{
			name: "happy path, update field, system owned standard",
			id:   standardSystemOwned.ID,
			request: openlaneclient.UpdateStandardInput{
				IsPublic: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		// {
		// 	name: "happy path, update multiple fields, org owned standard",
		// 	id:   standardSystemOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		ClearTags:     lo.ToPtr(true),
		// 		AppendDomains: []string{"mice", "meows"},
		// 		Status:        lo.ToPtr(enums.StandardDraft),
		// 		RevisionBump:  &models.Minor,
		// 	},
		// 	client: suite.client.api,
		// 	ctx:    systemAdminUser.UserCtx,
		// },
		// {
		// 	name: "update not allowed, no permissions",
		// 	id:   standardSystemOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		ClearTags: lo.ToPtr(true),
		// 	},
		// 	client:      suite.client.api,
		// 	ctx:         testUser1.UserCtx,
		// 	expectedErr: notAuthorizedErrorMsg,
		// },
		// {
		// 	name: "update not allowed, no permissions",
		// 	id:   standardSystemOwned.ID,
		// 	request: openlaneclient.UpdateStandardInput{
		// 		ClearTags: lo.ToPtr(true),
		// 	},
		// 	client:      suite.client.api,
		// 	ctx:         testUser2.UserCtx,
		// 	expectedErr: notFoundErrorMsg,
		// },
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateStandard(tc.ctx, tc.id, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.GoverningBodyLogoURL != nil {
				assert.Equal(t, *tc.request.GoverningBodyLogoURL, *resp.UpdateStandard.Standard.GoverningBodyLogoURL)
			}

			if len(tc.request.AddControlIDs) > 0 {
				assert.Equal(t, len(controlIDs)+len(additionalControlIDs), int(resp.UpdateStandard.Standard.Controls.TotalCount))
			}

			if tc.request.AppendTags != nil {
				assert.Contains(t, resp.UpdateStandard.Standard.Tags, "new-tag")
			}

			if tc.request.GoverningBody != nil {
				assert.Equal(t, *tc.request.GoverningBody, *resp.UpdateStandard.Standard.GoverningBody)
			}

			if tc.request.ShortName != nil {
				assert.Equal(t, *tc.request.ShortName, *resp.UpdateStandard.Standard.ShortName)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateStandard.Standard.Description)
			}

			if tc.request.Link != nil {
				assert.Equal(t, *tc.request.Link, *resp.UpdateStandard.Standard.Link)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateStandard.Standard.Status)
			}

			if tc.request.StandardType != nil {
				assert.Equal(t, *tc.request.StandardType, *resp.UpdateStandard.Standard.StandardType)
			}

			if tc.request.RevisionBump == &models.Major {
				assert.Equal(t, "v1.0.0", *resp.UpdateStandard.Standard.Revision)
			}
		})
	}

	// (&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standardOrgOwned.ID}).MustDelete(testUser1.UserCtx, suite)
	// (&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser1.UserCtx, suite)
	// (&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: adminControlIDs}).MustDelete(systemAdminUser.UserCtx, suite)
}

// func (suite *GraphTestSuite) TestMutationDeleteStandard() {
// 	t := suite.T()

// 	// create Standards to be deleted
// 	Standard1 := (&StandardBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
// 	Standard2 := (&StandardBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

// 	testCases := []struct {
// 		name        string
// 		idToDelete  string
// 		client      *openlaneclient.OpenlaneClient
// 		ctx         context.Context
// 		expectedErr string
// 	}{
// 		{
// 			name:        "not authorized, delete",
// 			idToDelete:  Standard1.ID,
// 			client:      suite.client.api,
// 			ctx:         testUser2.UserCtx,
// 			expectedErr: notFoundErrorMsg,
// 		},
// 		{
// 			name:       "happy path, delete",
// 			idToDelete: Standard1.ID,
// 			client:     suite.client.api,
// 			ctx:        testUser1.UserCtx,
// 		},
// 		{
// 			name:        "already deleted, not found",
// 			idToDelete:  Standard1.ID,
// 			client:      suite.client.api,
// 			ctx:         testUser1.UserCtx,
// 			expectedErr: "not found",
// 		},
// 		{
// 			name:       "happy path, delete using personal access token",
// 			idToDelete: Standard2.ID,
// 			client:     suite.client.apiWithPAT,
// 			ctx:        context.Background(),
// 		},
// 		{
// 			name:        "unknown id, not found",
// 			idToDelete:  ulids.New().String(),
// 			client:      suite.client.api,
// 			ctx:         testUser1.UserCtx,
// 			expectedErr: notFoundErrorMsg,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("Delete "+tc.name, func(t *testing.T) {
// 			resp, err := tc.client.DeleteStandard(tc.ctx, tc.idToDelete)
// 			if tc.expectedErr != "" {
// 				require.Error(t, err)
// 				assert.ErrorContains(t, err, tc.expectedErr)
// 				assert.Nil(t, resp)

// 				return
// 			}

// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 			assert.Equal(t, tc.idToDelete, resp.DeleteStandard.DeletedID)
// 		})
// 	}
// }
