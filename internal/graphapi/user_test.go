package graphapi_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/datumforge/entx"
	mock_fga "github.com/datumforge/fgax/mockery"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi"
	auth "github.com/theopenlane/core/pkg/auth"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryUser() {
	t := suite.T()

	// setup user context
	ctx, err := userContext()
	require.NoError(t, err)

	user1 := (&UserBuilder{client: suite.client}).MustNew(ctx, t)
	user2 := (&UserBuilder{client: suite.client}).MustNew(ctx, t)

	// setup valid user context
	reqCtx, err := auth.NewTestContextWithOrgID(user1.ID, user1.Edges.Setting.Edges.DefaultOrg.ID)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		queryID  string
		expected *ent.User
		errorMsg string
	}{
		{
			name:     "happy path user",
			queryID:  user1.ID,
			expected: user1,
		},
		{
			name:     "valid user, but no auth",
			queryID:  user2.ID,
			errorMsg: "user not found",
		},
		{
			name:     "invalid-id",
			queryID:  "tacos-for-dinner",
			errorMsg: "user not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errorMsg == "" {
				// mock check calls
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			resp, err := suite.client.api.GetUserByID(reqCtx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.User)
		})
	}

	(&UserCleanup{client: suite.client, ID: user1.ID}).MustDelete(reqCtx, t)
	(&UserCleanup{client: suite.client, ID: user2.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestQueryUsers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	user1 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	user2 := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)

	t.Run("Get Users", func(t *testing.T) {
		defer mock_fga.ClearMocks(suite.client.fga)

		resp, err := suite.client.api.GetAllUsers(reqCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Users.Edges)

		// make sure only the current user is returned
		assert.Equal(t, len(resp.Users.Edges), 1)

		// setup valid user context
		reqCtx, err := userContextWithID(user1.ID)
		require.NoError(t, err)

		resp, err = suite.client.api.GetAllUsers(reqCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Users.Edges)

		// only user that is making the request should be returned
		assert.Equal(t, len(resp.Users.Edges), 1)

		user1Found := false
		user2Found := false

		for _, o := range resp.Users.Edges {
			if o.Node.ID == user1.ID {
				user1Found = true
			} else if o.Node.ID == user2.ID {
				user2Found = true
			}
		}

		// only user 1 should be found
		if !user1Found {
			t.Errorf("user 1 was expected to be found but was not")
		}

		// user 2 should not be found
		if user2Found {
			t.Errorf("user 2 was not expected to be found but was returned")
		}
	})
}
func (suite *GraphTestSuite) TestMutationCreateUser() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	// weakPassword := "notsecure"
	strongPassword := "my&supers3cr3tpassw0rd!"

	testCases := []struct {
		name      string
		userInput openlaneclient.CreateUserInput
		errorMsg  string
	}{
		{
			name: "no auth create user",
			userInput: openlaneclient.CreateUserInput{
				FirstName:   lo.ToPtr(gofakeit.FirstName()),
				LastName:    lo.ToPtr(gofakeit.LastName()),
				DisplayName: gofakeit.LetterN(50),
				Email:       gofakeit.Email(),
				Password:    &strongPassword,
			},
			errorMsg: graphapi.ErrPermissionDenied.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errorMsg == "" {
				// mock writes to create personal org membership
				mock_fga.WriteAny(t, suite.client.fga)
			}

			resp, err := suite.client.api.CreateUser(reqCtx, tc.userInput)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateUser.User)

			// Make sure provided values match
			assert.Equal(t, tc.userInput.FirstName, resp.CreateUser.User.FirstName)
			assert.Equal(t, tc.userInput.LastName, resp.CreateUser.User.LastName)
			assert.Equal(t, tc.userInput.Email, resp.CreateUser.User.Email)

			// display name defaults to email if not provided
			if tc.userInput.DisplayName == "" {
				assert.Equal(t, tc.userInput.Email, resp.CreateUser.User.DisplayName)
			} else {
				assert.Equal(t, tc.userInput.DisplayName, resp.CreateUser.User.DisplayName)
			}

			// ensure a user setting was created
			assert.NotNil(t, resp.CreateUser.User.Setting)

			// ensure personal org is created
			// default org will always be the personal org when the user is first created
			personalOrgID := resp.CreateUser.User.Setting.DefaultOrg.ID

			org, err := suite.client.api.GetOrganizationByID(reqCtx, personalOrgID, nil)
			require.NoError(t, err)
			assert.Equal(t, personalOrgID, org.Organization.ID)
			assert.True(t, *org.Organization.PersonalOrg)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateUser() {
	t := suite.T()

	// setup user context
	ctx, err := userContext()
	require.NoError(t, err)

	firstNameUpdate := gofakeit.FirstName()
	lastNameUpdate := gofakeit.LastName()
	emailUpdate := gofakeit.Email()
	displayNameUpdate := gofakeit.LetterN(40)
	nameUpdateLong := gofakeit.LetterN(200)

	user := (&UserBuilder{client: suite.client}).MustNew(ctx, t)

	// setup valid user context
	reqCtx, err := auth.NewTestContextWithOrgID(user.ID, user.Edges.Setting.Edges.DefaultOrg.ID)
	require.NoError(t, err)

	weakPassword := "notsecure"

	testCases := []struct {
		name        string
		updateInput openlaneclient.UpdateUserInput
		expectedRes openlaneclient.UpdateUser_UpdateUser_User
		errorMsg    string
	}{
		{
			name: "update first name and password, happy path",
			updateInput: openlaneclient.UpdateUserInput{
				FirstName: &firstNameUpdate,
			},
			expectedRes: openlaneclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &user.LastName,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			},
		},
		{
			name: "update last name, happy path",
			updateInput: openlaneclient.UpdateUserInput{
				LastName: &lastNameUpdate,
			},
			expectedRes: openlaneclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate, // this would have been updated on the prior test
				LastName:    &lastNameUpdate,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			},
		},
		{
			name: "update email, happy path",
			updateInput: openlaneclient.UpdateUserInput{
				Email: &emailUpdate,
			},
			expectedRes: openlaneclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &lastNameUpdate, // this would have been updated on the prior test
				DisplayName: user.DisplayName,
				Email:       emailUpdate,
			},
		},
		{
			name: "update display name, happy path",
			updateInput: openlaneclient.UpdateUserInput{
				DisplayName: &displayNameUpdate,
			},
			expectedRes: openlaneclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &lastNameUpdate,
				DisplayName: displayNameUpdate,
				Email:       emailUpdate, // this would have been updated on the prior test
			},
		},
		{
			name: "update name, too long",
			updateInput: openlaneclient.UpdateUserInput{
				FirstName: &nameUpdateLong,
			},
			errorMsg: "value is greater than the required length",
		},
		{
			name: "update with weak password",
			updateInput: openlaneclient.UpdateUserInput{
				Password: &weakPassword,
			},
			errorMsg: auth.ErrPasswordTooWeak.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errorMsg == "" {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			// update user
			resp, err := suite.client.api.UpdateUser(reqCtx, user.ID, tc.updateInput)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateUser.User)

			// Make sure provided values match
			updatedUser := resp.GetUpdateUser().User
			assert.Equal(t, tc.expectedRes.FirstName, updatedUser.FirstName)
			assert.Equal(t, tc.expectedRes.LastName, updatedUser.LastName)
			assert.Equal(t, tc.expectedRes.DisplayName, updatedUser.DisplayName)
			assert.Equal(t, tc.expectedRes.Email, updatedUser.Email)
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteUser() {
	t := suite.T()

	// setup user context
	ctx, err := userContext()
	require.NoError(t, err)

	// bypass auth on object creation
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	user := (&UserBuilder{client: suite.client}).MustNew(ctx, t)

	userSetting := user.Edges.Setting

	// personal org will be the default org when the user is created
	personalOrgID := user.Edges.Setting.Edges.DefaultOrg.ID

	// setup valid user context
	reqCtx, err := auth.NewTestContextWithOrgID(user.ID, personalOrgID)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		userID   string
		errorMsg string
	}{
		{
			name:   "delete user, happy path",
			userID: user.ID,
		},
		{
			name:     "delete user, not found",
			userID:   "tacos-tuesday",
			errorMsg: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// mock check calls
			if tc.errorMsg == "" {
				mock_fga.CheckAny(t, suite.client.fga, true)

				mock_fga.WriteAny(t, suite.client.fga)
			}

			// delete user
			resp, err := suite.client.api.DeleteUser(reqCtx, tc.userID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.DeleteUser.DeletedID)

			// make sure the personal org is deleted
			org, err := suite.client.api.GetOrganizationByID(reqCtx, personalOrgID)
			require.Nil(t, org)
			require.Error(t, err)
			assert.ErrorContains(t, err, "not found")

			// make sure the deletedID matches the ID we wanted to delete
			assert.Equal(t, tc.userID, resp.DeleteUser.DeletedID)

			// make sure the user setting is deleted
			out, err := suite.client.api.GetUserSettingByID(reqCtx, userSetting.ID)
			require.Nil(t, out)
			require.Error(t, err)
			assert.ErrorContains(t, err, "not found")
		})
	}
}

func (suite *GraphTestSuite) TestMutationUserCascadeDelete() {
	t := suite.T()

	// setup user context
	ctx, err := userContext()
	require.NoError(t, err)

	user := (&UserBuilder{client: suite.client}).MustNew(ctx, t)

	reqCtx, err := auth.NewTestContextWithOrgID(user.ID, user.Edges.Setting.Edges.DefaultOrg.ID)
	require.NoError(t, err)

	token := (&PersonalAccessTokenBuilder{client: suite.client, OwnerID: user.ID}).MustNew(reqCtx, t)

	// mock checks
	mock_fga.CheckAny(t, suite.client.fga, true)
	// mock writes to clean up personal org
	mock_fga.WriteAny(t, suite.client.fga)

	// delete user
	resp, err := suite.client.api.DeleteUser(reqCtx, user.ID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.DeleteUser.DeletedID)

	// make sure the deletedID matches the ID we wanted to delete
	assert.Equal(t, user.ID, resp.DeleteUser.DeletedID)

	o, err := suite.client.api.GetUserByID(reqCtx, user.ID)

	require.Nil(t, o)
	require.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	g, err := suite.client.api.GetPersonalAccessTokenByID(reqCtx, token.ID)
	require.Error(t, err)

	require.Nil(t, g)
	assert.ErrorContains(t, err, "not found")

	ctx = entx.SkipSoftDelete(reqCtx)

	// skip checks because tuples will be deleted at this point
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	o, err = suite.client.api.GetUserByID(ctx, user.ID)
	require.NoError(t, err)

	require.Equal(t, o.User.ID, user.ID)

	// Bypass auth check to get owner of access token
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	g, err = suite.client.api.GetPersonalAccessTokenByID(ctx, token.ID)
	require.NoError(t, err)

	require.Equal(t, g.PersonalAccessToken.ID, token.ID)
}
