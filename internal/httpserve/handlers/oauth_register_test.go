package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func (suite *HandlerTestSuite) TestOauthRegister() {
	t := suite.T()

	// add login handler
	// Create operation for OauthRegister
	operation := suite.createImpersonationOperation("OauthRegister", "OAuth register")
	suite.registerTestHandler("POST", "oauth/register", operation, suite.h.OauthRegister)

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
		expectedStatus  int
		expectedErr     string
		expectedErrCode rout.ErrorCode
		wantErr         bool
		wantEmailJob    bool
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
			wantEmailJob:   true,
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
			wantEmailJob:   false, // should not send welcome email again
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
			wantEmailJob:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite.ClearTestData()

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
				assert.False(t, out.TFAEnabled) // we did not setup the user to have TFA
				assert.Equal(t, "Bearer", out.TokenType)

				if tt.wantEmailJob {
					job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
						[]rivertest.ExpectedJob{
							{
								Args: jobs.EmailArgs{
									Message: *newman.NewEmailMessageWithOptions(
										newman.WithSubject("Welcome to Meow Inc.!"),
										newman.WithTo([]string{tt.args.email}),
									),
								},
							},
						})
					require.NotNil(t, job)
				} else {
					rivertest.RequireNotInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
				}
			}
		})
	}
}
