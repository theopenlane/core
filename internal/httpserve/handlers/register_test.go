package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated/emailverificationtoken"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRegisterHandler(t *testing.T) {
	suite := suite.setupTest(t).withTestUsers().withAPIClient()

	// Register test handler
	suite.registerTestHandler("POST", "register", suite.h.RegisterHandler)
	suite.registerTestHandler("GET", "verify", suite.h.VerifyEmail)

	var bonkers = "b!a!n!a!n!a!s!"

	testCases := []struct {
		name               string
		email              string
		firstName          string
		lastName           string
		password           string
		emailExpected      bool
		expectedErrMessage string
		expectedErrorCode  rout.ErrorCode
		expectedStatus     int
		invitationType     string
		verifyEmail        bool
	}{
		{
			name:           "happy path",
			email:          "bananas@theopenlane.io",
			firstName:      "Princess",
			lastName:       "Fiona",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "happy path, no first and last name",
			email:          "oranges@theopenlane.io",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "happy path, first name only",
			email:          "berries@theopenlane.io",
			firstName:      "Princess",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "happy path, last name only",
			email:          "melon@theopenlane.io",
			lastName:       "Fiona",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "happy path, mixed case email can be verified",
			email:          "Mixed.Case@TheOpenlane.io",
			firstName:      "Princess",
			lastName:       "Fiona",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
			verifyEmail:    true,
		},
		{
			name:               "invalid email",
			email:              "bananas.net",
			firstName:          "Princess",
			lastName:           "Fiona",
			password:           bonkers,
			emailExpected:      false,
			expectedErrMessage: "email was invalid",
			expectedStatus:     http.StatusBadRequest,
			expectedErrorCode:  handlers.InvalidInputErrCode,
		},
		{
			name:               "missing email",
			firstName:          "Princess",
			lastName:           "Fiona",
			password:           bonkers,
			emailExpected:      false,
			expectedErrMessage: "missing required field: email",
			expectedStatus:     http.StatusBadRequest,
			expectedErrorCode:  handlers.InvalidInputErrCode,
		},
		{
			name:               "bad password",
			email:              "pancakes@theopenlane.io",
			firstName:          "Princess",
			lastName:           "Fiona",
			password:           "asfghjkl",
			emailExpected:      false,
			expectedErrMessage: "password is too weak",
			expectedStatus:     http.StatusBadRequest,
			expectedErrorCode:  handlers.InvalidInputErrCode,
		},
		{
			name:              "already registered",
			email:             "bananas@theopenlane.io",
			firstName:         "Princess",
			lastName:          "Fiona",
			password:          bonkers,
			emailExpected:     false,
			expectedStatus:    http.StatusConflict,
			expectedErrorCode: handlers.UserExistsErrCode,
		},
		{
			name:           "happy path with valid invitation token",
			email:          "invitee@theopenlane.io",
			firstName:      "Invited",
			lastName:       "User",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
			invitationType: "invitation",
		},
		{
			name:           "happy path with invitation token, no first/last name",
			email:          "invitee2@theopenlane.io",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
			invitationType: "invitation",
		},
		{
			name:               "invalid invitation token",
			email:              "invitee3@theopenlane.io",
			firstName:          "Invalid",
			lastName:           "Token",
			password:           bonkers,
			emailExpected:      true,
			expectedStatus:     http.StatusBadRequest,
			expectedErrMessage: "invite not found",
			invitationType:     "invalid_invitation",
		},
		{
			name:               "email mismatch with invitation token",
			email:              "wrongemail@theopenlane.io",
			firstName:          "Wrong",
			lastName:           "Email",
			password:           bonkers,
			emailExpected:      true,
			expectedErrMessage: "could not verify email",
			expectedStatus:     http.StatusBadRequest,
			invitationType:     "email_mismatch_invitation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.ClearTestData()

			// Make "already registered" independent from subtest order by using a unique email
			// and seeding the pre-existing user within the subtest.
			email := tc.email
			if tc.expectedErrorCode == handlers.UserExistsErrCode {
				email = ulids.New().String() + "@theopenlane.io"
				ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
				exists, err := suite.db.User.Query().Where(user.Email(email)).Exist(ctx)
				assert.NilError(t, err)
				if !exists {
					_ = suite.userBuilderWithInput(ctx, &userInput{
						email:         email,
						password:      tc.password,
						confirmedUser: true,
					})
				}
			}

			var inviteToken *string

			if tc.invitationType == "invitation" {
				ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
				invite := suite.db.Invite.Create().
					SetRecipient(email).
					SetRole(enums.RoleMember).
					SaveX(ctx)
				inviteToken = &invite.Token
			} else if tc.invitationType == "invalid_invitation" {

				ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
				_ = suite.db.Invite.Create().
					SetRecipient(email).
					SetRole(enums.RoleMember).
					SaveX(ctx)

				// create token still but use another one
				invalidToken := "invalid-token-123"
				inviteToken = &invalidToken
			} else if tc.invitationType == "email_mismatch_invitation" {
				ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
				invite := suite.db.Invite.Create().
					SetRecipient("correctemail@theopenlane.io").
					SetRole(enums.RoleMember).
					SaveX(ctx)
				inviteToken = &invite.Token
			}

			registerJSON := models.RegisterRequest{
				FirstName: tc.firstName,
				LastName:  tc.lastName,
				Email:     email,
				Password:  tc.password,
				Token:     inviteToken,
			}

			body, err := json.Marshal(registerJSON)
			if err != nil {
				assert.NilError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.RegisterResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))
			assert.Check(t, is.DeepEqual(tc.expectedErrorCode, out.ErrorCode))

			if tc.expectedStatus == http.StatusCreated {
				assert.Check(t, is.Equal(out.Email, strings.ToLower(email)))
				assert.Check(t, len(out.Message) > 0)
				assert.Check(t, len(out.ID) > 0)

				// setup context to get the data back
				ctx := auth.NewTestContextWithValidUser(out.ID)

				// we haven't set the user's default org yet in the context
				// so run the lookup as an internal operation
				ctx = auth.WithInternalOperationContext(ctx)

				// get the user and make sure things were created as expected
				u, err := suite.db.UserSetting.Query().Where(usersetting.UserID(out.ID)).WithDefaultOrg().Only(ctx)
				assert.NilError(t, err)

				assert.Check(t, u.Edges.DefaultOrg != nil)
				assert.Assert(t, len(u.Edges.DefaultOrg.ID) > 0)

				if tc.invitationType == "invitation" {
					// The user should be added to the organization that sent the invite (suite.testUser1's org)
					assert.Check(t, is.DeepEqual(suite.testUser1.OrganizationID, u.Edges.DefaultOrg.ID))

					// setup context with the joined org
					ctx = auth.NewTestContextWithOrgID(out.ID, suite.testUser1.OrganizationID)

					// make sure user is a member of the organization they were invited to
					orgMemberships, err := suite.api.GetOrgMembersByOrgID(ctx, &testclient.OrgMembershipWhereInput{
						OrganizationID: &suite.testUser1.OrganizationID,
					})
					assert.NilError(t, err)

					// find the membership for this user
					found := false
					for _, edge := range orgMemberships.OrgMemberships.Edges {
						if edge.Node.UserID == out.ID {
							assert.Check(t, is.DeepEqual(enums.RoleMember, edge.Node.Role))
							found = true
							break
						}
					}
					assert.Check(t, found, "user should be a member of the invited organization")
				} else {

					// for regular registration, setup context with user's personal org
					ctx = auth.NewTestContextWithOrgID(out.ID, u.Edges.DefaultOrg.ID)

					// make sure user is an owner of their personal org
					orgMemberships, err := suite.api.GetOrgMembersByOrgID(ctx, &testclient.OrgMembershipWhereInput{
						OrganizationID: &u.Edges.DefaultOrg.ID,
					})
					assert.NilError(t, err)
					assert.Assert(t, is.Len(orgMemberships.OrgMemberships.Edges, 1))
					assert.Check(t, is.DeepEqual(orgMemberships.OrgMemberships.Edges[0].Node.Role, enums.RoleOwner))
				}

				// get user to test display name
				user, err := suite.db.User.Get(ctx, out.ID)
				assert.NilError(t, err)

				// if name is set, it's used for display name, otherwise it's the email prefix
				if tc.firstName != "" && tc.lastName != "" {
					assert.Check(t, is.Equal(strings.TrimSpace(tc.firstName+" "+tc.lastName), user.DisplayName))
				} else {
					assert.Check(t, is.DeepEqual(strings.Split(tc.email, "@")[0], user.DisplayName))
				}
			} else {
				assert.Check(t, is.Contains(out.Error, tc.expectedErrMessage))
			}

			// verify emails sent through mock sender
			suite.WaitForEvents()

			msgs := suite.mockEmailSender().Messages()
			if tc.emailExpected {
				switch {
				case tc.invitationType == "invitation":
					// invite + registration + invite acceptance
					assert.Assert(t, is.Len(msgs, 3))
				case strings.Contains(tc.invitationType, "invitation"):
					// only the invitation email
					assert.Assert(t, is.Len(msgs, 1))
				default:
					// regular registration — verification email
					assert.Assert(t, len(msgs) > 0)
					assert.Check(t, is.Contains(msgs[0].Subject, "Please verify your email address"))
					assert.Check(t, is.DeepEqual([]string{strings.ToLower(tc.email)}, msgs[0].To))
				}
			} else {
				assert.Check(t, is.Len(msgs, 0))
			}

			if tc.verifyEmail {
				ctx := auth.WithInternalOperationContext(auth.NewTestContextWithValidUser(out.ID))

				evToken, err := suite.db.EmailVerificationToken.Query().Where(emailverificationtoken.OwnerID(out.ID)).Only(ctx)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(strings.ToLower(email), evToken.Email))

				verifyReq := httptest.NewRequest(http.MethodGet, "/verify?token="+evToken.Token, nil)
				verifyRecorder := httptest.NewRecorder()

				suite.e.ServeHTTP(verifyRecorder, verifyReq)

				var verifyOut *models.VerifyResponse
				assert.NilError(t, json.NewDecoder(verifyRecorder.Body).Decode(&verifyOut))

				assert.Check(t, is.Equal(http.StatusOK, verifyRecorder.Code))
				assert.Check(t, verifyOut.Success)
				assert.Check(t, is.Equal(strings.ToLower(email), verifyOut.Email))
			}
		})
	}
}

func TestRegisterHandler_EmailVerification(t *testing.T) {
	suite := suite.setupTest(t)

	// enable email verification for this test
	config := &validator.EmailVerificationConfig{
		Enabled: true,
		AllowedEmailTypes: validator.AllowedEmailTypes{
			Disposable: true,
			Free:       true,
			Role:       true,
		},
	}

	original := suite.db.EmailVerifier
	originalValidation := suite.db.EntConfig.EmailValidation
	suite.db.EmailVerifier = config.NewVerifier()

	t.Cleanup(func() {
		suite.db.EmailVerifier = original
		suite.db.EntConfig.EmailValidation = originalValidation
	})

	// Register test handler
	suite.registerTestHandler("POST", "register", suite.h.RegisterHandler)

	testCases := []struct {
		name              string
		email             string
		config            validator.EmailVerificationConfig
		expectedStatus    int
		expectedErr       string
		expectedErrorCode rout.ErrorCode
	}{
		{
			name:  "happy path, free allowed",
			email: "user@gmail.com",
			config: validator.EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: validator.AllowedEmailTypes{
					Disposable: true,
					Free:       true,
					Role:       true,
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "invalid domain",
			email: "test@company.com",
			config: validator.EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: validator.AllowedEmailTypes{
					Disposable: true,
					Free:       true,
					Role:       true,
				},
			},
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: handlers.InvalidInputErrCode,
			expectedErr:       validator.ErrEmailNotAllowed.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.db.EntConfig.EmailValidation = tc.config

			registerJSON := models.RegisterRequest{
				Email:    tc.email,
				Password: gofakeit.Password(true, true, true, true, false, 12),
			}

			body, err := json.Marshal(registerJSON)
			if err != nil {
				assert.NilError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.RegisterResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))
			assert.Check(t, is.DeepEqual(tc.expectedErrorCode, out.ErrorCode))

			if tc.expectedStatus == http.StatusCreated {
				assert.Check(t, is.DeepEqual(out.Email, tc.email))

			} else {
				assert.Check(t, is.Contains(out.Error, tc.expectedErr))
			}

		})
	}
}
