package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestForgotPasswordHandler(t *testing.T) {
	suite := suite.setupTest(t)

	// Register test handler
	suite.registerTestHandler("POST", "forgot-password", suite.h.ForgotPassword)

	ec := echocontext.NewTestEchoContext().Request().Context()

	// create user in the database
	ctx := auth.WithInternalOperationContext(ec)

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
				assert.NilError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectedStatus != http.StatusOK {
				var out *models.ForgotPasswordResponse

				// parse request body
				if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
					t.Error("error parsing response", err)
				}

				assert.Check(t, is.Contains(out.Error, tc.expectedErrMessage))
				assert.Check(t, !out.Success)
			}

			// verify email was sent through the mock sender
			suite.WaitForEvents()

			msgs := suite.mockEmailSender().Messages()
			if tc.emailExpected {
				assert.Assert(t, is.Len(msgs, 1))
				assert.Check(t, is.DeepEqual([]string{tc.email}, msgs[0].To))
				assert.Check(t, is.Contains(msgs[0].Subject, "Password Reset - Action Required"))
			} else {
				assert.Check(t, is.Len(msgs, 0))
			}
		})
	}
}
