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

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/shared/enums"
	models "github.com/theopenlane/shared/openapi"
)

func (suite *HandlerTestSuite) TestForgotPasswordHandler() {
	t := suite.T()

	// Register test handler with OpenAPI context
	suite.registerTestHandler("POST", "forgot-password", suite.createImpersonationOperation("ForgotPassword", "Test forgot password"), suite.h.ForgotPassword)

	ec := echocontext.NewTestEchoContext().Request().Context()

	// create user in the database
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(false).
		SaveX(ctx)

	_ = suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail("asandler@theopenlane.io").
		SetPassword(validPassword).
		SetSetting(userSetting).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SaveX(ctx)

	var mitb = "mitb@theopenlane.io"

	testCases := []struct {
		name               string
		from               string
		email              string
		emailExpected      bool
		expectedErrMessage string
		expectedStatus     int
	}{
		{
			name:           "happy path",
			email:          "asandler@theopenlane.io",
			from:           mitb,
			emailExpected:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "email does not exist, should still return 200",
			email:          "asandler1@theopenlane.io",
			from:           mitb,
			emailExpected:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:               "email not sent in request",
			from:               mitb,
			emailExpected:      false,
			expectedStatus:     http.StatusBadRequest,
			expectedErrMessage: "email is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.ClearTestData()

			resendJSON := models.ForgotPasswordRequest{
				Email: tc.email,
			}

			body, err := json.Marshal(resendJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus != http.StatusOK {
				var out *models.ForgotPasswordReply

				// parse request body
				if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
					t.Error("error parsing response", err)
				}

				assert.Contains(t, out.Error, tc.expectedErrMessage)
				assert.False(t, out.Success)
			}

			// ensure email was added to the job queue
			if tc.emailExpected {
				job := rivertest.RequireInserted[*riverpgxv5.Driver](context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
				require.NotNil(t, job)
				assert.Equal(t, []string{tc.email}, job.Args.Message.To)
				assert.Contains(t, job.Args.Message.Subject, "Password Reset - Action Required")
			} else {
				rivertest.RequireNotInserted(ctx, t, riverpgxv5.New(suite.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
			}
		})
	}
}
