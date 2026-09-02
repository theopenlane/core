package graphapi_test

import (
	"context"
	"reflect"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryTFASetting(t *testing.T) {
	t.Parallel()
	// create a user for this test
	testUser := suite.SeedOrgOwner(t)

	(&th.TFASettingBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)

	testCases := []struct {
		name     string
		userID   string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:   "happy path user",
			client: suite.Client.API,
			ctx:    testUser.Owner.UserCtx,
		},
		{
			name:   "happy path, using personal access token",
			client: testUser.PatClient,
			ctx:    context.Background(),
		},
		{
			name:     "valid user, but not auth",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTFASetting(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}

	// cleanup
	th.CleanupOrganizationDataWithContext(testUser.Owner.UserCtx, t)
}

func TestMutationCreateTFASetting(t *testing.T) {
	// create a user for this test
	t.Parallel()
	localTestUser := suite.SeedOrgOwner(t)
	localTestUser2 := suite.UserBuilder(context.Background(), t)
	testUserAnother := suite.UserBuilder(context.Background(), t)

	testCases := []struct {
		name   string
		userID string
		input  testclient.CreateTFASettingInput
		client *testclient.TestClient
		ctx    context.Context
		errMsg string
	}{
		{
			name:   "happy path",
			userID: localTestUser2.ID,
			input: testclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    localTestUser2.UserCtx,
		},
		{
			name:   "happy path, using personal access token",
			userID: localTestUser.Owner.ID,
			input: testclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: localTestUser.PatClient,
			ctx:    context.Background(),
		},
		{
			name:   "unable to create using api token",
			userID: localTestUser.Owner.ID,
			input: testclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: localTestUser.APIClient,
			ctx:    context.Background(),
			errMsg: rout.ErrBadRequest.Error(),
		},
		{
			name:   "already exists",
			userID: localTestUser.Owner.ID,
			input: testclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
			errMsg: "tfasetting already exists",
		},
		{
			name:   "create with not enabling totp should not return qr code",
			userID: testUserAnother.ID,
			input: testclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(false),
			},
			client: suite.Client.API,
			ctx:    testUserAnother.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// create tfa setting for user
			resp, err := tc.client.CreateTFASetting(tc.ctx, tc.input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Make sure provided values match
			assert.Check(t, is.DeepEqual(tc.input.TotpAllowed, resp.CreateTFASetting.TfaSetting.TotpAllowed))

			if *tc.input.TotpAllowed {
				assert.Check(t, resp.CreateTFASetting.QRCode != nil)
				assert.Check(t, resp.CreateTFASetting.TfaSecret != nil)
			} else {
				assert.Check(t, is.Equal(*resp.CreateTFASetting.QRCode, ""))
				assert.Check(t, is.Equal(*resp.CreateTFASetting.TfaSecret, ""))
			}

			assert.Assert(t, resp.CreateTFASetting.TfaSetting.Owner != nil)
			assert.Check(t, is.Equal(tc.userID, resp.CreateTFASetting.TfaSetting.Owner.ID))

			// make sure user setting was not updated
			userSetting, err := th.SharedTestUser1.UserInfo.Setting(th.SharedTestUser1.UserCtx)
			assert.NilError(t, err)

			assert.Check(t, !userSetting.IsTfaEnabled)
		})
	}

	// cleanup
	_, err := suite.Client.API.GetTFASetting(localTestUser2.UserCtx)
	assert.NilError(t, err)

	th.CleanupOrganizationDataWithContext(localTestUser.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(localTestUser2.UserCtx, t)
	th.CleanupOrganizationDataWithContext(testUserAnother.UserCtx, t)
}

func TestMutationUpdateTFASetting(t *testing.T) {
	t.Parallel()
	localTestUser := suite.SeedOrgOwner(t)

	// create tfa settings for user
	(&th.TFASettingBuilder{Client: suite.Client}).MustNew(localTestUser.Owner.UserCtx, t)

	recoveryCodes := []string{}

	testCases := []struct {
		name   string
		input  testclient.UpdateTFASettingInput
		client *testclient.TestClient
		ctx    context.Context
		errMsg string
	}{
		{
			name: "update verify",
			input: testclient.UpdateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
				Verified:    lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
		},
		{
			name: "regen codes using personal access token",
			input: testclient.UpdateTFASettingInput{
				RegenBackupCodes: lo.ToPtr(true),
			},
			client: localTestUser.PatClient,
			ctx:    context.Background(),
		},
		{
			name: "regen codes using api token not allowed",
			input: testclient.UpdateTFASettingInput{
				RegenBackupCodes: lo.ToPtr(true),
			},
			client: localTestUser.APIClient,
			ctx:    context.Background(),
			errMsg: rout.ErrBadRequest.Error(),
		},
		{
			name: "regen codes - false",
			input: testclient.UpdateTFASettingInput{
				RegenBackupCodes: lo.ToPtr(false),
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
		},
		{
			name: "update totp to false should clear settings",
			input: testclient.UpdateTFASettingInput{
				TotpAllowed: lo.ToPtr(false),
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
		},
		{
			name: "update TotpAllowed to true should enable TFA",
			input: testclient.UpdateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// update tfa settings
			resp, err := tc.client.UpdateTFASetting(tc.ctx, tc.input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// backup codes should only be regenerated on explicit request
			// and should only be returned on initial verification or regen request
			if (tc.input.RegenBackupCodes != nil && *tc.input.RegenBackupCodes) ||
				(tc.input.Verified != nil && *tc.input.Verified) {
				// recovery codes should be returned
				assert.Check(t, len(resp.UpdateTFASetting.RecoveryCodes) != 0)

				if tc.input.RegenBackupCodes != nil {
					if *tc.input.RegenBackupCodes {
						assert.Assert(t, !reflect.DeepEqual(recoveryCodes, resp.UpdateTFASetting.RecoveryCodes))
					} else {
						assert.Check(t, is.DeepEqual(recoveryCodes, resp.UpdateTFASetting.RecoveryCodes))
					}
				}
			} else {
				assert.Check(t, is.Len(resp.UpdateTFASetting.RecoveryCodes, 0))
			}

			if tc.input.TotpAllowed == nil || *tc.input.TotpAllowed {
				assert.Check(t, resp.UpdateTFASetting.QRCode != nil)
				assert.Check(t, resp.UpdateTFASetting.TfaSecret != nil)
			} else if !*tc.input.TotpAllowed { // settings were cleared
				assert.Check(t, is.Equal(*resp.UpdateTFASetting.QRCode, ""))
				assert.Check(t, is.Equal(*resp.UpdateTFASetting.TfaSecret, ""))
				assert.Check(t, is.Len(resp.UpdateTFASetting.RecoveryCodes, 0))
				assert.Check(t, !resp.UpdateTFASetting.TfaSetting.Verified)
				assert.Check(t, !*resp.UpdateTFASetting.TfaSetting.TotpAllowed)
			}

			// make sure user setting is updated correctly
			userSettings, err := suite.Client.API.GetUserSettingByID(localTestUser.Owner.UserCtx, localTestUser.Owner.UserInfo.Edges.Setting.ID)
			assert.NilError(t, err)

			if resp.UpdateTFASetting.TfaSetting.Verified {
				assert.Check(t, *userSettings.UserSetting.IsTfaEnabled)
			}

			// ensure TFA is disabled if totp is not allowed
			if !*resp.UpdateTFASetting.TfaSetting.TotpAllowed {
				assert.Check(t, !*userSettings.UserSetting.IsTfaEnabled)
			}

			// set at the end so we can compare later
			recoveryCodes = resp.UpdateTFASetting.RecoveryCodes
		})
	}

	// cleanup
	th.CleanupOrganizationDataWithContext(localTestUser.Owner.UserCtx, t)
}
