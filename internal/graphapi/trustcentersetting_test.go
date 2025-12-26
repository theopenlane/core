package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/common/enums"
)

// TestCreateTrustCenterSetting tests the createTrustCenterSetting mutation
// Note: Trust center settings are created automatically when a trust center is created (both live and preview).
// This test verifies that we can create a deleted setting again after deletion.
func TestCreateTrustCenterSetting(t *testing.T) {
	// Test 1: happy path - recreate a deleted live setting
	t.Run("Create happy path - recreate deleted live setting", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.client.api.DeleteTrustCenterSetting(testUser1.UserCtx, settingID)
		assert.NilError(t, err)

		// Recreate the setting
		resp, err := suite.client.api.CreateTrustCenterSetting(testUser1.UserCtx, testclient.CreateTrustCenterSettingInput{
			TrustCenterID: &trustCenter.ID,
			Title:         lo.ToPtr("Test Setting"),
			Overview:      lo.ToPtr("Test Overview"),
			PrimaryColor:  lo.ToPtr("#FF0000"),
			Environment:   lo.ToPtr(enums.TrustCenterEnvironmentLive),
		})

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.CreateTrustCenterSetting.TrustCenterSetting.ID != "")
		assert.Check(t, is.Equal(trustCenter.ID, *resp.CreateTrustCenterSetting.TrustCenterSetting.TrustCenterID))
		assert.Check(t, is.Equal("Test Setting", *resp.CreateTrustCenterSetting.TrustCenterSetting.Title))
		assert.Check(t, is.Equal("#FF0000", *resp.CreateTrustCenterSetting.TrustCenterSetting.PrimaryColor))

		// Clean up
		(&Cleanup[*generated.TrustCenterSettingDeleteOne]{
			client: suite.client.db.TrustCenterSetting,
			ID:     resp.CreateTrustCenterSetting.TrustCenterSetting.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// Test 2: happy path - recreate with all color fields
	t.Run("Create happy path - recreate with all color fields", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.client.api.DeleteTrustCenterSetting(testUser1.UserCtx, settingID)
		assert.NilError(t, err)

		// Recreate with all color fields
		resp, err := suite.client.api.CreateTrustCenterSetting(testUser1.UserCtx, testclient.CreateTrustCenterSettingInput{
			TrustCenterID:            &trustCenter.ID,
			Title:                    lo.ToPtr("Full Color Setting"),
			PrimaryColor:             lo.ToPtr("#FF0000"),
			ForegroundColor:          lo.ToPtr("#000000"),
			BackgroundColor:          lo.ToPtr("#FFFFFF"),
			AccentColor:              lo.ToPtr("#0000FF"),
			SecondaryBackgroundColor: lo.ToPtr("#F0F0F0"),
			SecondaryForegroundColor: lo.ToPtr("#333333"),
			Environment:              lo.ToPtr(enums.TrustCenterEnvironmentLive),
		})

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("#FF0000", *resp.CreateTrustCenterSetting.TrustCenterSetting.PrimaryColor))
		assert.Check(t, is.Equal("#000000", *resp.CreateTrustCenterSetting.TrustCenterSetting.ForegroundColor))

		// Clean up
		(&Cleanup[*generated.TrustCenterSettingDeleteOne]{
			client: suite.client.db.TrustCenterSetting,
			ID:     resp.CreateTrustCenterSetting.TrustCenterSetting.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// Test 3: happy path - recreate with theme mode
	t.Run("Create happy path - recreate with theme mode", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.client.api.DeleteTrustCenterSetting(testUser1.UserCtx, settingID)
		assert.NilError(t, err)

		// Recreate with theme mode
		resp, err := suite.client.api.CreateTrustCenterSetting(testUser1.UserCtx, testclient.CreateTrustCenterSettingInput{
			TrustCenterID: &trustCenter.ID,
			Title:         lo.ToPtr("Theme Setting"),
			ThemeMode:     lo.ToPtr(enums.TrustCenterThemeModeAdvanced),
			Font:          lo.ToPtr("Arial, sans-serif"),
			Environment:   lo.ToPtr(enums.TrustCenterEnvironmentLive),
		})

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(enums.TrustCenterThemeModeAdvanced, *resp.CreateTrustCenterSetting.TrustCenterSetting.ThemeMode))

		// Clean up
		(&Cleanup[*generated.TrustCenterSettingDeleteOne]{
			client: suite.client.db.TrustCenterSetting,
			ID:     resp.CreateTrustCenterSetting.TrustCenterSetting.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// Test 4: not authorized - view only user cannot create
	t.Run("Create not authorized - view only user", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.client.api.DeleteTrustCenterSetting(testUser1.UserCtx, settingID)
		assert.NilError(t, err)

		// Try to recreate as view only user
		_, err = suite.client.api.CreateTrustCenterSetting(viewOnlyUser.UserCtx, testclient.CreateTrustCenterSettingInput{
			TrustCenterID: &trustCenter.ID,
			Title:         lo.ToPtr("Unauthorized"),
			Environment:   lo.ToPtr(enums.TrustCenterEnvironmentLive),
		})

		assert.ErrorContains(t, err, notAuthorizedErrorMsg)

		// Clean up
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})
}

// TestQueryTrustCenterSetting tests the trustCenterSetting query
func TestQueryTrustCenterSetting(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		settingID   string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:      "happy path - query trust center setting by ID",
			settingID: trustCenter.Edges.Setting.ID,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name:        "trust center setting not found",
			settingID:   "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Query "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterSettingByID(tc.ctx, tc.settingID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.settingID, resp.TrustCenterSetting.ID))
			assert.Check(t, resp.TrustCenterSetting.TrustCenterID != nil)
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

// TestUpdateTrustCenterSetting tests the updateTrustCenterSetting mutation
func TestUpdateTrustCenterSetting(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		settingID   string
		input       testclient.UpdateTrustCenterSettingInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:      "happy path - update title",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Updated Title"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update multiple fields",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title:           lo.ToPtr("New Title"),
				Overview:        lo.ToPtr("New Overview"),
				PrimaryColor:    lo.ToPtr("#00FF00"),
				ForegroundColor: lo.ToPtr("#111111"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update theme mode",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeEasy),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "not authorized - view only user",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Unauthorized"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "not authorized - different org user",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Unauthorized"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "trust center setting not found",
			settingID: "non-existent-id",
			input: testclient.UpdateTrustCenterSettingInput{
				Title: lo.ToPtr("Not Found"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTrustCenterSetting(tc.ctx, tc.settingID, tc.input, nil, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.settingID, resp.UpdateTrustCenterSetting.TrustCenterSetting.ID))

			if tc.input.Title != nil {
				assert.Check(t, is.Equal(*tc.input.Title, *resp.UpdateTrustCenterSetting.TrustCenterSetting.Title))
			}

			if tc.input.PrimaryColor != nil {
				assert.Check(t, is.Equal(*tc.input.PrimaryColor, *resp.UpdateTrustCenterSetting.TrustCenterSetting.PrimaryColor))
			}

			if tc.input.ThemeMode != nil {
				assert.Check(t, is.Equal(*tc.input.ThemeMode, *resp.UpdateTrustCenterSetting.TrustCenterSetting.ThemeMode))
			}
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

// TestDeleteTrustCenterSetting tests the deleteTrustCenterSetting mutation
func TestDeleteTrustCenterSetting(t *testing.T) {
	// Test 1: happy path - delete trust center setting
	t.Run("Delete happy path - delete trust center setting", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		resp, err := suite.client.api.DeleteTrustCenterSetting(testUser1.UserCtx, settingID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(settingID, resp.DeleteTrustCenterSetting.DeletedID))

		// Verify the setting is deleted
		_, err = suite.client.api.GetTrustCenterSettingByID(testUser1.UserCtx, settingID)
		assert.ErrorContains(t, err, notFoundErrorMsg)

		// Clean up
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// Test 2: not authorized - view only user
	t.Run("Delete not authorized - view only user", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		_, err := suite.client.api.DeleteTrustCenterSetting(viewOnlyUser.UserCtx, settingID)
		assert.ErrorContains(t, err, notAuthorizedErrorMsg)

		// Clean up
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// Test 3: not authorized - different org user
	t.Run("Delete not authorized - different org user", func(t *testing.T) {
		trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		settingID := trustCenter.Edges.Setting.ID

		_, err := suite.client.api.DeleteTrustCenterSetting(testUser2.UserCtx, settingID)
		assert.ErrorContains(t, err, notAuthorizedErrorMsg)

		// Clean up
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// Test 4: trust center setting not found
	t.Run("Delete trust center setting not found", func(t *testing.T) {
		_, err := suite.client.api.DeleteTrustCenterSetting(testUser1.UserCtx, "non-existent-id")
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})
}
