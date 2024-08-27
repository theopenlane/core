package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestAccountRolesOrganizationHandler() {
	t := suite.T()

	// add handler
	suite.e.GET("account/roles/organization", suite.h.AccountRolesOrganizationHandler)
	suite.e.GET("account/roles/organization/:id", suite.h.AccountRolesOrganizationHandler)

	// bypass auth
	ctx := context.Background()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	mock_fga.WriteAny(t, suite.fga)

	// setup test data
	requestor := suite.db.User.Create().
		SetEmail("mp@theopenlane.io").
		SetFirstName("Mikey").
		SetLastName("Polo").
		SaveX(ctx)

	reqCtx, err := userContextWithID(requestor.ID)
	require.NoError(t, err)

	mock_fga.ClearMocks(suite.fga)

	testCases := []struct {
		name      string
		id        string
		target    string
		mockRoles []string
		errMsg    string
	}{
		{
			name:      "happy path, no id provided",
			target:    "/account/roles/organization",
			mockRoles: []string{"can_view"},
		},
		{
			name:      "happy path, id provided",
			target:    "/account/roles/organization/ulid_id_of_org",
			mockRoles: []string{"can_view"},
		},
		{
			name:   "org not authorized",
			target: "/account/roles/organization/another_ulid_id_of_org",
			errMsg: "invalid input",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.fga)

			if tc.errMsg == "" {
				mock_fga.BatchCheck(t, suite.fga, tc.mockRoles, handlers.DefaultAllRelations)
			}

			req := httptest.NewRequest(http.MethodGet, tc.target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

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
			assert.Equal(t, tc.mockRoles, out.Roles)
			assert.Equal(t, "ulid_id_of_org", out.OrganizationID)
		})
	}
}
