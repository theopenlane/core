package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
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

	// create users in the database
	validPassword := "sup3rs3cu7e!"

	validConfirmedUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: true,
	})

	validConfirmedUserRestrictedOrg := suite.userBuilderWithInput(ctx, &userInput{
		email:         "meow@example.com",
		password:      validPassword,
		confirmedUser: true,
	})

	invalidConfirmedUserRestrictedOrg := suite.userBuilderWithInput(ctx, &userInput{
		email:         "meow@foobar.com",
		password:      validPassword,
		confirmedUser: true,
	})

	validUnconfirmedUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: false,
	})

	orgSetting := suite.db.OrganizationSetting.Create().SetInput(
		generated.CreateOrganizationSettingInput{
			AllowedEmailDomains: []string{"example.com"},
		},
	).SaveX(ctx)

	input := generated.CreateOrganizationInput{
		Name:      "restricted",
		SettingID: &orgSetting.ID,
	}

	// setup allow context with the client in the context which is required for hooks that run
	allowCtx := privacy.DecisionContext(validConfirmedUserRestrictedOrg.UserCtx, privacy.Allow)
	allowCtx = ent.NewContext(allowCtx, suite.db)

	org := suite.db.Organization.Create().SetInput(input).SaveX(allowCtx)

	// update the user settings to have the default org set that is the domain restricted org
	suite.db.UserSetting.UpdateOneID(validConfirmedUserRestrictedOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(org.ID).SaveX(allowCtx)

	suite.db.UserSetting.UpdateOneID(invalidConfirmedUserRestrictedOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(org.ID).SaveX(allowCtx)

	testCases := []struct {
		name           string
		username       string
		password       string
		expectedOrgID  string
		expectedErr    error
		expectedStatus int
	}{
		{
			name:           "happy path, valid credentials",
			username:       validConfirmedUser.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  validConfirmedUser.OrganizationID,
		},
		{
			name:           "happy path, domain restricted org",
			username:       validConfirmedUserRestrictedOrg.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  org.ID,
		},
		{
			name:           "domain restricted org, email not allowed, switch to personal org",
			username:       invalidConfirmedUserRestrictedOrg.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  invalidConfirmedUserRestrictedOrg.PersonalOrgID,
		},
		{
			name:           "email unverified",
			username:       validUnconfirmedUser.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    auth.ErrUnverifiedUser,
		},
		{
			name:           "invalid password",
			username:       validConfirmedUser.UserInfo.Email,
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
			username:       validConfirmedUser.UserInfo.Email,
			password:       "",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.NewMissingRequiredFieldError("password"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
				assert.True(t, out.TFAEnabled) // we set the user to have TFA enabled in the tests
				require.NotNil(t, out.AccessToken)

				// check the claims to ensure the user is in the correct org
				token, _, err := new(jwt.Parser).ParseUnverified(out.AccessToken, jwt.MapClaims{})
				require.NoError(t, err)

				claims, ok := token.Claims.(jwt.MapClaims)
				require.True(t, ok)

				assert.Equal(t, tc.expectedOrgID, claims["org"])
			} else {
				assert.Contains(t, out.Error, tc.expectedErr.Error())
			}
		})
	}
}
