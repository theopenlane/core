package graphapi_test

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/rout"

	auth "github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryUser() {
	t := suite.T()

	testCases := []struct {
		name     string
		queryID  string
		expected *ent.User
		errorMsg string
	}{
		{
			name:     "happy path user",
			queryID:  testUser1.ID,
			expected: testUser1.UserInfo,
		},
		{
			name:     "valid user, but no auth",
			queryID:  testUser2.ID,
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
			resp, err := suite.client.api.GetUserByID(testUser1.UserCtx, tc.queryID)

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
}

func (suite *GraphTestSuite) TestQueryUsers() {
	t := suite.T()

	t.Run("Get Users", func(t *testing.T) {
		resp, err := suite.client.api.GetAllUsers(testUser1.UserCtx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Users.Edges)

		// make sure only the current user is returned
		assert.Equal(t, len(resp.Users.Edges), 1)

		// setup valid user context
		reqCtx, err := userContextWithID(testUser1.ID)
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
			if o.Node.ID == testUser1.ID {
				user1Found = true
			} else if o.Node.ID == testUser2.ID {
				user2Found = true
			}
		}

		// only user 1 should be found
		assert.True(t, user1Found)
		// user 2 should not be found
		assert.False(t, user2Found)
	})
}

func (suite *GraphTestSuite) TestMutationCreateUser() {
	t := suite.T()

	strongPassword := "my&supers3cr3tpassw0rd!"

	testCases := []struct {
		name       string
		userInput  openlaneclient.CreateUserInput
		avatarFile *graphql.Upload
		errorMsg   string
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
			errorMsg: rout.ErrPermissionDenied.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.CreateUser(testUser1.UserCtx, tc.userInput, tc.avatarFile)

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

			org, err := suite.client.api.GetOrganizationByID(testUser1.UserCtx, personalOrgID)
			require.NoError(t, err)
			assert.Equal(t, personalOrgID, org.Organization.ID)
			assert.True(t, *org.Organization.PersonalOrg)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateUser() {
	t := suite.T()

	firstNameUpdate := gofakeit.FirstName()
	lastNameUpdate := gofakeit.LastName()
	emailUpdate := gofakeit.Email()
	displayNameUpdate := gofakeit.LetterN(40)
	nameUpdateLong := gofakeit.LetterN(200)

	user := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	orgID := user.Edges.Setting.Edges.DefaultOrg.ID

	// setup valid user context
	reqCtx, err := auth.NewTestContextWithOrgID(user.ID, orgID)
	require.NoError(t, err)

	weakPassword := "notsecure"

	avatarFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	require.NoError(t, err)

	invalidAvatarFile, err := objects.NewUploadFile("testdata/uploads/hello.txt")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		updateInput openlaneclient.UpdateUserInput
		avatarFile  *graphql.Upload
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
			name: "update avatar",
			avatarFile: &graphql.Upload{
				File:        avatarFile.File,
				Filename:    avatarFile.Filename,
				Size:        avatarFile.Size,
				ContentType: avatarFile.ContentType,
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
			name: "update avatar with invalid file",
			avatarFile: &graphql.Upload{
				File:        invalidAvatarFile.File,
				Filename:    invalidAvatarFile.Filename,
				Size:        invalidAvatarFile.Size,
				ContentType: invalidAvatarFile.ContentType,
			},
			errorMsg: "unsupported mime type uploaded: text/plain",
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
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*tc.avatarFile})
				} else {
					expectUploadCheckOnly(t, suite.client.objectStore.Storage)
				}
			}

			// update user
			resp, err := suite.client.api.UpdateUser(reqCtx, user.ID, tc.updateInput, tc.avatarFile)
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

			if tc.avatarFile != nil {
				assert.NotNil(t, updatedUser.AvatarLocalFileID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteUser() {
	t := suite.T()

	// bypass auth on object creation
	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

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

	user := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	reqCtx, err := auth.NewTestContextWithOrgID(user.ID, user.Edges.Setting.Edges.DefaultOrg.ID)
	require.NoError(t, err)

	token := (&PersonalAccessTokenBuilder{client: suite.client, OwnerID: user.ID}).MustNew(reqCtx, t)

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
}
