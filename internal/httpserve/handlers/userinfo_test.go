package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ent "github.com/theopenlane/ent/generated"
)

func (suite *HandlerTestSuite) TestUserInfoHandler() {
	t := suite.T()

	// Create operation for UserInfo
	operation := suite.createImpersonationOperation("UserInfo", "Get user information")
	suite.registerTestHandler("GET", "oauth/userinfo", operation, suite.h.UserInfo)

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "happy path",
			ctx:     testUser1.UserCtx,
			wantErr: false,
		},
		{
			name:    "empty context",
			ctx:     context.Background(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest(http.MethodGet, "/oauth/userinfo", nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(tt.ctx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *ent.User

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				return
			}

			assert.Equal(t, http.StatusOK, recorder.Code)
			require.NotNil(t, out)

			assert.Equal(t, testUser1.ID, out.ID)
		})
	}
}
