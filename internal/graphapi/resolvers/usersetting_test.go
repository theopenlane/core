package resolvers_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryUserSetting(t *testing.T) {
	// setup user context
	reqCtx := testUser1.UserCtx

	user2 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	user2Setting, err := user2.Setting(reqCtx)
	assert.NilError(t, err)

	// setup valid user context
	user1SettingResp, err := suite.client.api.GetUserSettings(reqCtx, testclient.UserSettingWhereInput{})
	assert.NilError(t, err)
	assert.Check(t, is.Len(user1SettingResp.UserSettings.Edges, 1))

	user1Setting := user1SettingResp.UserSettings.Edges[0].Node

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		expected *testclient.GetUserSettings_UserSettings_Edges_Node
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

				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.expected.Status, resp.UserSetting.Status)
		})
	}

	(&Cleanup[*generated.UserDeleteOne]{client: suite.client.db.User, ID: user2.ID}).MustDelete(reqCtx, t)
}

func TestQueryUserSettings(t *testing.T) {
	// setup user context
	reqCtx := testUser1.UserCtx

	user1 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	user1Setting, err := user1.Setting(reqCtx)
	assert.NilError(t, err)

	// create another user to make sure we don't get their settings back
	_ = (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)

	t.Run("Get User Settings", func(t *testing.T) {
		resp, err := suite.client.api.GetAllUserSettings(reqCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.UserSettings.Edges != nil)

		// make sure only the current user settings are returned
		assert.Check(t, is.Equal(len(resp.UserSettings.Edges), 1))

		// setup valid user context
		reqCtx := auth.NewTestContextWithValidUser(user1.ID)

		resp, err = suite.client.api.GetAllUserSettings(reqCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.UserSettings.Edges != nil)
		assert.Equal(t, user1Setting.ID, resp.UserSettings.Edges[0].Node.ID)
	})
}

func TestMutationUpdateUserSetting(t *testing.T) {
	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	om := (&OrgMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	// create another user to make sure we don't get their settings back
	(&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name          string
		userSettingID string
		updateInput   testclient.UpdateUserSettingInput
		client        *testclient.TestClient
		ctx           context.Context
		expectedRes   testclient.UpdateUserSetting_UpdateUserSetting_UserSetting
		errorMsg      string
	}{
		{
			name:          "update default org and tags",
			userSettingID: testUser1.UserInfo.Edges.Setting.ID,
			updateInput: testclient.UpdateUserSettingInput{
				DefaultOrgID: &org.ID,
				Tags:         []string{"mitb", "funk"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			expectedRes: testclient.UpdateUserSetting_UpdateUserSetting_UserSetting{
				Status: enums.UserStatusActive,
				Tags:   []string{"mitb", "funk"},
				DefaultOrg: &testclient.UpdateUserSetting_UpdateUserSetting_UserSetting_DefaultOrg{
					ID: org.ID,
				},
			},
		},
		{
			name:          "update default org and tags for view only user",
			userSettingID: viewOnlyUser.UserInfo.Edges.Setting.ID,
			updateInput: testclient.UpdateUserSettingInput{
				DefaultOrgID: &om.OrganizationID,
				Tags:         []string{"mitb", "funk"},
			},
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
			expectedRes: testclient.UpdateUserSetting_UpdateUserSetting_UserSetting{
				Status: enums.UserStatusActive,
				Tags:   []string{"mitb", "funk"},
				DefaultOrg: &testclient.UpdateUserSetting_UpdateUserSetting_UserSetting_DefaultOrg{
					ID: om.OrganizationID,
				},
			},
		},
		{
			name:          "update default org to org without access",
			userSettingID: testUser1.UserInfo.Edges.Setting.ID,
			updateInput: testclient.UpdateUserSettingInput{
				DefaultOrgID: &org2.ID,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "Organization with the specified ID was not found",
		},
		{
			name:          "update status to invalid",
			userSettingID: testUser1.UserInfo.Edges.Setting.ID,
			updateInput: testclient.UpdateUserSettingInput{
				Status: &enums.UserStatusInvalid,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "INVALID is not a valid UserSettingUserStatus",
		},
		{
			name:          "update status to suspended using personal access token",
			userSettingID: testUser1.UserInfo.Edges.Setting.ID,
			updateInput: testclient.UpdateUserSettingInput{
				Status: &enums.UserStatusSuspended,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
			expectedRes: testclient.UpdateUserSetting_UpdateUserSetting_UserSetting{
				Status: enums.UserStatusSuspended,
				Tags:   []string{"mitb", "funk"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// update user
			resp, err := tc.client.UpdateUserSetting(tc.ctx, tc.userSettingID, tc.updateInput)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Make sure provided values match
			assert.Check(t, is.Equal(tc.expectedRes.Status, resp.UpdateUserSetting.UserSetting.Status))
			assert.DeepEqual(t, tc.expectedRes.Tags, resp.UpdateUserSetting.UserSetting.Tags)

			if tc.updateInput.DefaultOrgID != nil {
				assert.Check(t, is.Equal(tc.expectedRes.DefaultOrg.ID, resp.UpdateUserSetting.UserSetting.DefaultOrg.ID))
			}
		})
	}

	// cleanup created organizations
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, ID: org.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.OrganizationDeleteOne]{client: suite.client.db.Organization, ID: org2.ID}).MustDelete(testUser2.UserCtx, t)
}
