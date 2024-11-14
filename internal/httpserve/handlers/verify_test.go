package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestVerifyHandler() {
	t := suite.T()

	// add handler
	suite.e.GET("verify", suite.h.VerifyEmail)

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
			expectedMessage: "Token expired, a new token has been issued. Please try again",
			expectedStatus:  http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			} else {
				assert.Contains(t, out.Error, tc.expectedMessage)
			}
		})
	}
}
