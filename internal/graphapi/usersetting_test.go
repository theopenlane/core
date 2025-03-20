package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
)

func (suite *GraphTestSuite) TestQueryUserSetting() {
	t := suite.T()

	// setup user context
	reqCtx := testUser1.UserCtx

	user2 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	user2Setting, err := user2.Setting(reqCtx)
	require.NoError(t, err)

	// setup valid user context
	user1SettingResp, err := suite.client.api.GetUserSettings(reqCtx, openlaneclient.UserSettingWhereInput{})
	require.NoError(t, err)
	require.Len(t, user1SettingResp.UserSettings.Edges, 1)

	user1Setting := user1SettingResp.UserSettings.Edges[0].Node

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		expected *openlaneclient.GetUserSettings_UserSettings_Edges_Node
		errorMsg string
	}{
		{
			name:     "happy path user",
			queryID:  user1Setting.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			expected: user1Setting,
		},
		{
			name:     "happy path user, using personal access token",
			queryID:  user1Setting.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			expected: user1Setting,
		},
		{
			name:     "valid user, but not auth",
			queryID:  user2Setting.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetUserSettingByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UserSetting)
			require.Equal(t, tc.expected.Status, resp.UserSetting.Status)
		})
	}

	(&Cleanup[*generated.UserDeleteOne]{client: suite.client.db.User, ID: user2.ID}).MustDelete(reqCtx, suite)
}

func (suite *GraphTestSuite) TestQueryUserSettings() {
	t := suite.T()

	// setup user context
	reqCtx := testUser1.UserCtx

	user1 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	user1Setting, err := user1.Setting(reqCtx)
	require.NoError(t, err)

	// create another user to make sure we don't get their settings back
	_ = (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)

	t.Run("Get User Settings", func(t *testing.T) {
		resp, err := suite.client.api.GetAllUserSettings(reqCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.UserSettings.Edges)

		// make sure only the current user settings are returned
		assert.Equal(t, len(resp.UserSettings.Edges), 1)

		// setup valid user context
		reqCtx := auth.NewTestContextWithValidUser(user1.ID)

		resp, err = suite.client.api.GetAllUserSettings(reqCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.UserSettings.Edges)
		require.Equal(t, user1Setting.ID, resp.UserSettings.Edges[0].Node.ID)
	})
}

func (suite *GraphTestSuite) TestMutationUpdateUserSetting() {
	t := suite.T()

	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create another user to make sure we don't get their settings back
	(&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		updateInput openlaneclient.UpdateUserSettingInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedRes openlaneclient.UpdateUserSetting_UpdateUserSetting_UserSetting
		errorMsg    string
	}{
		{
			name: "update default org and tags",
			updateInput: openlaneclient.UpdateUserSettingInput{
				DefaultOrgID: &org.ID,
				Tags:         []string{"mitb", "funk"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: openlaneclient.UpdateUserSetting_UpdateUserSetting_UserSetting{
				Status: enums.UserStatusActive,
				Tags:   []string{"mitb", "funk"},
				DefaultOrg: &openlaneclient.UpdateUserSetting_UpdateUserSetting_UserSetting_DefaultOrg{
					ID: org.ID,
				},
			},
		},
		{
			name: "update default org to org without access",
			updateInput: openlaneclient.UpdateUserSettingInput{
				DefaultOrgID: &org2.ID,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "Organization with the specified ID was not found",
		},
		{
			name: "update status to invalid",
			updateInput: openlaneclient.UpdateUserSettingInput{
				Status: &enums.UserStatusInvalid,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "INVALID is not a valid UserSettingUserStatus",
		},
		{
			name: "update status to suspended using personal access token",
			updateInput: openlaneclient.UpdateUserSettingInput{
				Status: &enums.UserStatusSuspended,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
			expectedRes: openlaneclient.UpdateUserSetting_UpdateUserSetting_UserSetting{
				Status: enums.UserStatusSuspended,
				Tags:   []string{"mitb", "funk"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// update user
			resp, err := tc.client.UpdateUserSetting(tc.ctx, testUser1.UserInfo.Edges.Setting.ID, tc.updateInput)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateUserSetting.UserSetting)

			// Make sure provided values match
			assert.Equal(t, tc.expectedRes.Status, resp.UpdateUserSetting.UserSetting.Status)
			assert.ElementsMatch(t, tc.expectedRes.Tags, resp.UpdateUserSetting.UserSetting.Tags)

			if tc.updateInput.DefaultOrgID != nil {
				assert.Equal(t, tc.expectedRes.DefaultOrg.ID, resp.UpdateUserSetting.UserSetting.DefaultOrg.ID)
			}
		})
	}
}
