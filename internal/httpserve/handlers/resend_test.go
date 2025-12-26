package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/common/enums"
	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestResendHandler() {
	t := suite.T()

	// add handler
	// Create operation for ResendEmail
	operation := suite.createImpersonationOperation("ResendEmail", "Resend email verification")
	suite.registerTestHandler("POST", "resend", operation, suite.h.ResendEmail)

	ec := echocontext.NewTestEchoContext().Request().Context()

	ctx := privacy.DecisionContext(ec, privacy.Allow)

	// create user in the database
	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(false).
		SaveX(ctx)

	_ = suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail("bsanderson@theopenlane.io").
		SetPassword(validPassword).
		SetSetting(userSetting).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SaveX(ctx)

	// create user in the database
	userSetting2 := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	_ = suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail("dabraham@theopenlane.io").
		SetPassword(validPassword).
		SetSetting(userSetting2).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SaveX(ctx)

	testCases := []struct {
		name            string
		email           string
		expectedMessage string
		expectedStatus  int
	}{
		{
			name:            "happy path",
			email:           "bsanderson@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "received your request to be resend",
		},
		{
			name:            "happy path, attempt 2",
			email:           "bsanderson@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "received your request to be resend",
		},
		{
			name:            "happy path, attempt 3",
			email:           "bsanderson@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "received your request to be resend",
		},
		{
			name:            "happy path, attempt 4",
			email:           "bsanderson@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "received your request to be resend",
		},
		{
			name:            "happy path, attempt 5",
			email:           "bsanderson@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "received your request to be resend",
		},
		{
			name:            "happy path, attempt 6 should fail",
			email:           "bsanderson@theopenlane.io",
			expectedStatus:  http.StatusTooManyRequests,
			expectedMessage: "max attempts",
		},
		{
			name:            "email does not exist, should still return 204",
			email:           "bsanderson1@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "received your request to be resend",
		},
		{
			name:            "email confirmed",
			email:           "dabraham@theopenlane.io",
			expectedStatus:  http.StatusOK,
			expectedMessage: "email is already confirmed",
		},
		{
			name:           "email not sent in request",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resendJSON := models.ResendRequest{
				Email: tc.email,
			}

			body, err := json.Marshal(resendJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/resend", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.ResendReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			require.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				require.NotEmpty(t, out)
				assert.NotEmpty(t, out.Message)
			} else {
				assert.Contains(t, out.Error, tc.expectedMessage)
			}
		})
	}
}
