package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/httpsling"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestAccountRolesHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// add handler
	suite.registerTestHandler("POST", "account/roles", suite.h.AccountRolesHandler)

	testCases := []struct {
		name          string
		request       models.AccountRolesRequest
		expectedRoles []string
		errMsg        string
	}{
		{
			name: "happy path, default roles access",
			request: models.AccountRolesRequest{
				ObjectID:   suite.testUser1.OrganizationID,
				ObjectType: "organization",
			},
		},
		{
			name: "happy path, default roles access",
			request: models.AccountRolesRequest{
				ObjectIDs:  []string{suite.testUser1.OrganizationID, suite.testUser1.PersonalOrgID, suite.testUser2.OrganizationID},
				ObjectType: "organization",
			},
		},
		{
			name: "happy path, provide roles",
			request: models.AccountRolesRequest{
				ObjectID:   suite.testUser1.OrganizationID,
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
				assert.NilError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(suite.testUser1.UserCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.AccountRolesResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.errMsg != "" {
				assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
				assert.Check(t, !out.Success)
				assert.Check(t, is.DeepEqual(tc.errMsg, out.Error))

				return
			}

			assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
			assert.Check(t, out.Success)

			// if no roles are provided we expect all the roles, adding a number to assume its higher
			// than our current model has
			if len(tc.expectedRoles) == 0 && len(tc.request.ObjectIDs) == 0 {
				assert.Check(t, len(out.Roles) > 5)
			} else if len(tc.request.ObjectIDs) > 0 {
				assert.Check(t, is.Len(out.ObjectRoles, len(tc.request.ObjectIDs)))
				assert.Check(t, is.Len(out.ObjectRoles[suite.testUser2.OrganizationID], 0))
				assert.Check(t, len(out.ObjectRoles[suite.testUser1.OrganizationID]) > 5)
				assert.Check(t, len(out.ObjectRoles[suite.testUser1.PersonalOrgID]) > 5)
			} else {
				assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values(tc.expectedRoles)), slices.Sorted(slices.Values(out.Roles))))
			}
		})
	}
}
