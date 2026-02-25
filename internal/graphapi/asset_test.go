package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryAsset(t *testing.T) {
	asset := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: asset.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: asset.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: asset.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "asset not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "asset not found, using not authorized user",
			queryID:  asset.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAssetByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Asset.ID))
			assert.Check(t, resp.Asset.Name != "")

		})
	}

	(&Cleanup[*generated.AssetDeleteOne]{client: suite.client.db.Asset, ID: asset.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryAssets(t *testing.T) {
	asset1 := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	asset2 := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
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
			name:            "another user, no assets should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllAssets(tc.ctx, nil, nil, nil, nil, []*testclient.AssetOrder{})
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Assets.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.AssetDeleteOne]{client: suite.client.db.Asset, IDs: []string{asset1.ID, asset2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateAsset(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateAssetInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateAssetInput{
				Name: "theopenlane.io",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input as org admin",
			request: testclient.CreateAssetInput{
				Name:                "theopenlane.io",
				Description:         lo.ToPtr("description"),
				InternalOwnerUserID: &viewOnlyUser.ID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateAssetInput{
				Name:                "theopenlane.io",
				Description:         lo.ToPtr("description"),
				InternalOwnerUserID: &viewOnlyUser.ID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateAssetInput{
				Name:        "theopenlane.io",
				Description: lo.ToPtr("description"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateAssetInput{
				Name: "comply.fyi",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required field",
			request:     testclient.CreateAssetInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateAsset(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.CreateAsset.Asset.ID != "")
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateAsset.Asset.Name))

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateAsset.Asset.Description))
			} else {
				assert.Check(t, *resp.CreateAsset.Asset.Description == "", "expected Description to be nil or empty but was %v", *resp.CreateAsset.Asset.Description)
			}

			if tc.request.InternalOwnerUserID != nil {
				assert.Check(t, is.Equal(*tc.request.InternalOwnerUserID, *resp.CreateAsset.Asset.InternalOwnerUserID))
			} else {
				assert.Check(t, *resp.CreateAsset.Asset.InternalOwnerUserID == "", "expected InternalOwnerUserID to be nil but was %v", resp.CreateAsset.Asset.InternalOwnerUserID)
			}

			if tc.request.InternalOwnerGroupID != nil {
				assert.Check(t, is.Equal(*tc.request.InternalOwnerGroupID, *resp.CreateAsset.Asset.InternalOwnerGroupID))
			} else {
				assert.Check(t, *resp.CreateAsset.Asset.InternalOwnerGroupID == "", "expected InternalOwnerGroupID to be nil but was %v", resp.CreateAsset.Asset.InternalOwnerGroupID)
			}

			(&Cleanup[*generated.AssetDeleteOne]{client: suite.client.db.Asset, ID: resp.CreateAsset.Asset.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateAsset(t *testing.T) {
	asset := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateAssetInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field by admin user",
			request: testclient.UpdateAssetInput{
				Description: lo.ToPtr("updated description"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateAssetInput{
				InternalOwnerUserID: &adminUser.ID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions as view only user",
			request: testclient.UpdateAssetInput{
				InternalOwnerUserID: &viewOnlyUser.ID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, not allowed to add edge to without access to group",
			request: testclient.UpdateAssetInput{
				InternalOwnerGroupID: &viewOnlyUser2.GroupID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update allowed to add edge to group if user has access to group",
			request: testclient.UpdateAssetInput{
				InternalOwnerGroupID: &testUser1.GroupID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateAssetInput{
				Description: lo.ToPtr("updated description again"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateAsset(tc.ctx, asset.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateAsset.Asset.Description))
			}

			if tc.request.InternalOwnerUserID != nil {
				assert.Check(t, is.Equal(*tc.request.InternalOwnerUserID, *resp.UpdateAsset.Asset.InternalOwnerUserID))
			}
		})
	}

	(&Cleanup[*generated.AssetDeleteOne]{client: suite.client.db.Asset, ID: asset.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteAsset(t *testing.T) {
	asset1 := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	asset2 := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	asset3 := (&AssetBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  asset1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  asset1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: asset1.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  asset1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: asset2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: asset3.ID,
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
			resp, err := tc.client.DeleteAsset(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteAsset.DeletedID))
		})
	}
}
