package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestResendHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// add handler
	suite.registerTestHandler("POST", "resend", suite.h.ResendEmail)

	ec := echocontext.NewTestEchoContext().Request().Context()

	ctx := auth.WithInternalOperationContext(ec)

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

	// Helper to keep request/response handling consistent across subtests.
	send := func(t *testing.T, email string) (*httptest.ResponseRecorder, *models.ResendResponse) {
		resendJSON := models.ResendRequest{
			Email: email,
		}

		body, err := json.Marshal(resendJSON)
		if err != nil {
			assert.NilError(t, err)
		}

		req := httptest.NewRequest(http.MethodPost, "/resend", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		// Set writer for tests that write on the response
		recorder := httptest.NewRecorder()

		// Using the ServerHTTP on echo will trigger the router and middleware
		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()

		var out *models.ResendResponse
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			t.Error("error parsing response", err)
		}

		return recorder, out
	}

	// Exercise the sequential rate-limit behavior in one subtest to avoid
	// order-dependence across table entries.
	t.Run("rate limiting after max attempts", func(t *testing.T) {
		email := "bsanderson@theopenlane.io"

		for i := 1; i <= 6; i++ {
			recorder, out := send(t, email)

			if i <= 5 {
				assert.Assert(t, is.Equal(http.StatusOK, recorder.Code))
				assert.Assert(t, out != nil)
				assert.Check(t, len(out.Message) > 0)
			} else {
				assert.Assert(t, is.Equal(http.StatusTooManyRequests, recorder.Code))
				assert.Assert(t, out != nil)
				assert.Check(t, is.Contains(out.Error, "max attempts"))
			}
		}
	})

	testCases := []struct {
		name            string
		email           string
		expectedMessage string
		expectedStatus  int
	}{
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
			recorder, out := send(t, tc.email)

			assert.Assert(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectedStatus == http.StatusOK {
				assert.Assert(t, out != nil)
				assert.Check(t, len(out.Message) > 0)
			} else {
				assert.Check(t, is.Contains(out.Error, tc.expectedMessage))
			}
		})
	}
}
