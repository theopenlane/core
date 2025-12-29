package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestAccountRolesHandler() {
	t := suite.T()

	// add handler
	// Create operation for AccountRolesHandler
	operation := suite.createImpersonationOperation("AccountRolesHandler", "Get account roles")
	suite.registerTestHandler("POST", "account/roles", operation, suite.h.AccountRolesHandler)

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
		},
		{
			name: "happy path, default roles access",
			request: models.AccountRolesRequest{
				ObjectIDs:  []string{testUser1.OrganizationID, testUser1.PersonalOrgID, testUser2.OrganizationID},
				ObjectType: "organization",
			},
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
			errMsg: "object_id is required",
		},
		{
			name: "missing object type",
			request: models.AccountRolesRequest{
				ObjectID: "org-id",
			},
			errMsg: "object_type is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			// if no roles are provided we expect all the roles, adding a number to assume its higher
			// than our current model has
			if len(tc.expectedRoles) == 0 && len(tc.request.ObjectIDs) == 0 {
				assert.Greater(t, len(out.Roles), 5)
			} else if len(tc.request.ObjectIDs) > 0 {
				assert.Len(t, out.ObjectRoles, len(tc.request.ObjectIDs))
				assert.Len(t, out.ObjectRoles[testUser2.OrganizationID], 0)
				assert.Greater(t, len(out.ObjectRoles[testUser1.OrganizationID]), 5)
				assert.Greater(t, len(out.ObjectRoles[testUser1.PersonalOrgID]), 5)
			} else {
				assert.ElementsMatch(t, tc.expectedRoles, out.Roles)
			}
		})
	}
}
