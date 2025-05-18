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
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestRegisterJobRunner() {
	t := suite.T()

	// add handler
	suite.e.POST("/v1/runners", suite.h.RegisterJobRunner)

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
		name    string
		wantErr bool
		request models.JobRunnerRegistrationRequest
		errMsg  string
	}{
		{
			name:    "valid registration",
			wantErr: false,
			request: models.JobRunnerRegistrationRequest{
				Token:     token,
				Name:      "test-runner",
				IPAddress: gofakeit.IPv4Address(),
			},
		},
		// {
		// 	name:    "invalid token",
		// 	wantErr: true,
		// 	request: models.JobRunnerRegistrationRequest{
		// 		Token:     "invalid-token",
		// 		Name:      "test-runner",
		// 		IPAddress: gofakeit.IPv4Address(),
		// 	},
		// 	errMsg: "unauthorized",
		// },
		// {
		// 	name:    "empty token",
		// 	wantErr: true,
		// 	request: models.JobRunnerRegistrationRequest{
		// 		Token:     "",
		// 		Name:      "test-runner",
		// 		IPAddress: gofakeit.IPv4Address(),
		// 	},
		// 	errMsg: "invalid input",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

			// Create request body
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/runners", strings.NewReader(string(reqBody)))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(ctx))

			res := recorder.Result()
			defer res.Body.Close()

			var out models.JobRunnerRegistrationResponse

			// parse response body
			err = json.NewDecoder(res.Body).Decode(&out)
			require.NoError(t, err)

			if tc.wantErr {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assert.Contains(t, out.Reply.Error, tc.errMsg)
				return
			}

			assert.Equal(t, http.StatusCreated, recorder.Code)
			assert.True(t, out.Reply.Success)
			assert.Equal(t, "Job runner node registered", out.Message)
		})
	}
}
