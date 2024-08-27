package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/pkg/middleware/echocontext"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestLoginHandler() {
	t := suite.T()

	// add login handler
	suite.e.POST("login", suite.h.LoginHandler)

	ec := echocontext.NewTestEchoContext().Request().Context()

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	// add mocks for writes
	mock_fga.WriteAny(t, suite.fga)

	// create user in the database
	validConfirmedUser := "rsanchez@theopenlane.io"
	validPassword := "sup3rs3cu7e!"

	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	_ = suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail(validConfirmedUser).
		SetPassword(validPassword).
		SetSetting(userSetting).
		SaveX(ctx)

	validUnconfirmedUser := "msmith@theopenlane.io"

	userSetting = suite.db.UserSetting.Create().
		SetEmailConfirmed(false).
		SaveX(ctx)

	_ = suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail(validUnconfirmedUser).
		SetPassword(validPassword).
		SetSetting(userSetting).
		SaveX(ctx)

	testCases := []struct {
		name           string
		username       string
		password       string
		expectedErr    error
		expectedStatus int
	}{
		{
			name:           "happy path, valid credentials",
			username:       validConfirmedUser,
			password:       validPassword,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "email unverified",
			username:       validUnconfirmedUser,
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    auth.ErrUnverifiedUser,
		},
		{
			name:           "invalid password",
			username:       validConfirmedUser,
			password:       "thisisnottherightone",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.ErrInvalidCredentials,
		},
		{
			name:           "user not found",
			username:       "rick.sanchez@theopenlane.io",
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    auth.ErrNoAuthUser,
		},
		{
			name:           "empty username",
			username:       "",
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.NewMissingRequiredFieldError("username"),
		},
		{
			name:           "empty password",
			username:       validConfirmedUser,
			password:       "",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.NewMissingRequiredFieldError("password"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.fga)

			loginJSON := models.LoginRequest{
				Username: tc.username,
				Password: tc.password,
			}

			body, err := json.Marshal(loginJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.LoginReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				assert.True(t, out.Success)
			} else {
				assert.Contains(t, out.Error, tc.expectedErr.Error())
			}
		})
	}
}
