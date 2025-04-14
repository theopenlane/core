package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestAccountRolesOrganizationHandler() {
	t := suite.T()

	// add handler
	suite.e.GET("account/roles/organization", suite.h.AccountRolesOrganizationHandler)
	suite.e.GET("account/roles/organization/:id", suite.h.AccountRolesOrganizationHandler)

	testCases := []struct {
		name   string
		id     string
		target string
		errMsg string
	}{
		{
			name:   "happy path, no id provided",
			target: "/account/roles/organization",
		},
		{
			name:   "happy path, id provided",
			target: "/account/roles/organization/" + testUser1.OrganizationID,
		},
		{
			name:   "org not authorized",
			target: "/account/roles/organization/another_ulid_id_of_org",
			errMsg: "invalid input",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(testUser1.UserCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.AccountRolesOrganizationReply

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
			assert.ElementsMatch(t, []string{"can_view", "can_edit", "can_delete", "audit_log_viewer", "can_invite_admins", "can_invite_members"}, out.Roles)
			assert.Equal(t, testUser1.OrganizationID, out.OrganizationID)
		})
	}
}
