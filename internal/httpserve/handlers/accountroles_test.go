package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/httpsling"

	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestAccountRolesHandler() {
	t := suite.T()

	// add handler
	suite.e.POST("account/roles", suite.h.AccountRolesHandler)

	testCases := []struct {
		name          string
		request       models.AccountRolesRequest
		expectedRoles []string
		errMsg        string
	}{
		{
			name: "happy path, default roles access",
			request: models.AccountRolesRequest{
				ObjectID:   testUser1.OrganizationID,
				ObjectType: "organization",
			},
			expectedRoles: []string{"can_view", "can_edit", "can_delete", "audit_log_viewer", "can_invite_admins", "can_invite_members"},
		},
		{
			name: "happy path, provide roles",
			request: models.AccountRolesRequest{
				ObjectID:   testUser1.OrganizationID,
				ObjectType: "organization",
				Relations:  []string{"owner"},
			},
			expectedRoles: []string{"owner"},
		},
		{
			name: "missing object id",
			request: models.AccountRolesRequest{
				ObjectType: "organization",
			},
			errMsg: "objectId is required",
		},
		{
			name: "missing object type",
			request: models.AccountRolesRequest{
				ObjectID: "org-id",
			},
			errMsg: "objectType is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errMsg == "" {
				if len(tc.request.Relations) == 0 {
					tc.request.Relations = handlers.DefaultAllRelations
				}
			}

			target := "/account/roles"

			body, err := json.Marshal(tc.request)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(testUser1.UserCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.AccountRolesReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.errMsg != "" {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assert.False(t, out.Success)
				assert.Equal(t, tc.errMsg, out.Error)

				return
			}

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.True(t, out.Success)
			assert.Equal(t, tc.expectedRoles, out.Roles)
		})
	}
}
