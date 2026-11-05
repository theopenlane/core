package graphapi_test

import (
	"context"
	"strings"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

// TestCreateTrustCenterSetting tests the createTrustCenterSetting mutation
// Note: Trust center settings are created automatically when a trust center is created (both live and preview).
// This test verifies that we can create a deleted setting again after deletion.
func TestCreateTrustCenterSetting(t *testing.T) {
	t.Parallel()
	// Test 1: happy path - recreate a deleted live setting
	t.Run("Create happy path - recreate deleted live setting", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.Client.API.DeleteTrustCenterSetting(tcOrg.Owner.UserCtx, settingID)
		assert.NilError(t, err)

		// Recreate the setting
		resp, err := suite.Client.API.CreateTrustCenterSetting(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSettingInput{
			TrustCenterID: &tcOrg.TrustCenter.ID,
			Title:         new("Test Setting"),
			Overview:      new("Test Overview"),
			PrimaryColor:  new("#FF0000"),
			Environment:   new(enums.TrustCenterEnvironmentLive),
		}, nil, nil, nil, nil)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.CreateTrustCenterSetting.TrustCenterSetting.ID != "")
		assert.Check(t, is.Equal(tcOrg.TrustCenter.ID, *resp.CreateTrustCenterSetting.TrustCenterSetting.TrustCenterID))
		assert.Check(t, is.Equal("Test Setting", *resp.CreateTrustCenterSetting.TrustCenterSetting.Title))
		assert.Check(t, is.Equal("#FF0000", *resp.CreateTrustCenterSetting.TrustCenterSetting.PrimaryColor))

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})

	// Test 2: happy path - recreate with all color fields
	t.Run("Create happy path - recreate with all color fields", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.Client.API.DeleteTrustCenterSetting(tcOrg.Owner.UserCtx, settingID)
		assert.NilError(t, err)

		// Recreate with all color fields
		resp, err := suite.Client.API.CreateTrustCenterSetting(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSettingInput{
			Title:                    new("Full Color Setting"),
			PrimaryColor:             new("#FF0000"),
			ForegroundColor:          new("#000000"),
			BackgroundColor:          new("#FFFFFF"),
			AccentColor:              new("#0000FF"),
			SecondaryBackgroundColor: new("#F0F0F0"),
			SecondaryForegroundColor: new("#333333"),
			Environment:              new(enums.TrustCenterEnvironmentLive),
		}, nil, nil, nil, nil)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("#FF0000", *resp.CreateTrustCenterSetting.TrustCenterSetting.PrimaryColor))
		assert.Check(t, is.Equal("#000000", *resp.CreateTrustCenterSetting.TrustCenterSetting.ForegroundColor))

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})

	// Test 3: happy path - recreate with theme mode
	t.Run("Create happy path - recreate with theme mode", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.Client.API.DeleteTrustCenterSetting(tcOrg.Owner.UserCtx, settingID)
		assert.NilError(t, err)

		// Recreate with theme mode
		resp, err := suite.Client.API.CreateTrustCenterSetting(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSettingInput{
			Title:       new("Theme Setting"),
			ThemeMode:   new(enums.TrustCenterThemeModeAdvanced),
			Font:        new("Arial, sans-serif"),
			Environment: new(enums.TrustCenterEnvironmentLive),
		}, nil, nil, nil, nil)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(enums.TrustCenterThemeModeAdvanced, *resp.CreateTrustCenterSetting.TrustCenterSetting.ThemeMode))

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})

	// Test 4: not authorized - view only user cannot create
	t.Run("Create not authorized - view only user", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		// Delete the live setting
		_, err := suite.Client.API.DeleteTrustCenterSetting(tcOrg.Owner.UserCtx, settingID)
		assert.NilError(t, err)

		// Try to recreate as view only user
		_, err = suite.Client.API.CreateTrustCenterSetting(tcOrg.Member.UserCtx, testclient.CreateTrustCenterSettingInput{
			Title:       new("Unauthorized"),
			Environment: new(enums.TrustCenterEnvironmentLive),
		}, nil, nil, nil, nil)

		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})
}

// TestQueryTrustCenterSetting tests the trustCenterSetting query
func TestQueryTrustCenterSetting(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

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
			client:    suite.Client.API,
			ctx:       tcOrg.SuperAdmin.UserCtx,
		},
		{
			name:      "happy path - query trust center preview setting by ID",
			settingID: trustCenter.Edges.PreviewSetting.ID,
			client:    suite.Client.API,
			ctx:       tcOrg.SuperAdmin.UserCtx,
		},
		{
			name:        "trust center setting not found",
			settingID:   "non-existent-id",
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

// TestUpdateTrustCenterSetting tests the updateTrustCenterSetting mutation
func TestUpdateTrustCenterSetting(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	testCases := []struct {
		name        string
		settingID   string
		input       testclient.UpdateTrustCenterSettingInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
		expectJob   bool
	}{
		{
			name:      "happy path - update title by admin user",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: new("Updated Title"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name:      "happy path - update title of preview setting",
			settingID: trustCenter.Edges.PreviewSetting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: new("Updated Title Preview"),
			},
			client:    suite.Client.API,
			ctx:       tcOrg.SuperAdmin.UserCtx,
			expectJob: true, // updating preview setting should enqueue a job to create preview domain
		},
		{
			name:      "happy path - update multiple fields",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title:           new("New Title"),
				Overview:        new("New Overview"),
				PrimaryColor:    new("#00FF00"),
				ForegroundColor: new("#111111"),
				SecurityContact: new("Security@example.com"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:      "happy path - update theme mode",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: new(enums.TrustCenterThemeModeEasy),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:      "not authorized - view only user",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: new("Unauthorized"),
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:      "not authorized - different org user",
			settingID: trustCenter.Edges.Setting.ID,
			input: testclient.UpdateTrustCenterSettingInput{
				Title: new("Unauthorized"),
			},
			client:      suite.Client.API,
			ctx:         tcOrg2.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:      "trust center setting not found",
			settingID: "non-existent-id",
			input: testclient.UpdateTrustCenterSettingInput{
				Title: new("Not Found"),
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.UpdateTrustCenterSetting(tc.ctx, tc.settingID, tc.input, nil, nil, nil, nil, nil, nil)
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

			if tc.input.SecurityContact != nil {
				assert.Check(t, is.Equal(strings.ToLower(*tc.input.SecurityContact), *resp.UpdateTrustCenterSetting.TrustCenterSetting.SecurityContact))
			}

			if tc.expectJob {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobspec.CreatePreviewDomainArgs{
								TrustCenterID:            *resp.UpdateTrustCenterSetting.TrustCenterSetting.TrustCenterID,
								TrustCenterPreviewZoneID: th.MappableDomainZoneTestID,
								TrustCenterCnameTarget:   th.CnameTargetTest,
							},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.CreatePreviewDomainArgs{}, nil)
			}
		})
	}

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

// TestUpdateTrustCenterSettingSupportUser verifies that an org-scoped support session
// (auth.NewOrgSupportCaller, CapOrgSupport) can update trust center branding on behalf
// of the org it is scoped to, the same way a fully-scoped API token can
func TestUpdateTrustCenterSettingSupportUser(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithSupportUser())
	trustCenter := tcOrg.TrustCenter
	settingID := trustCenter.Edges.Setting.ID

	resp, err := suite.Client.API.UpdateTrustCenterSetting(tcOrg.SupportCtx, settingID, testclient.UpdateTrustCenterSettingInput{
		Title:        new("Support Updated Branding"),
		PrimaryColor: new("#123456"),
	}, nil, nil, nil, nil, nil, nil)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(settingID, resp.UpdateTrustCenterSetting.TrustCenterSetting.ID))
	assert.Check(t, is.Equal("Support Updated Branding", *resp.UpdateTrustCenterSetting.TrustCenterSetting.Title))
	assert.Check(t, is.Equal("#123456", *resp.UpdateTrustCenterSetting.TrustCenterSetting.PrimaryColor))

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

// TestSubprocessorNotifyWatermarkInitialized verifies enabling notify-on-subprocessor-change stamps
// the notification watermark so subscribers are only notified about changes made after opting in,
// not the trust center's pre-existing subprocessor list
func TestSubprocessorNotifyWatermarkInitialized(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	settingID := tcOrg.TrustCenter.Edges.Setting.ID

	dbCtx := th.SetInternalContext(tcOrg.Owner.UserCtx, suite.Client.DB)

	setting, err := suite.Client.DB.TrustCenterSetting.Get(dbCtx, settingID)
	assert.NilError(t, err)
	assert.Check(t, setting.SubprocessorsNotifiedAt == nil)

	// enabling the flag stamps the watermark
	_, err = suite.Client.API.UpdateTrustCenterSetting(tcOrg.Owner.UserCtx, settingID, testclient.UpdateTrustCenterSettingInput{
		NotifySubscribersOnSubprocessorChange: new(true),
	}, nil, nil, nil, nil, nil, nil)
	assert.NilError(t, err)

	enabled, err := suite.Client.DB.TrustCenterSetting.Get(dbCtx, settingID)
	assert.NilError(t, err)
	assert.Assert(t, enabled.SubprocessorsNotifiedAt != nil)

	stamped := *enabled.SubprocessorsNotifiedAt

	// updating the setting with the flag already on leaves the watermark alone
	_, err = suite.Client.API.UpdateTrustCenterSetting(tcOrg.Owner.UserCtx, settingID, testclient.UpdateTrustCenterSettingInput{
		Title:                                 new("Updated Title"),
		NotifySubscribersOnSubprocessorChange: new(true),
	}, nil, nil, nil, nil, nil, nil)
	assert.NilError(t, err)

	unchanged, err := suite.Client.DB.TrustCenterSetting.Get(dbCtx, settingID)
	assert.NilError(t, err)
	assert.Assert(t, unchanged.SubprocessorsNotifiedAt != nil)
	assert.Check(t, stamped.Equal(*unchanged.SubprocessorsNotifiedAt))

	// a watermark set explicitly alongside the flag wins over the hook's stamp
	explicit := time.Now().Add(-3 * time.Hour).Truncate(time.Second)
	assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(settingID).
		SetNotifySubscribersOnSubprocessorChange(false).
		Exec(dbCtx))
	assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(settingID).
		SetNotifySubscribersOnSubprocessorChange(true).
		SetSubprocessorsNotifiedAt(explicit).
		Exec(dbCtx))

	overridden, err := suite.Client.DB.TrustCenterSetting.Get(dbCtx, settingID)
	assert.NilError(t, err)
	assert.Assert(t, overridden.SubprocessorsNotifiedAt != nil)
	assert.Check(t, explicit.Equal(*overridden.SubprocessorsNotifiedAt))

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

// TestDeleteTrustCenterSetting tests the deleteTrustCenterSetting mutation
func TestDeleteTrustCenterSetting(t *testing.T) {
	t.Parallel()
	// Test 1: happy path - delete trust center setting
	t.Run("Delete happy path - delete trust center setting", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		resp, err := suite.Client.API.DeleteTrustCenterSetting(tcOrg.Owner.UserCtx, settingID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(settingID, resp.DeleteTrustCenterSetting.DeletedID))

		// Verify the setting is deleted
		_, err = suite.Client.API.GetTrustCenterSettingByID(tcOrg.Owner.UserCtx, settingID)
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})

	// Test 2: not authorized - view only user
	t.Run("Delete not authorized - view only user", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		_, err := suite.Client.API.DeleteTrustCenterSetting(tcOrg.Member.UserCtx, settingID)
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})

	// Test 3: not authorized - different org user
	t.Run("Delete not authorized - different org user", func(t *testing.T) {
		t.Parallel()

		tcOrg := th.CreateFreshOrgWithTrustCenter(t)
		settingID := tcOrg.TrustCenter.Edges.Setting.ID

		_, err := suite.Client.API.DeleteTrustCenterSetting(th.SharedTestUser2.UserCtx, settingID)
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)

		// Clean up
		th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	})

	// Test 4: trust center setting not found
	t.Run("Delete trust center setting not found", func(t *testing.T) {
		t.Parallel()

		localTestUser := suite.UserBuilder(context.Background(), t) // create new org with no trust center
		_, err := suite.Client.API.DeleteTrustCenterSetting(localTestUser.UserCtx, "non-existent-id")
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)

		th.CleanupOrganizationDataWithContext(localTestUser.UserCtx, t)
	})
}
