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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestRegisterJobRunner() {
	t := suite.T()

	// add handler
	// Create operation for RegisterJobRunner
	operation := suite.createImpersonationOperation("RegisterJobRunner", "Register job runner")
	suite.registerTestHandler("POST", "/v1/runners", operation, suite.h.RegisterJobRunner)

	// Create a valid registration token
	ctx := context.Background()
	ctx = privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	token := "test-token-123"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := suite.db.JobRunnerRegistrationToken.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Exec(ctx)

	require.NoError(t, err)

	testCases := []struct {
		name           string
		wantErr        bool
		request        models.JobRunnerRegistrationRequest
		errMsg         string
		expectedStatus int
	}{
		{
			name:           "valid registration",
			wantErr:        false,
			expectedStatus: http.StatusCreated,
			request: models.JobRunnerRegistrationRequest{
				Token:     token,
				Name:      "test-runner",
				IPAddress: gofakeit.IPv4Address(),
			},
		},
		{
			name:           "invalid token",
			wantErr:        true,
			expectedStatus: http.StatusUnauthorized,
			request: models.JobRunnerRegistrationRequest{
				Token:     "invalid-token",
				Name:      "test-runner",
				IPAddress: gofakeit.IPv4Address(),
			},
			errMsg: "unauthorized",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/runners", strings.NewReader(string(reqBody)))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(ctx))

			res := recorder.Result()
			defer res.Body.Close()

			var resp *models.JobRunnerRegistrationReply

			// parse response body
			err = json.NewDecoder(res.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.wantErr {
				assert.False(t, resp.Reply.Success)
				return
			}

			assert.Contains(t, "Job runner node registered", resp.Message)
		})
	}
}

func (suite *HandlerTestSuite) TestRegisterJobRunner_ExpiredToken() {
	t := suite.T()

	// add handler
	// Create operation for RegisterJobRunner
	operation := suite.createImpersonationOperation("RegisterJobRunner", "Register job runner")
	suite.registerTestHandler("POST", "/v1/runners", operation, suite.h.RegisterJobRunner)

	// Create a valid registration token
	ctx := context.Background()
	ctx = privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	token := "test-token-12345"
	expiresAt := time.Now().Add(-24 * time.Hour)

	err := suite.db.JobRunnerRegistrationToken.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Exec(ctx)

	require.NoError(t, err)

	testCases := []struct {
		name           string
		wantErr        bool
		request        models.JobRunnerRegistrationRequest
		errMsg         string
		expectedStatus int
	}{
		{
			name:           "expired token",
			wantErr:        false,
			expectedStatus: http.StatusUnauthorized,
			request: models.JobRunnerRegistrationRequest{
				Token:     token,
				Name:      "test-runner",
				IPAddress: gofakeit.IPv4Address(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/runners", strings.NewReader(string(reqBody)))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(ctx))

			res := recorder.Result()
			defer res.Body.Close()

			var resp *models.JobRunnerRegistrationReply

			// parse response body
			err = json.NewDecoder(res.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.wantErr {
				assert.False(t, resp.Reply.Success)
				return
			}

			assert.Contains(t, "Job runner node registered", resp.Message)
		})
	}
}
