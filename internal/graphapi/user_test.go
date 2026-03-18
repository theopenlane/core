package graphapi_test

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/rout"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	auth "github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryUser(t *testing.T) {
	testCases := []struct {
		name     string
		queryID  string
		expected ent.User
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
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, len(resp.User.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.User.DisplayID, "USR-"))
		})
	}
}

func TestQueryUsers(t *testing.T) {

	t.Run("Get Users", func(t *testing.T) {
		resp, err := suite.client.api.GetAllUsers(testUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Users.Edges != nil)

		// make sure only the current user is returned
		assert.Check(t, is.Len(resp.Users.Edges, 1))

		// setup valid user context
		reqCtx := testUser1.UserCtx

		resp, err = suite.client.api.GetAllUsers(reqCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Users.Edges != nil)

		// only user that is making the request should be returned
		assert.Check(t, is.Len(resp.Users.Edges, 1))

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
		assert.Check(t, user1Found)
		// user 2 should not be found
		assert.Check(t, !user2Found)
	})
}

func TestMutationCreateUser(t *testing.T) {
	strongPassword := "my&supers3cr3tpassw0rd!"

	testCases := []struct {
		name       string
		userInput  testclient.CreateUserInput
		avatarFile *graphql.Upload
		errorMsg   string
	}{
		{
			name: "no auth create user",
			userInput: testclient.CreateUserInput{
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
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Make sure provided values match
			assert.Check(t, is.DeepEqual(tc.userInput.FirstName, resp.CreateUser.User.FirstName))
			assert.Check(t, is.DeepEqual(tc.userInput.LastName, resp.CreateUser.User.LastName))
			assert.Check(t, is.Equal(tc.userInput.Email, resp.CreateUser.User.Email))

			// display name defaults to email if not provided
			if tc.userInput.DisplayName == "" {
				assert.Check(t, is.Equal(tc.userInput.Email, resp.CreateUser.User.DisplayName))
			} else {
				assert.Check(t, is.Equal(tc.userInput.DisplayName, resp.CreateUser.User.DisplayName))
			}

			// ensure personal org is created
			// default org will always be the personal org when the user is first created
			personalOrgID := resp.CreateUser.User.Setting.DefaultOrg.ID

			org, err := suite.client.api.GetOrganizationByID(testUser1.UserCtx, personalOrgID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(personalOrgID, org.Organization.ID))
			assert.Check(t, *org.Organization.PersonalOrg)
		})
	}
}

func TestMutationUpdateUser(t *testing.T) {
	firstNameUpdate := gofakeit.FirstName()
	lastNameUpdate := gofakeit.LastName()
	emailUpdate := gofakeit.Email()
	displayNameUpdate := gofakeit.LetterN(40)
	nameUpdateLong := gofakeit.LetterN(200)

	user := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	orgID := user.Edges.Setting.Edges.DefaultOrg.ID

	// setup valid user context
	reqCtx := auth.NewTestContextWithOrgID(user.ID, orgID)

	weakPassword := "notsecure"

	avatarFile := uploadFile(t, logoFilePath)

	invalidAvatarFile := uploadFile(t, txtFilePath)

	testCases := []struct {
		name        string
		updateInput testclient.UpdateUserInput
		avatarFile  *graphql.Upload
		expectedRes testclient.UpdateUser_UpdateUser_User
		errorMsg    string
	}{
		{
			name: "update first name and password, happy path",
			updateInput: testclient.UpdateUserInput{
				FirstName: &firstNameUpdate,
			},
			expectedRes: testclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &user.LastName,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			},
		},
		{
			name:       "update avatar",
			avatarFile: avatarFile,
			expectedRes: testclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &user.LastName,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			},
		},
		{
			name:       "update avatar with invalid file",
			avatarFile: invalidAvatarFile,
			errorMsg:   "unsupported mime type uploaded: text/plain",
		},
		{
			name: "update last name, happy path",
			updateInput: testclient.UpdateUserInput{
				LastName: &lastNameUpdate,
			},
			expectedRes: testclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate, // this would have been updated on the prior test
				LastName:    &lastNameUpdate,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			},
		},
		{
			name: "update email, happy path",
			updateInput: testclient.UpdateUserInput{
				Email: &emailUpdate,
			},
			expectedRes: testclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &lastNameUpdate, // this would have been updated on the prior test
				DisplayName: user.DisplayName,
				Email:       emailUpdate,
			},
		},
		{
			name: "update display name, happy path",
			updateInput: testclient.UpdateUserInput{
				DisplayName: &displayNameUpdate,
			},
			expectedRes: testclient.UpdateUser_UpdateUser_User{
				ID:          user.ID,
				FirstName:   &firstNameUpdate,
				LastName:    &lastNameUpdate,
				DisplayName: displayNameUpdate,
				Email:       emailUpdate, // this would have been updated on the prior test
			},
		},
		{
			name: "update name, too long",
			updateInput: testclient.UpdateUserInput{
				FirstName: &nameUpdateLong,
			},
			errorMsg: "value is greater than the required length",
		},
		{
			name: "update with weak password",
			updateInput: testclient.UpdateUserInput{
				Password: &weakPassword,
			},
			errorMsg: auth.ErrPasswordTooWeak.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.avatarFile != nil {
				if tc.errorMsg == "" {
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.avatarFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}

			// update user
			resp, err := suite.client.api.UpdateUser(reqCtx, user.ID, tc.updateInput, tc.avatarFile)
			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Make sure provided values match
			updatedUser := resp.GetUpdateUser().User
			assert.Check(t, is.DeepEqual(tc.expectedRes.FirstName, updatedUser.FirstName))
			assert.Check(t, is.DeepEqual(tc.expectedRes.LastName, updatedUser.LastName))
			assert.Check(t, is.Equal(tc.expectedRes.DisplayName, updatedUser.DisplayName))
			assert.Check(t, is.Equal(tc.expectedRes.Email, updatedUser.Email))

			if tc.avatarFile != nil {
				assert.Check(t, updatedUser.AvatarLocalFileID != nil)
				assert.Check(t, updatedUser.AvatarFile.PresignedURL != nil)
			}
		})
	}
}

func TestMutationDeleteUser(t *testing.T) {
	// bypass auth on object creation
	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	user := (&UserBuilder{client: suite.client}).MustNew(ctx, t)

	userSetting := user.Edges.Setting

	// personal org will be the default org when the user is created
	personalOrgID := user.Edges.Setting.Edges.DefaultOrg.ID

	// setup valid user context
	reqCtx := auth.NewTestContextWithOrgID(user.ID, personalOrgID)

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
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.DeleteUser(reqCtx, tc.userID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.DeleteUser.DeletedID != "")

			// make sure the personal org is deleted
			// add allow context to bypass auth since the tuple will be deleted
			reqCtx = privacy.DecisionContext(reqCtx, privacy.Allow)

			_, err = suite.client.api.GetOrganizationByID(reqCtx, personalOrgID)

			assert.ErrorContains(t, err, notFoundErrorMsg)

			// make sure the deletedID matches the ID we wanted to delete
			assert.Check(t, is.Equal(tc.userID, resp.DeleteUser.DeletedID))

			// make sure the user setting is deleted
			_, err = suite.client.api.GetUserSettingByID(reqCtx, userSetting.ID)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}
}

func TestMutationUserCascadeDelete(t *testing.T) {
	user := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	reqCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.Setting.Edges.DefaultOrg.ID)

	token := (&PersonalAccessTokenBuilder{client: suite.client, OrganizationIDs: []string{user.Edges.Setting.Edges.DefaultOrg.ID}}).MustNew(reqCtx, t)

	resp, err := suite.client.api.DeleteUser(reqCtx, user.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.DeleteUser.DeletedID != "")

	// make sure the deletedID matches the ID we wanted to delete
	assert.Check(t, is.Equal(user.ID, resp.DeleteUser.DeletedID))

	_, err = suite.client.api.GetUserByID(reqCtx, user.ID)

	assert.ErrorContains(t, err, notFoundErrorMsg)

	_, err = suite.client.api.GetPersonalAccessTokenByID(reqCtx, token.ID)

	assert.ErrorContains(t, err, notFoundErrorMsg)
}
