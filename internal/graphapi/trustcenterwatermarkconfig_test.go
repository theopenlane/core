package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestMutationCreateTrustCenterWatermarkConfig(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	// delete the auto created watermark config for the trust center
	// so we can test creating a new one
	allowCtx := privacy.DecisionContext(tcOrg.Owner.UserCtx, privacy.Allow)
	trustCenterWatermarkConfig, err := suite.Client.DB.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(allowCtx)

	assert.NilError(t, err)
	(&th.Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{Client: suite.Client.DB.TrustCenterWatermarkConfig, ID: trustCenterWatermarkConfig.ID}).MustDelete(tcOrg.Owner.UserCtx, t)

	createImageUpload := th.LogoFileFunc(t)
	testCases := []struct {
		name          string
		input         testclient.CreateTrustCenterWatermarkConfigInput
		watermarkFile *graphql.Upload
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal, text",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Test Text"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name: "happy path, minimal, logo",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
			},
			watermarkFile: createImageUpload(),
			client:        suite.Client.API,
			ctx:           tcOrg.SuperAdmin.UserCtx,
		},
		{
			name: "happy path, all fields as admin",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
				Text:          lo.ToPtr("Test Text"),
				FontSize:      lo.ToPtr(48.0),
				Opacity:       lo.ToPtr(0.3),
				Rotation:      lo.ToPtr(45.0),
				Color:         lo.ToPtr("#808080"),
				Font:          &enums.FontHelvetica,
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "not authorized",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
				Text:          lo.ToPtr("Test Text"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "missing required field, trust center id, no trust center found for org",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Test Text"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: "trustCenterID is required",
		},
		{
			name: "missing required field, text or logo",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: "text_or_logo_id_not_null",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.watermarkFile != nil {
				if tc.expectedErr == "" {
					th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.watermarkFile})
				} else {
					th.ExpectUploadCheckOnly(t, suite.Client.MockProvider)
				}
			}
			resp, err := tc.client.CreateTrustCenterWatermarkConfig(tc.ctx, tc.input, tc.watermarkFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.ID != "")

			if tc.input.TrustCenterID != nil {
				assert.Check(t, is.Equal(*tc.input.TrustCenterID, *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.TrustCenterID))
			}

			if tc.input.Text != nil {
				assert.Check(t, *tc.input.Text == *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.Text)
			}

			if tc.input.FontSize != nil {
				assert.Check(t, *tc.input.FontSize == *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.FontSize)
			}

			if tc.input.Opacity != nil {
				assert.Check(t, *tc.input.Opacity == *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.Opacity)
			}

			if tc.input.Rotation != nil {
				assert.Check(t, *tc.input.Rotation == *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.Rotation)
			}

			if tc.input.Color != nil {
				assert.Check(t, *tc.input.Color == *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.Color)
			}

			if tc.input.Font != nil {
				assert.Check(t, *tc.input.Font == *resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.Font)
			}

			(&th.Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{Client: suite.Client.DB.TrustCenterWatermarkConfig, ID: resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.ID}).MustDelete(tc.ctx, t)
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenterWatermarkConfig(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	allowCtx := privacy.DecisionContext(tcOrg.Owner.UserCtx, privacy.Allow)
	watermarkConfig, err := suite.Client.DB.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(allowCtx)

	assert.NilError(t, err)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path",
			queryID: watermarkConfig.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Admin.UserCtx,
		},
		{
			name:    "happy path by system admin",
			queryID: watermarkConfig.ID,
			client:  suite.Client.API,
			ctx:     th.SharedSystemAdminUser.UserCtx,
		},
		{
			name:        "not found",
			queryID:     "non-existent-id",
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized",
			queryID:     watermarkConfig.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "anonymous user cannot access trust center watermark config",
			queryID:     watermarkConfig.ID,
			client:      suite.Client.API,
			ctx:         th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.OrganizationID),
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterWatermarkConfigByID(tc.ctx, tc.queryID)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterWatermarkConfig.ID))

			// check the list as well
			resp2, err := tc.client.GetTrustCenterWatermarkConfigs(tc.ctx, nil, nil, &testclient.TrustCenterWatermarkConfigWhereInput{
				TrustCenterID: &trustCenter.ID,
			})
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.expectedErr != "" {
				assert.Check(t, is.Len(resp2.TrustCenterWatermarkConfigs.Edges, 0))

				return
			}

			assert.Check(t, is.Len(resp2.TrustCenterWatermarkConfigs.Edges, 1))
			assert.Check(t, is.Equal(tc.queryID, resp2.TrustCenterWatermarkConfigs.Edges[0].Node.ID))
		})
	}
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterWatermarkConfig(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	allowCtx := privacy.DecisionContext(tcOrg.Owner.UserCtx, privacy.Allow)
	watermarkConfig, err := suite.Client.DB.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(allowCtx)

	assert.NilError(t, err)

	createImageUpload := th.LogoFileFunc(t)
	testCases := []struct {
		name          string
		input         testclient.UpdateTrustCenterWatermarkConfigInput
		watermarkFile *graphql.Upload
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, update text as admin",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Updated Text"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "happy path, update logo",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Updated Text"),
			},
			watermarkFile: createImageUpload(),
			client:        suite.Client.API,
			ctx:           tcOrg.Owner.UserCtx,
		},
		{
			name: "happy path, update all fields as admin",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text:     lo.ToPtr("Updated Text"),
				FontSize: lo.ToPtr(48.0),
				Opacity:  lo.ToPtr(0.3),
				Rotation: lo.ToPtr(45.0),
				Color:    lo.ToPtr("#808080"),
				Font:     &enums.FontHelvetica,
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "not authorized",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Updated Text"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.watermarkFile != nil {
				if tc.expectedErr == "" {
					th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.watermarkFile})
				} else {
					th.ExpectUploadCheckOnly(t, suite.Client.MockProvider)
				}
			}
			resp, err := tc.client.UpdateTrustCenterWatermarkConfig(tc.ctx, watermarkConfig.ID, tc.input, tc.watermarkFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(watermarkConfig.ID, resp.UpdateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.ID))
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
