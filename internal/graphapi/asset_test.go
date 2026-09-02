package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryAsset(t *testing.T) {
	asset := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: asset.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: asset.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "asset not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "asset not found, using not authorized user",
			queryID:  asset.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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

	(&th.Cleanup[*generated.AssetDeleteOne]{Client: suite.Client.DB.Asset, ID: asset.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryAssets(t *testing.T) {
	asset1 := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	asset2 := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no assets should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
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

	(&th.Cleanup[*generated.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: []string{asset1.ID, asset2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input as org admin",
			request: testclient.CreateAssetInput{
				Name:                "theopenlane.io",
				Description:         lo.ToPtr("description"),
				InternalOwnerUserID: &th.SharedViewOnlyUser.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateAssetInput{
				Name:                "theopenlane.io",
				Description:         lo.ToPtr("description"),
				InternalOwnerUserID: &th.SharedViewOnlyUser.ID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateAssetInput{
				Name:        "theopenlane.io",
				Description: lo.ToPtr("description"),
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateAssetInput{
				Name: "comply.fyi",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "missing required field",
			request:     testclient.CreateAssetInput{},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
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

			(&th.Cleanup[*generated.AssetDeleteOne]{Client: suite.Client.DB.Asset, ID: resp.CreateAsset.Asset.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateAsset(t *testing.T) {
	asset := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateAssetInput{
				InternalOwnerUserID: &th.SharedAdminUser.ID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions as view only user",
			request: testclient.UpdateAssetInput{
				InternalOwnerUserID: &th.SharedViewOnlyUser.ID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, not allowed to add edge to without access to group",
			request: testclient.UpdateAssetInput{
				InternalOwnerGroupID: &th.SharedViewOnlyUser2.GroupID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update allowed to add edge to group if user has access to group",
			request: testclient.UpdateAssetInput{
				InternalOwnerGroupID: &th.SharedTestUser1.GroupID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateAssetInput{
				Description: lo.ToPtr("updated description again"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

	(&th.Cleanup[*generated.AssetDeleteOne]{Client: suite.Client.DB.Asset, ID: asset.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteAsset(t *testing.T) {
	asset1 := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	asset2 := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	asset3 := (&th.AssetBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  asset1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: asset1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedAdminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  asset1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: asset2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: asset3.ID,
			client:     suite.Client.APIWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
