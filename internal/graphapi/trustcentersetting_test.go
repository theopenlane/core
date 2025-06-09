package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryTrustCenterSettingByID(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: trustCenter.Edges.Setting.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: trustCenter.Edges.Setting.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:     "trust center setting not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not authorized to query org",
			queryID:  trustCenter.Edges.Setting.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterSettingByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterSetting.ID))
			assert.Check(t, resp.TrustCenterSetting.Title != nil)
			assert.Check(t, resp.TrustCenterSetting.OwnerID != nil)
			assert.Check(t, is.Equal(testUser1.OrganizationID, *resp.TrustCenterSetting.OwnerID))
			assert.Check(t, is.Equal(trustCenter.ID, resp.TrustCenterSetting.TrustCenterID))
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTrustCenterSetting(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                 string
		trustCenterSettingID string
		request              openlaneclient.UpdateTrustCenterSettingInput
		client               *openlaneclient.OpenlaneClient
		ctx                  context.Context
		expectedErr          string
	}{
		{
			name:                 "happy path, update title",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Updated Trust Center Title"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:                 "happy path, update overview",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Overview: lo.ToPtr("Updated overview for the trust center"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:                 "happy path, update logo URL",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				LogoURL: lo.ToPtr("https://example.com/new-logo.png"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:                 "happy path, update favicon URL",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				FaviconURL: lo.ToPtr("https://example.com/new-favicon.ico"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:                 "happy path, update primary color",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				PrimaryColor: lo.ToPtr("#FF5733"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:                 "happy path, using admin user",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Admin Updated Title"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:                 "happy path, using personal access token",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("PAT Updated Title"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:                 "not authorized, view only user",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Unauthorized Update"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:                 "not authorized, different org user",
			trustCenterSettingID: trustCenter.Edges.Setting.ID,
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Unauthorized Update"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:                 "trust center setting not found",
			trustCenterSettingID: "non-existent-id",
			request: openlaneclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Test Update"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTrustCenterSetting(tc.ctx, tc.trustCenterSettingID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.trustCenterSettingID, resp.UpdateTrustCenterSetting.TrustCenterSetting.ID))

			// Check updated fields
			if tc.request.Title != nil {
				assert.Check(t, is.Equal(*tc.request.Title, *resp.UpdateTrustCenterSetting.TrustCenterSetting.Title))
			}

			if tc.request.Overview != nil {
				assert.Check(t, is.Equal(*tc.request.Overview, *resp.UpdateTrustCenterSetting.TrustCenterSetting.Overview))
			}

			if tc.request.LogoURL != nil {
				assert.Check(t, is.Equal(*tc.request.LogoURL, *resp.UpdateTrustCenterSetting.TrustCenterSetting.LogoURL))
			}

			if tc.request.FaviconURL != nil {
				assert.Check(t, is.Equal(*tc.request.FaviconURL, *resp.UpdateTrustCenterSetting.TrustCenterSetting.FaviconURL))
			}

			if tc.request.PrimaryColor != nil {
				assert.Check(t, is.Equal(*tc.request.PrimaryColor, *resp.UpdateTrustCenterSetting.TrustCenterSetting.PrimaryColor))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateTrustCenterSetting.TrustCenterSetting.Tags))
			}
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenterSettings(t *testing.T) {
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int64
		where           *openlaneclient.TrustCenterSettingWhereInput
	}{
		{
			name:            "return all for user1",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "return all, view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 1,
		},
		{
			name:   "query by owner ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterSettingWhereInput{
				OwnerID: &testUser1.OrganizationID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by trust center ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterSettingWhereInput{
				TrustCenterID: &trustCenter1.ID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by non-existent trust center ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterSettingWhereInput{
				TrustCenterID: lo.ToPtr("non-existent-id"),
			},
			expectedResults: 0,
		},
		{
			name:            "different user sees only their settings",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterSettings(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenterSettings.TotalCount))

			// Verify that users only see trust center settings from their organization
			if tc.ctx == testUser1.UserCtx || tc.ctx == viewOnlyUser.UserCtx {
				for _, edge := range resp.TrustCenterSettings.Edges {
					assert.Check(t, is.Equal(testUser1.OrganizationID, *edge.Node.OwnerID))
				}
			} else if tc.ctx == testUser2.UserCtx {
				for _, edge := range resp.TrustCenterSettings.Edges {
					assert.Check(t, is.Equal(testUser2.OrganizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, IDs: []string{trustCenter1.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}
