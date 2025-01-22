package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryTFASetting() {
	t := suite.T()

	(&TFASettingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t, testUser1.ID)

	testCases := []struct {
		name     string
		userID   string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:   "happy path user",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "happy path, using personal access token",
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "valid user, but not auth",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTFASetting(tc.ctx)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateTFASetting() {
	t := suite.T()

	testCases := []struct {
		name   string
		userID string
		input  openlaneclient.CreateTFASettingInput
		client *openlaneclient.OpenlaneClient
		ctx    context.Context
		errMsg string
	}{
		{
			name:   "happy path",
			userID: testUser2.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name:   "happy path, using personal access token",
			userID: testUser1.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "unable to create using api token",
			userID: testUser1.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
			errMsg: rout.ErrBadRequest.Error(),
		},
		{
			name:   "already exists",
			userID: testUser1.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			errMsg: "tfasetting already exists",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// create tfa setting for user
			resp, err := tc.client.CreateTFASetting(tc.ctx, tc.input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateTFASetting.TfaSetting)

			// Make sure provided values match
			assert.Equal(t, tc.input.TotpAllowed, resp.CreateTFASetting.TfaSetting.TotpAllowed)
			require.NotEmpty(t, resp.CreateTFASetting.TfaSetting.Owner)
			assert.Equal(t, tc.userID, resp.CreateTFASetting.TfaSetting.Owner.ID)

			// make sure user setting was not updated
			userSetting, err := testUser1.UserInfo.Setting(testUser1.UserCtx)
			require.NoError(t, err)

			assert.False(t, userSetting.IsTfaEnabled)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateTFASetting() {
	t := suite.T()

	(&TFASettingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t, testUser1.ID)

	recoveryCodes := []string{}

	testCases := []struct {
		name   string
		input  openlaneclient.UpdateTFASettingInput
		client *openlaneclient.OpenlaneClient
		ctx    context.Context
		errMsg string
	}{
		{
			name: "update verify",
			input: openlaneclient.UpdateTFASettingInput{
				Verified: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "regen codes using personal access token",
			input: openlaneclient.UpdateTFASettingInput{
				RegenBackupCodes: lo.ToPtr(true),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "regen codes using api token not allowed",
			input: openlaneclient.UpdateTFASettingInput{
				RegenBackupCodes: lo.ToPtr(true),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
			errMsg: rout.ErrBadRequest.Error(),
		},
		{
			name: "regen codes - false",
			input: openlaneclient.UpdateTFASettingInput{
				RegenBackupCodes: lo.ToPtr(false),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// update tfa settings
			resp, err := tc.client.UpdateTFASetting(tc.ctx, tc.input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateTFASetting.TfaSetting)

			// Make sure provided values match
			assert.NotEmpty(t, resp.UpdateTFASetting.RecoveryCodes)

			// backup codes should only be regenerated on explicit request
			if tc.input.RegenBackupCodes != nil {
				if *tc.input.RegenBackupCodes {
					assert.NotEqual(t, recoveryCodes, resp.UpdateTFASetting.RecoveryCodes)
				} else {
					assert.Equal(t, recoveryCodes, resp.UpdateTFASetting.RecoveryCodes)
				}
			}

			// make sure user setting was not updated
			userSettings, err := tc.client.GetAllUserSettings(tc.ctx)
			require.NoError(t, err)
			require.Len(t, userSettings.UserSettings.Edges, 1)

			if resp.UpdateTFASetting.TfaSetting.Verified {
				assert.True(t, *userSettings.UserSettings.Edges[0].Node.IsTfaEnabled)
			}

			// set at the end so we can compare later
			recoveryCodes = resp.UpdateTFASetting.RecoveryCodes
		})
	}
}
