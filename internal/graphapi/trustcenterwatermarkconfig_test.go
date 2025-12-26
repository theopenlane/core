package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestMutationCreateTrustCenterWatermarkConfig(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// delete the auto created watermark config for the trust center
	// so we can test creating a new one
	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	trustCenterWatermarkConfig, err := suite.client.db.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(allowCtx)

	assert.NilError(t, err)
	(&Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{client: suite.client.db.TrustCenterWatermarkConfig, ID: trustCenterWatermarkConfig.ID}).MustDelete(testUser1.UserCtx, t)

	createPNGUpload := func() *graphql.Upload {
		pngFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pngFile.RawFile,
			Filename:    pngFile.OriginalName,
			Size:        pngFile.Size,
			ContentType: pngFile.ContentType,
		}
	}
	testCases := []struct {
		name        string
		input       testclient.CreateTrustCenterWatermarkConfigInput
		logoFile    *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal, text",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Test Text"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, minimal, logo",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
			},
			logoFile: createPNGUpload(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
		},
		{
			name: "happy path, all fields",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
				Text:          lo.ToPtr("Test Text"),
				FontSize:      lo.ToPtr(48.0),
				Opacity:       lo.ToPtr(0.3),
				Rotation:      lo.ToPtr(45.0),
				Color:         lo.ToPtr("#808080"),
				Font:          &enums.FontHelvetica,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "not authorized",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
				Text:          lo.ToPtr("Test Text"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field, trust center id, no trust center found for org",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Test Text"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: "trustCenterID is required",
		},
		{
			name: "missing required field, text or logo",
			input: testclient.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "text_or_logo_id_not_null",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				if tc.expectedErr == "" {
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.logoFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}
			resp, err := tc.client.CreateTrustCenterWatermarkConfig(tc.ctx, tc.input, tc.logoFile)
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

			(&Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{client: suite.client.db.TrustCenterWatermarkConfig, ID: resp.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenterWatermarkConfig(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	watermarkConfig, err := suite.client.db.TrustCenterWatermarkConfig.Query().
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
			name:        "happy path",
			queryID:     watermarkConfig.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "",
		},
		{
			name:        "not found",
			queryID:     "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized",
			queryID:     watermarkConfig.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "anonymous user cannot access trust center watermark config",
			queryID:     watermarkConfig.ID,
			client:      suite.client.api,
			ctx:         createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID),
			expectedErr: notFoundErrorMsg,
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
		})
	}
	(&Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{client: suite.client.db.TrustCenterWatermarkConfig, ID: watermarkConfig.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTrustCenterWatermarkConfig(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	watermarkConfig, err := suite.client.db.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(allowCtx)

	assert.NilError(t, err)

	createPNGUpload := func() *graphql.Upload {
		pngFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pngFile.RawFile,
			Filename:    pngFile.OriginalName,
			Size:        pngFile.Size,
			ContentType: pngFile.ContentType,
		}
	}
	testCases := []struct {
		name        string
		input       testclient.UpdateTrustCenterWatermarkConfigInput
		logoFile    *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update text",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Updated Text"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update logo",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Updated Text"),
			},
			logoFile: createPNGUpload(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
		},
		{
			name: "happy path, update all fields",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text:     lo.ToPtr("Updated Text"),
				FontSize: lo.ToPtr(48.0),
				Opacity:  lo.ToPtr(0.3),
				Rotation: lo.ToPtr(45.0),
				Color:    lo.ToPtr("#808080"),
				Font:     &enums.FontHelvetica,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "not authorized",
			input: testclient.UpdateTrustCenterWatermarkConfigInput{
				Text: lo.ToPtr("Updated Text"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				if tc.expectedErr == "" {
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.logoFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}
			resp, err := tc.client.UpdateTrustCenterWatermarkConfig(tc.ctx, watermarkConfig.ID, tc.input, tc.logoFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(watermarkConfig.ID, resp.UpdateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig.ID))
		})
	}
	(&Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{client: suite.client.db.TrustCenterWatermarkConfig, ID: watermarkConfig.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
