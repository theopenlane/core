package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/echox/middleware/echocontext"

	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/privacy"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/enums"
	models "github.com/theopenlane/core/pkg/openapi"
)

func (suite *HandlerTestSuite) TestResetPasswordHandler() {
	t := suite.T()

	// setup request request
	// Create operation for ResetPassword
	operation := suite.createImpersonationOperation("ResetPassword", "Reset user password")
	suite.registerTestHandler("POST", "password-reset", operation, suite.h.ResetPassword)

	ec := echocontext.NewTestEchoContext().Request().Context()

	var newPassword = "6z9Fqc-E-9v32NsJzLNU" //nolint:gosec

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)

	testCases := []struct {
		name           string
		email          string
		newPassword    string
		tokenSet       bool
		tokenProvided  string
		ttl            string
		expectedResp   string
		expectedStatus int
		from           string
	}{
		{
			name:           "happy path",
			email:          "kelsier@theopenlane.io",
			tokenSet:       true,
			newPassword:    newPassword,
			from:           "mitb@theopenlane.io",
			expectedResp:   emptyResponse,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "bad token (user not found)",
			email:          "eventure@theopenlane.io",
			tokenSet:       true,
			tokenProvided:  "thisisnotavalidtoken",
			newPassword:    newPassword,
			from:           "notactuallyanemail",
			expectedResp:   "password reset token invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "weak password",
			email:          "sazed@theopenlane.io",
			tokenSet:       true,
			newPassword:    "weak1",
			from:           "nottodaysatan",
			expectedResp:   "password is too weak",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "same password",
			email:          "sventure@theopenlane.io",
			tokenSet:       true,
			newPassword:    validPassword,
			from:           "mmhmm",
			expectedResp:   "password was already used",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing token",
			email:          "dockson@theopenlane.io",
			tokenSet:       false,
			newPassword:    newPassword,
			from:           "yadayadayada",
			expectedResp:   "token is required",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "expired reset token",
			email:          "tensoon@theopenlane.io",
			newPassword:    "6z9Fqc-E-9v32NsJzLNP",
			tokenSet:       true,
			from:           "zonkertons",
			ttl:            expiredTTL,
			expectedResp:   "reset token is expired, please request a new token using forgot-password",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create user in the database
			rt, _, err := suite.createUserWithResetToken(t, ec, tc.email, tc.ttl)
			require.NoError(t, err)

			pwResetJSON := models.ResetPasswordRequest{
				Password: tc.newPassword,
			}

			if tc.tokenSet {
				pwResetJSON.Token = rt.Token
				if tc.tokenProvided != "" {
					pwResetJSON.Token = tc.tokenProvided
				}
			}

			body, err := json.Marshal(pwResetJSON)
			if err != nil {
				require.NoError(t, err)
			}

			suite.ClearTestData()

			req := httptest.NewRequest(http.MethodPost, "/password-reset", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			// get result
			res := recorder.Result()
			defer res.Body.Close()

			// check status
			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var out *models.ResetPasswordReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.expectedStatus != http.StatusOK {
				assert.Contains(t, out.Error, tc.expectedResp)
			} else {
				job := rivertest.RequireInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()), &jobs.EmailArgs{
					Message: *newman.NewEmailMessageWithOptions(
						newman.WithSubject("Openlane Password Reset - Action Required"),
					),
				}, nil)
				require.NotNil(t, job)
				require.Equal(t, []string{tc.email}, job.Args.Message.To)
			}
		})
	}
}

// createUserWithResetToken creates a user with a valid reset token and returns the token, user id, and error if one occurred
func (suite *HandlerTestSuite) createUserWithResetToken(t *testing.T, ec context.Context, email string, ttl string) (*ent.PasswordResetToken, string, error) {
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	u := suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail(email).
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
	if err := user.CreatePasswordResetToken(); err != nil {
		return nil, "", err
	}

	if ttl != "" {
		user.PasswordResetExpires.String = ttl
	}

	// set expiry if provided in test case
	expiry, err := time.Parse(time.RFC3339Nano, user.PasswordResetExpires.String)
	if err != nil {
		return nil, "", err
	}

	// store token in db
	pr := suite.db.PasswordResetToken.Create().
		SetOwner(u).
		SetToken(user.PasswordResetToken.String).
		SetEmail(user.Email).
		SetSecret(user.PasswordResetSecret).
		SetTTL(expiry).
		SaveX(ctx)

	return pr, u.ID, nil
}
