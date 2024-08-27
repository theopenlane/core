package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/rout"
)

func (suite *GraphTestSuite) TestQueryTFASetting() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	(&TFASettingBuilder{client: suite.client}).MustNew(reqCtx, t, testUser.ID)

	user2 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	user2Ctx, err := userContextWithID(user2.ID)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		userID   string
		client   *openlaneclient.OpenLaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:   "happy path user",
			client: suite.client.api,
			ctx:    reqCtx,
		},
		{
			name:   "happy path, using personal access token",
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "valid user, but not auth",
			client:   suite.client.api,
			ctx:      user2Ctx,
			errorMsg: "tfa_setting not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

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

	(&UserCleanup{client: suite.client, ID: user2.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestMutationCreateTFASetting() {
	t := suite.T()

	// setup user context
	ctx, err := userContext()
	require.NoError(t, err)

	user := (&UserBuilder{client: suite.client}).MustNew(ctx, t)
	require.NoError(t, err)

	// setup valid user context
	reqCtx, err := userContextWithID(user.ID)
	require.NoError(t, err)

	testCases := []struct {
		name   string
		userID string
		input  openlaneclient.CreateTFASettingInput
		client *openlaneclient.OpenLaneClient
		ctx    context.Context
		errMsg string
	}{
		{
			name:   "happy path",
			userID: user.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    reqCtx,
		},
		{
			name:   "happy path, using personal access token",
			userID: testUser.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "unable to create using api token",
			userID: testUser.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
			errMsg: rout.ErrBadRequest.Error(),
		},
		{
			name:   "already exists",
			userID: user.ID,
			input: openlaneclient.CreateTFASettingInput{
				TotpAllowed: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    reqCtx,
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
			assert.Empty(t, resp.CreateTFASetting.TfaSetting.RecoveryCodes)
			assert.Equal(t, tc.userID, resp.CreateTFASetting.TfaSetting.Owner.ID)

			// make sure user setting was not updated
			userSetting, err := user.Setting(ctx)
			require.NoError(t, err)

			assert.False(t, userSetting.IsTfaEnabled)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateTFASetting() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	(&TFASettingBuilder{client: suite.client}).MustNew(reqCtx, t, testUser.ID)

	recoveryCodes := []string{}

	testCases := []struct {
		name   string
		input  openlaneclient.UpdateTFASettingInput
		client *openlaneclient.OpenLaneClient
		ctx    context.Context
		errMsg string
	}{
		{
			name: "update verify",
			input: openlaneclient.UpdateTFASettingInput{
				Verified: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    reqCtx,
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
			ctx:    reqCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

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
			assert.NotEmpty(t, resp.UpdateTFASetting.TfaSetting.RecoveryCodes)

			// backup codes should only be regenerated on explicit request
			if tc.input.RegenBackupCodes != nil {
				if *tc.input.RegenBackupCodes {
					assert.NotEqual(t, recoveryCodes, resp.UpdateTFASetting.TfaSetting.RecoveryCodes)
				} else {
					assert.Equal(t, recoveryCodes, resp.UpdateTFASetting.TfaSetting.RecoveryCodes)
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
			recoveryCodes = resp.UpdateTFASetting.TfaSetting.RecoveryCodes
		})
	}
}
