package graphapi_test

import (
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/rout"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	auth "github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
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
			queryID:  th.SharedTestUser1.ID,
			expected: th.SharedTestUser1.UserInfo,
		},
		{
			name:     "valid user, but no auth",
			queryID:  th.SharedTestUser2.ID,
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
			resp, err := suite.Client.API.GetUserByID(th.SharedTestUser1.UserCtx, tc.queryID)

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
		resp, err := suite.Client.API.GetAllUsers(th.SharedTestUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Users.Edges != nil)

		// make sure only the current user is returned
		assert.Check(t, is.Len(resp.Users.Edges, 1))

		// setup valid user context
		reqCtx := th.SharedTestUser1.UserCtx

		resp, err = suite.Client.API.GetAllUsers(reqCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Users.Edges != nil)

		// only user that is making the request should be returned
		assert.Check(t, is.Len(resp.Users.Edges, 1))

		user1Found := false
		user2Found := false

		for _, o := range resp.Users.Edges {
			if o.Node.ID == th.SharedTestUser1.ID {
				user1Found = true
			} else if o.Node.ID == th.SharedTestUser2.ID {
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
			resp, err := suite.Client.API.CreateUser(th.SharedTestUser1.UserCtx, tc.userInput, tc.avatarFile, nil)

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

			org, err := suite.Client.API.GetOrganizationByID(th.SharedTestUser1.UserCtx, personalOrgID)
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

	user := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	orgID := user.Edges.Setting.Edges.DefaultOrg.ID

	// setup valid user context
	reqCtx := auth.NewTestContextWithOrgID(user.ID, orgID)

	weakPassword := "notsecure"

	avatarFile := th.UploadFile(t, th.LogoFilePath)

	invalidAvatarFile := th.UploadFile(t, th.TxtFilePath)

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
					th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.avatarFile})
				} else {
					th.ExpectUploadCheckOnly(t, suite.Client.MockProvider)
				}
			}

			// update user
			resp, err := suite.Client.API.UpdateUser(reqCtx, user.ID, tc.updateInput, tc.avatarFile, nil)
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
	ctx := privacy.DecisionContext(th.SharedTestUser1.UserCtx, privacy.Allow)

	user := (&th.UserBuilder{Client: suite.Client}).MustNew(ctx, t)

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
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.DeleteUser(reqCtx, tc.userID)

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

			_, err = suite.Client.API.GetOrganizationByID(reqCtx, personalOrgID)

			assert.ErrorContains(t, err, th.NotFoundErrorMsg)

			// make sure the deletedID matches the ID we wanted to delete
			assert.Check(t, is.Equal(tc.userID, resp.DeleteUser.DeletedID))

			// make sure the user setting is deleted
			_, err = suite.Client.API.GetUserSettingByID(reqCtx, userSetting.ID)
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)
		})
	}
}

func TestQueryUserSupportContext(t *testing.T) {
	orgID := th.SharedTestUser1.OrganizationID

	caller := auth.NewOrgSupportCaller(orgID, auth.SupportSubjectID, th.SupportSubjectName, th.SupportSubjectEmail)
	supportCtx := auth.WithCaller(th.SharedTestUser1.UserCtx, caller)

	t.Run("Self returns synthetic support user", func(t *testing.T) {
		resp, err := suite.Client.API.GetSelf(supportCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		assert.Check(t, is.Equal(auth.SupportSubjectID, resp.Self.ID))
		assert.Check(t, is.Equal(th.SupportSubjectName, resp.Self.DisplayName))
		assert.Check(t, is.Equal(th.SupportSubjectEmail, resp.Self.Email))
		assert.Check(t, resp.Self.Setting.EmailConfirmed)
		assert.Check(t, resp.Self.Setting.DefaultOrg != nil)
		assert.Check(t, is.Equal(orgID, resp.Self.Setting.DefaultOrg.ID))
	})

	t.Run("User returns synthetic support user", func(t *testing.T) {
		resp, err := suite.Client.API.GetUserByID(supportCtx, auth.SupportSubjectID)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		assert.Check(t, is.Equal(auth.SupportSubjectID, resp.User.ID))
		assert.Check(t, is.Equal(th.SupportSubjectName, resp.User.DisplayName))
		assert.Check(t, is.Equal(th.SupportSubjectEmail, resp.User.Email))
	})
}

func TestMutationDeleteUser_OrgOwnerCannotBeDeleted(t *testing.T) {
	t.Parallel()
	localTestOrg := suite.SeedFreshOrgUsers(t)

	_, err := suite.Client.API.DeleteUser(localTestOrg.Owner.UserCtx, localTestOrg.Owner.ID)

	assert.ErrorContains(t, err, hooks.ErrOrgOwnerCannotBeDeleted.Error())

	// cleanup
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestMutationDeleteUser_OrgOwnerCannotDeleteFromAnotherOrg(t *testing.T) {
	t.Parallel()

	newOrgAsOwner := suite.SeedOrgOwner(t)

	// storing here to use in cleanup because addUserToOrganization updates the active org context of the user
	ownerOrgCtx := newOrgAsOwner.Owner.UserCtx

	newOrgAsMember := suite.SeedFreshMinimalOrgUsers(t, false)

	suite.AddUserToOrganization(newOrgAsMember.Owner.UserCtx, t, newOrgAsOwner.Owner, enums.RoleMember, newOrgAsMember.Owner.OrganizationID)

	_, err := suite.Client.API.DeleteUser(newOrgAsOwner.Owner.UserCtx, newOrgAsOwner.Owner.ID)

	assert.ErrorContains(t, err, hooks.ErrOrgOwnerCannotBeDeleted.Error())

	th.CleanupOrganizationDataWithContext(ownerOrgCtx, t)
	th.CleanupOrganizationDataWithContext(newOrgAsMember.Owner.UserCtx, t)
}

func TestMutationUserCascadeDelete(t *testing.T) {
	user := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	reqCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.Setting.Edges.DefaultOrg.ID)

	token := (&th.PersonalAccessTokenBuilder{Client: suite.Client, OrganizationIDs: []string{user.Edges.Setting.Edges.DefaultOrg.ID}}).MustNew(reqCtx, t)

	resp, err := suite.Client.API.DeleteUser(reqCtx, user.ID)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.DeleteUser.DeletedID != "")

	// make sure the deletedID matches the ID we wanted to delete
	assert.Check(t, is.Equal(user.ID, resp.DeleteUser.DeletedID))

	_, err = suite.Client.API.GetUserByID(reqCtx, user.ID)

	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	_, err = suite.Client.API.GetPersonalAccessTokenByID(reqCtx, token.ID)

	assert.ErrorContains(t, err, th.NotFoundErrorMsg)
}
