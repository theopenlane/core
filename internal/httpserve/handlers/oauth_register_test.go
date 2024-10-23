package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestOauthRegister() {
	t := suite.T()

	// add login handler
	suite.e.POST("oauth/register", suite.h.OauthRegister)

	type args struct {
		name     string
		email    string
		provider enums.AuthProvider
		username string
		userID   string
		token    string
	}

	tests := []struct {
		name            string
		args            args
		writes          bool
		expectedStatus  int
		expectedErr     string
		expectedErrCode rout.ErrorCode
		wantErr         bool
	}{
		{
			name: "happy path, github",
			args: args{
				name:     "Ant Man",
				email:    "antman@theopenlane.io",
				provider: enums.AuthProviderGitHub,
				username: "scarletwitch",
				userID:   "123456",
				token:    "gh_thistokenisvalid",
			},
			expectedStatus: http.StatusOK,
			writes:         true,
		},
		{
			name: "happy path, github, same user",
			args: args{
				name:     "Ant Man",
				email:    "antman@theopenlane.io",
				provider: enums.AuthProviderGitHub,
				username: "scarletwitch",
				userID:   "123456",
				token:    "gh_thistokenisvalid",
			},
			expectedStatus: http.StatusOK,
			writes:         false, // user already created, no FGA writes this time
		},
		{
			name: "mismatch email",
			args: args{
				name:     "Ant Man",
				email:    "antman@marvel.com",
				provider: enums.AuthProviderGitHub,
				username: "scarletwitch",
				userID:   "123456",
				token:    "gh_thistokenisvalid",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedErrCode: handlers.InvalidInputErrCode,
			writes:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.writes {
				// add mocks for writes when a new user is created
				// once for the personal org, and once for the _self relation
				mock_fga.WriteAny(t, suite.fga)
			}

			registerJSON := models.OauthTokenRequest{
				Name:             tt.args.name,
				Email:            tt.args.email,
				AuthProvider:     tt.args.provider.String(),
				ExternalUserID:   tt.args.userID,
				ExternalUserName: tt.args.username,
				ClientToken:      tt.args.token,
			}

			body, err := json.Marshal(registerJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(string(body)))
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

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.Equal(t, tt.expectedErrCode, out.ErrorCode)

			if tt.expectedStatus == http.StatusOK {
				assert.NotNil(t, out.AccessToken)
				assert.NotNil(t, out.RefreshToken)
				assert.True(t, out.Success)
				assert.Equal(t, "Bearer", out.TokenType)
			}
		})
	}
}
