package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestAvailableAuthTypeHandler() {
	t := suite.T()

	suite.e.POST("login/methods", suite.h.AvailableAuthTypeHandler)

	ctx := rule.WithInternalContext(
		echocontext.NewTestEchoContext().Request().Context())

	// create users in the database
	validPassword := "sup3rs3cu7e!"

	validConfirmedUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: true,
	})

	testCases := []struct {
		name            string
		username        string
		expectedMethods []enums.AuthProvider
		expectedErr     error
		expectedStatus  int
		setPasskey      bool
	}{
		{
			name:           "happy path, just credentials",
			username:       validConfirmedUser.UserInfo.Email,
			expectedStatus: http.StatusOK,
			expectedMethods: []enums.AuthProvider{
				enums.AuthProviderCredentials,
			},
		},
		{
			name:           "happy path, credentials + passkeys",
			username:       validConfirmedUser.UserInfo.Email,
			expectedStatus: http.StatusOK,
			expectedMethods: []enums.AuthProvider{
				enums.AuthProviderCredentials,
				enums.AuthProviderWebauthn,
			},
			setPasskey: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authMethodsRequest := models.AvailableAuthTypeLoginRequest{
				Username: tc.username,
			}

			body, err := json.Marshal(authMethodsRequest)
			if err != nil {
				require.NoError(t, err)
			}

			if tc.setPasskey {
				_, err := suite.db.UserSetting.
					UpdateOneID(validConfirmedUser.UserInfo.Edges.Setting.ID).
					SetIsWebauthnAllowed(true).
					Save(rule.WithInternalContext(ctx))
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login/methods", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.AvailableAuthTypeReply

			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedErr != nil {
				if tc.expectedErr.Error() != " " {
					assert.Contains(t, out.Error, tc.expectedErr.Error())
					return
				}
			}

			assert.True(t, tc.expectedStatus == http.StatusOK)
			assert.Len(t, tc.expectedMethods, len(out.Methods))
		})
	}
}
