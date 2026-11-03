package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestUserInfoHandler(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("GET", "oauth/userinfo", suite.h.UserInfo)

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
				assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))

				return
			}

			assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
			assert.Assert(t, out != nil)

			assert.Check(t, is.DeepEqual(testUser1.ID, out.ID))
		})
	}
}
