package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func (suite *HandlerTestSuite) TestRegisterHandler() {
	t := suite.T()

	// Register test handler with OpenAPI context
	suite.registerTestHandler("POST", "register", suite.createImpersonationOperation("RegisterHandler", "Test register"), suite.h.RegisterHandler)

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

			var inviteToken *string

			if tc.invitationType == "invitation" {
				ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
				invite := suite.db.Invite.Create().
					SetRecipient(tc.email).
					SetRole(enums.RoleMember).
					SaveX(ctx)
				inviteToken = &invite.Token
			} else if tc.invitationType == "invalid_invitation" {

				ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
				_ = suite.db.Invite.Create().
					SetRecipient(tc.email).
					SetRole(enums.RoleMember).
					SaveX(ctx)

				// create token still but use another one
				invalidToken := "invalid-token-123"
				inviteToken = &invalidToken
			} else if tc.invitationType == "email_mismatch_invitation" {
				ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
				invite := suite.db.Invite.Create().
					SetRecipient("correctemail@theopenlane.io").
					SetRole(enums.RoleMember).
					SaveX(ctx)
				inviteToken = &invite.Token
			}

			registerJSON := models.RegisterRequest{
				FirstName: tc.firstName,
				LastName:  tc.lastName,
				Email:     tc.email,
				Password:  tc.password,
				Token:     inviteToken,
			}

			body, err := json.Marshal(registerJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.RegisterReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedErrorCode, out.ErrorCode)

			if tc.expectedStatus == http.StatusCreated {
				assert.Equal(t, out.Email, tc.email)
				assert.NotEmpty(t, out.Message)
				assert.NotEmpty(t, out.ID)

				// setup context to get the data back
				ctx := auth.NewTestContextWithValidUser(out.ID)

				// we haven't set the user's default org yet in the context
				// so allow the request to go through
				ctx = privacy.DecisionContext(ctx, privacy.Allow)

				// get the user and make sure things were created as expected
				u, err := suite.db.UserSetting.Query().Where(usersetting.UserID(out.ID)).WithDefaultOrg().Only(ctx)
				require.NoError(t, err)

				assert.NotEmpty(t, u.Edges.DefaultOrg)
				require.NotEmpty(t, u.Edges.DefaultOrg.ID)

				if tc.invitationType == "invitation" {
					// The user should be added to the organization that sent the invite (testUser1's org)
					assert.Equal(t, testUser1.OrganizationID, u.Edges.DefaultOrg.ID)

					// setup context with the joined org
					ctx = auth.NewTestContextWithOrgID(out.ID, testUser1.OrganizationID)

					// make sure user is a member of the organization they were invited to
					orgMemberships, err := suite.api.GetOrgMembersByOrgID(ctx, &testclient.OrgMembershipWhereInput{
						OrganizationID: &testUser1.OrganizationID,
					})
					require.NoError(t, err)

					// find the membership for this user
					found := false
					for _, edge := range orgMemberships.OrgMemberships.Edges {
						if edge.Node.UserID == out.ID {
							assert.Equal(t, enums.RoleMember, edge.Node.Role)
							found = true
							break
						}
					}
					assert.True(t, found, "user should be a member of the invited organization")
				} else {

					// for regular registration, setup context with user's personal org
					ctx = auth.NewTestContextWithOrgID(out.ID, u.Edges.DefaultOrg.ID)

					// make sure user is an owner of their personal org
					orgMemberships, err := suite.api.GetOrgMembersByOrgID(ctx, &testclient.OrgMembershipWhereInput{
						OrganizationID: &u.Edges.DefaultOrg.ID,
					})
					require.NoError(t, err)
					require.Len(t, orgMemberships.OrgMemberships.Edges, 1)
					assert.Equal(t, orgMemberships.OrgMemberships.Edges[0].Node.Role, enums.RoleOwner)
				}

				// get user to test display name
				user, err := suite.db.User.Get(ctx, out.ID)
				require.NoError(t, err)

				// if name is set, it's used for display name, otherwise it's the email prefix
				if tc.firstName != "" && tc.lastName != "" {
					assert.Equal(t, strings.TrimSpace(tc.firstName+" "+tc.lastName), user.DisplayName)
				} else {
					assert.Equal(t, strings.Split(tc.email, "@")[0], user.DisplayName)
				}
			} else {
				assert.Contains(t, out.Error, tc.expectedErrMessage)
			}

			// wait for messages
			if tc.emailExpected {
				if tc.invitationType == "invitation" {
					// the actual invite email
					// user registration email
					// user invite acceptance email
					job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
						[]rivertest.ExpectedJob{
							{
								Args: jobs.EmailArgs{},
							},
							{
								Args: jobs.EmailArgs{},
							},
							{
								Args: jobs.EmailArgs{},
							},
						})
					require.NotNil(t, job)

				} else if strings.Contains(tc.invitationType, "invitation") {
					// only the invitation email is sent
					job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
						[]rivertest.ExpectedJob{
							{
								Args: jobs.EmailArgs{},
							},
						})
					require.NotNil(t, job)

				} else {
					// For regular registration, expect both verification and welcome emails
					job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
						[]rivertest.ExpectedJob{
							{
								Args: jobs.EmailArgs{
									Message: *newman.NewEmailMessageWithOptions(
										newman.WithSubject("Please verify your email address to login to Meow Inc."),
										newman.WithTo([]string{tc.email}),
									),
								},
							},
						})
					require.NotNil(t, job)
				}
			} else {
				rivertest.RequireNotInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestRegisterHandler_EmailVerification() {
	t := suite.T()

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
	suite.db.EmailVerifier = config.NewVerifier()

	// Register test handler with OpenAPI context
	suite.registerTestHandler("POST", "register", suite.createImpersonationOperation("RegisterHandler", "Test register"), suite.h.RegisterHandler)

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
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.RegisterReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedErrorCode, out.ErrorCode)

			if tc.expectedStatus == http.StatusCreated {
				assert.Equal(t, out.Email, tc.email)

			} else {
				assert.Contains(t, out.Error, tc.expectedErr)
			}

		})
	}

	// reset back to original state
	suite.db.EmailVerifier = original
}
