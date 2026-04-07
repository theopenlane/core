package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryPlatform(t *testing.T) {
	// create an platform to be queried using testUser1
	platform := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Platform
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path, same org user",
			queryID: platform.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: platform.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: platform.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "Platform not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "Platform not found, using not authorized user",
			queryID:  platform.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetPlatformByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Platform.ID))
			assert.Check(t, is.Equal(platform.Name, resp.Platform.Name))
		})
	}

	(&Cleanup[*generated.PlatformDeleteOne]{client: suite.client.db.Platform, ID: platform.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryPlatforms(t *testing.T) {
	// create multiple objects to be queried using testUser1
	platform1 := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	platform2 := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path, admin user of the same org",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no Platforms should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllPlatforms(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Platforms.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.PlatformDeleteOne]{client: suite.client.db.Platform, IDs: []string{platform1.ID, platform2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreatePlatform(t *testing.T) {
	asset := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	vendor := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	asset2 := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	pdfFile := uploadFile(t, pdfFilePath)
	pngFile := uploadFile(t, logoFilePath)

	testCases := []struct {
		name                  string
		request               testclient.CreatePlatformInput
		archDiagrams          []*graphql.Upload
		dataFlowDiagrams      []*graphql.Upload
		trustBoundaryDiagrams []*graphql.Upload
		client                *testclient.TestClient
		ctx                   context.Context
		expectedErr           string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreatePlatformInput{
				Name: gofakeit.AppName() + ulids.New().String(),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, admin user edges and files",
			request: testclient.CreatePlatformInput{
				Name:                     gofakeit.AppName() + ulids.New().String(),
				Description:              lo.ToPtr(gofakeit.Paragraph()),
				TrustBoundaryDescription: lo.ToPtr(gofakeit.Paragraph()),
				DataFlowSummary:          lo.ToPtr(gofakeit.Paragraph()),
				ScopeStatement:           lo.ToPtr(gofakeit.Paragraph()),
				ExternalUUID:             lo.ToPtr(uuid.New().String()),
				ExternalReferenceID:      lo.ToPtr("PLT-123"),
				SourceAssetIDs:           []string{asset.ID},
				SourceEntityIDs:          []string{vendor.ID},
				ContainsPii:              lo.ToPtr(true),
			},
			archDiagrams:          []*graphql.Upload{pngFile, pdfFile},
			dataFlowDiagrams:      []*graphql.Upload{pdfFile},
			trustBoundaryDiagrams: []*graphql.Upload{pdfFile},
			client:                suite.client.api,
			ctx:                   adminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreatePlatformInput{
				Name:            gofakeit.AppName() + ulids.New().String(),
				SourceAssetIDs:  []string{asset2.ID},
				SourceEntityIDs: []string{vendor.ID},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreatePlatformInput{
				Name:                     gofakeit.AppName() + ulids.New().String(),
				Description:              lo.ToPtr(gofakeit.Paragraph()),
				TrustBoundaryDescription: lo.ToPtr(gofakeit.Paragraph()),
				DataFlowSummary:          lo.ToPtr(gofakeit.Paragraph()),
				ScopeStatement:           lo.ToPtr(gofakeit.Paragraph()),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreatePlatformInput{
				Name:                     gofakeit.AppName() + ulids.New().String(),
				Description:              lo.ToPtr(gofakeit.Paragraph()),
				TrustBoundaryDescription: lo.ToPtr(gofakeit.Paragraph()),
				DataFlowSummary:          lo.ToPtr(gofakeit.Paragraph()),
				ScopeStatement:           lo.ToPtr(gofakeit.Paragraph()),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name field",
			request:     testclient.CreatePlatformInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if len(tc.archDiagrams) > 0 {
				expectUploadNillable(t, suite.client.mockProvider, tc.archDiagrams)
			}

			if len(tc.dataFlowDiagrams) > 0 {
				expectUploadNillable(t, suite.client.mockProvider, tc.dataFlowDiagrams)
			}

			if len(tc.trustBoundaryDiagrams) > 0 {
				expectUploadNillable(t, suite.client.mockProvider, tc.trustBoundaryDiagrams)
			}

			resp, err := tc.client.CreatePlatform(tc.ctx, tc.request, tc.archDiagrams, tc.dataFlowDiagrams, tc.trustBoundaryDiagrams)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Name, resp.CreatePlatform.Platform.Name))

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreatePlatform.Platform.Description))
			}

			if tc.request.TrustBoundaryDescription != nil {
				assert.Check(t, is.Equal(*tc.request.TrustBoundaryDescription, *resp.CreatePlatform.Platform.TrustBoundaryDescription))
			}

			if tc.request.DataFlowSummary != nil {
				assert.Check(t, is.Equal(*tc.request.DataFlowSummary, *resp.CreatePlatform.Platform.DataFlowSummary))
			}

			if tc.request.ScopeStatement != nil {
				assert.Check(t, is.Equal(*tc.request.ScopeStatement, *resp.CreatePlatform.Platform.ScopeStatement))
			}

			if tc.request.ExternalUUID != nil {
				assert.Check(t, is.Equal(*tc.request.ExternalUUID, *resp.CreatePlatform.Platform.ExternalUUID))
			}

			if tc.request.ExternalReferenceID != nil {
				assert.Check(t, is.Equal(*tc.request.ExternalReferenceID, *resp.CreatePlatform.Platform.ExternalReferenceID))
			}

			if len(tc.request.SourceAssetIDs) > 0 {
				assert.Check(t, is.DeepEqual(tc.request.SourceAssetIDs[0], resp.CreatePlatform.Platform.SourceAssets.Edges[0].Node.ID))
			}

			if len(tc.request.SourceEntityIDs) > 0 {
				assert.Check(t, is.DeepEqual(tc.request.SourceEntityIDs[0], resp.CreatePlatform.Platform.SourceEntities.Edges[0].Node.ID))
			}

			if len(tc.request.OutOfScopeAssetIDs) > 0 {
				assert.Check(t, is.DeepEqual(tc.request.OutOfScopeAssetIDs[0], resp.CreatePlatform.Platform.OutOfScopeAssets.Edges[0].Node.ID))
			}

			if len(tc.request.OutOfScopeVendorIDs) > 0 {
				assert.Check(t, is.DeepEqual(tc.request.OutOfScopeVendorIDs[0], resp.CreatePlatform.Platform.OutOfScopeVendors.Edges[0].Node.ID))
			}

			if tc.request.ContainsPii != nil {
				assert.Check(t, is.Equal(*tc.request.ContainsPii, *resp.CreatePlatform.Platform.ContainsPii))
			}

			if len(tc.archDiagrams) > 0 {
				assert.Check(t, is.Len(resp.CreatePlatform.Platform.ArchitectureDiagrams.Edges, len(tc.archDiagrams)))
			}

			if len(tc.dataFlowDiagrams) > 0 {
				assert.Check(t, is.Len(resp.CreatePlatform.Platform.DataFlowDiagrams.Edges, len(tc.dataFlowDiagrams)))
			}

			if len(tc.trustBoundaryDiagrams) > 0 {
				assert.Check(t, is.Len(resp.CreatePlatform.Platform.TrustBoundaryDiagrams.Edges, len(tc.trustBoundaryDiagrams)))
			}

			// cleanup each object created
			(&Cleanup[*generated.PlatformDeleteOne]{client: suite.client.db.Platform, ID: resp.CreatePlatform.Platform.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup assets and entities created for the tests
	(&Cleanup[*generated.AssetDeleteOne]{client: suite.client.db.Asset, IDs: []string{asset.ID, asset2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: vendor.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdatePlatform(t *testing.T) {
	platform := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	asset := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	vendor := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	outOfScopeAsset := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	outOfScopeVendor := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdatePlatformInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field as admin",
			request: testclient.UpdatePlatformInput{
				Description:            lo.ToPtr(gofakeit.Paragraph()),
				AddEntityIDs:           []string{vendor.ID},
				AddAssetIDs:            []string{asset.ID},
				AddOutOfScopeAssetIDs:  []string{outOfScopeAsset.ID},
				AddOutOfScopeVendorIDs: []string{outOfScopeVendor.ID},
				ExternalReferenceID:    lo.ToPtr("PLT-456"),
				ContainsPii:            lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields with personal access token",
			request: testclient.UpdatePlatformInput{
				Description:              lo.ToPtr(gofakeit.Paragraph()),
				ScopeStatement:           lo.ToPtr(gofakeit.Paragraph()),
				DataFlowSummary:          lo.ToPtr(gofakeit.Paragraph()),
				TrustBoundaryDescription: lo.ToPtr(gofakeit.Paragraph()),
				ExternalReferenceID:      lo.ToPtr("PLT-456"),
				ContainsPii:              lo.ToPtr(false),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdatePlatformInput{
				Name: lo.ToPtr("New Name"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdatePlatformInput{
				Name: lo.ToPtr("New Name"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdatePlatform(tc.ctx, platform.ID, tc.request, nil, nil, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdatePlatform.Platform.Description))
			}

			if tc.request.ScopeStatement != nil {
				assert.Check(t, is.Equal(*tc.request.ScopeStatement, *resp.UpdatePlatform.Platform.ScopeStatement))
			}

			if tc.request.DataFlowSummary != nil {
				assert.Check(t, is.Equal(*tc.request.DataFlowSummary, *resp.UpdatePlatform.Platform.DataFlowSummary))
			}

			if tc.request.TrustBoundaryDescription != nil {
				assert.Check(t, is.Equal(*tc.request.TrustBoundaryDescription, *resp.UpdatePlatform.Platform.TrustBoundaryDescription))
			}

			if tc.request.ContainsPii != nil {
				assert.Check(t, is.Equal(*tc.request.ContainsPii, *resp.UpdatePlatform.Platform.ContainsPii))
			}

			if len(tc.request.AddOutOfScopeAssetIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdatePlatform.Platform.OutOfScopeAssets.Edges, len(tc.request.AddOutOfScopeAssetIDs)))
			}

			if len(tc.request.AddOutOfScopeVendorIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdatePlatform.Platform.OutOfScopeVendors.Edges, len(tc.request.AddOutOfScopeVendorIDs)))
			}

			if len(tc.request.AddAssetIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdatePlatform.Platform.Assets.Edges, len(tc.request.AddAssetIDs)))
			}

			if len(tc.request.AddEntityIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdatePlatform.Platform.Entities.Edges, len(tc.request.AddEntityIDs)))
			}
		})
	}

	(&Cleanup[*generated.PlatformDeleteOne]{client: suite.client.db.Platform, ID: platform.ID}).MustDelete(testUser1.UserCtx, t)
	// cleanup assets and entities created for the tests
	(&Cleanup[*generated.AssetDeleteOne]{client: suite.client.db.Asset, IDs: []string{asset.ID, outOfScopeAsset.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: []string{vendor.ID, outOfScopeVendor.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeletePlatform(t *testing.T) {
	// create objects to be deleted
	platform1 := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	platform2 := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	platform3 := (&PlatformBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  platform1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  platform1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: platform1.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  platform1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: platform2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: platform3.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeletePlatform(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeletePlatform.DeletedID))
		})
	}
}
