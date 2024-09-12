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
	"github.com/rShetty/asyncwait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/emails/mock"

	"github.com/theopenlane/utils/echocontext"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestResetPasswordHandler() {
	t := suite.T()

	// setup request request
	suite.e.POST("password-reset", suite.h.ResetPassword)

	ec := echocontext.NewTestEchoContext().Request().Context()

	var newPassword = "6z9Fqc-E-9v32NsJzLNU" //nolint:gosec

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)

	testCases := []struct {
		name                 string
		email                string
		newPassword          string
		tokenSet             bool
		tokenProvided        string
		ttl                  string
		emailExpected        bool
		expectedEmailSubject string
		expectedResp         string
		expectedStatus       int
		from                 string
	}{
		{
			name:                 "happy path",
			email:                "kelsier@theopenlane.io",
			tokenSet:             true,
			newPassword:          newPassword,
			from:                 "mitb@theopenlane.io",
			emailExpected:        true,
			expectedEmailSubject: emails.PasswordResetSuccessRE,
			expectedResp:         emptyResponse,
			expectedStatus:       http.StatusOK,
		},
		{
			name:           "bad token (user not found)",
			email:          "eventure@theopenlane.io",
			tokenSet:       true,
			tokenProvided:  "thisisnotavalidtoken",
			newPassword:    newPassword,
			emailExpected:  false,
			from:           "notactuallyanemail",
			expectedResp:   "password reset token invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "weak password",
			email:          "sazed@theopenlane.io",
			tokenSet:       true,
			newPassword:    "weak1",
			emailExpected:  false,
			from:           "nottodaysatan",
			expectedResp:   "password is too weak",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "same password",
			email:          "sventure@theopenlane.io",
			tokenSet:       true,
			newPassword:    validPassword,
			emailExpected:  false,
			from:           "mmhmm",
			expectedResp:   "password was already used",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing token",
			email:          "dockson@theopenlane.io",
			tokenSet:       false,
			newPassword:    newPassword,
			emailExpected:  false,
			from:           "yadayadayada",
			expectedResp:   "token is required",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "expired reset token",
			email:          "tensoon@theopenlane.io",
			newPassword:    "6z9Fqc-E-9v32NsJzLNP",
			tokenSet:       true,
			emailExpected:  false,
			from:           "zonkertons",
			ttl:            expiredTTL,
			expectedResp:   "reset token is expired, please request a new token using forgot-password",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sent := time.Now()

			mock.ResetEmailMock()

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
			}

			// Test that one verify email was sent to each user
			messages := []*mock.EmailMetadata{
				{
					To:        tc.email,
					From:      tc.from,
					Subject:   tc.expectedEmailSubject,
					Timestamp: sent,
				},
			}

			// wait for messages
			predicate := func() bool {
				return suite.h.TaskMan.GetQueueLength() == 0
			}
			successful := asyncwait.NewAsyncWait(maxWaitInMillis, pollIntervalInMillis).Check(predicate)

			if successful != true {
				t.Errorf("max wait of email send")
			}

			if tc.emailExpected {
				mock.CheckEmails(t, messages)
			} else {
				mock.CheckEmails(t, nil)
			}
		})
	}
}

// createUserWithResetToken creates a user with a valid reset token and returns the token, user id, and error if one occurred
func (suite *HandlerTestSuite) createUserWithResetToken(t *testing.T, ec context.Context, email string, ttl string) (*ent.PasswordResetToken, string, error) {
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	// add mocks for writes
	mock_fga.WriteAny(t, suite.fga)

	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	u := suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail(email).
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

	// clear mocks
	mock_fga.ClearMocks(suite.fga)

	return pr, u.ID, nil
}
