package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/enums"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestVerifyHandler() {
	t := suite.T()

	// add handler
	// Create operation for VerifyEmail
	operation := suite.createImpersonationOperation("VerifyEmail", "Verify email address")
	suite.registerTestHandler("GET", "verify", operation, suite.h.VerifyEmail)

	ec := echocontext.NewTestEchoContext().Request().Context()

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)

	testCases := []struct {
		name            string
		userConfirmed   bool
		email           string
		ttl             string
		tokenSet        bool
		expectedMessage string
		expectedStatus  int
	}{
		{
			name:            "happy path, unconfirmed user",
			userConfirmed:   false,
			email:           "mitb@theopenlane.io",
			tokenSet:        true,
			expectedMessage: "success",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "happy path, already confirmed user",
			userConfirmed:   true,
			email:           "sitb@theopenlane.io",
			tokenSet:        true,
			expectedMessage: "success",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "missing token",
			userConfirmed:   true,
			email:           "santa@theopenlane.io",
			tokenSet:        false,
			expectedMessage: "token is required",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "expired token, but not already confirmed",
			userConfirmed:   false,
			email:           "elf@theopenlane.io",
			tokenSet:        true,
			ttl:             expiredTTL,
			expectedMessage: "Token expired, a new token has been issued. Please check your email and try again.",
			expectedStatus:  http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.ClearTestData()

			// set privacy allow in order to allow the creation of the users without
			// authentication in the tests
			ctx := privacy.DecisionContext(ec, privacy.Allow)

			// create user in the database
			userSetting := suite.db.UserSetting.Create().
				SetEmailConfirmed(tc.userConfirmed).
				SaveX(ctx)

			u := suite.db.User.Create().
				SetFirstName(gofakeit.FirstName()).
				SetLastName(gofakeit.LastName()).
				SetEmail(tc.email).
				SetPassword(validPassword).
				SetSetting(userSetting).
				SetLastLoginProvider(enums.AuthProviderCredentials).
				SetLastSeen(time.Now()).
				SaveX(ctx)

			user := handlers.User{
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Email:     u.Email,
				ID:        u.ID,
			}

			// create token
			if err := user.CreateVerificationToken(); err != nil {
				require.NoError(t, err)
			}

			if tc.ttl != "" {
				user.EmailVerificationExpires.String = tc.ttl
			}

			ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
			if err != nil {
				require.NoError(t, err)
			}

			// store token in db
			allowCtx := privacy.DecisionContext(ec, privacy.Allow)
			et := suite.db.EmailVerificationToken.Create().
				SetOwner(u).
				SetToken(user.EmailVerificationToken.String).
				SetEmail(user.Email).
				SetSecret(user.EmailVerificationSecret).
				SetTTL(ttl).
				SaveX(allowCtx)

			target := "/verify"
			if tc.tokenSet {
				target = fmt.Sprintf("/verify?token=%s", et.Token)
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var out *models.VerifyReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.expectedStatus >= http.StatusOK && tc.expectedStatus <= http.StatusCreated {
				assert.Contains(t, out.Message, tc.expectedMessage)

				if tc.userConfirmed {
					// check the claims to ensure the the claims are set correctly
					token, _, err := new(jwt.Parser).ParseUnverified(out.AccessToken, jwt.MapClaims{})
					require.NoError(t, err)

					claims, ok := token.Claims.(jwt.MapClaims)
					require.True(t, ok)

					assert.NotEmpty(t, claims["org"])
				} else {
					job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
						[]rivertest.ExpectedJob{
							{
								Args: jobs.EmailArgs{
									Message: *newman.NewEmailMessageWithOptions(
										newman.WithSubject("Welcome to Meow Inc.!"),
										newman.WithTo([]string{tc.email}),
									),
								},
							},
						})
					require.NotNil(t, job)
				}
			} else {
				assert.Contains(t, out.Error, tc.expectedMessage)
			}
		})
	}
}
