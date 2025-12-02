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
	models "github.com/theopenlane/shared/openapi"
)

func (suite *HandlerTestSuite) TestAccountAccessHandler() {
	t := suite.T()

	// add handler
	// Create operation for AccountAccessHandler
	operation := suite.createImpersonationOperation("AccountAccessHandler", "Check account access")
	suite.registerTestHandler("POST", "account/access", operation, suite.h.AccountAccessHandler)

	testCases := []struct {
		name    string
		request models.AccountAccessRequest
		allowed bool
		errMsg  string
	}{
		{
			name: "happy path, allow access",
			request: models.AccountAccessRequest{
				ObjectID:   testUser1.OrganizationID,
				ObjectType: "organization",
				Relation:   "can_view",
			},
			allowed: true,
		},
		{
			name: "access denied",
			request: models.AccountAccessRequest{
				ObjectID:   "another-org-id",
				ObjectType: "organization",
				Relation:   "can_delete",
			},
			allowed: false,
		},
		{
			name: "missing object id",
			request: models.AccountAccessRequest{
				ObjectType: "organization",
				Relation:   "can_delete",
			},
			errMsg: "object_id is required",
		},
		{
			name: "missing object type",
			request: models.AccountAccessRequest{
				ObjectID: "org-id",
				Relation: "can_delete",
			},
			errMsg: "object_type is required",
		},
		{
			name: "missing relation",
			request: models.AccountAccessRequest{
				ObjectID:   "org-id",
				ObjectType: "organization",
			},
			errMsg: "relation is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := "/account/access"

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

			var out *models.AccountAccessReply

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
			assert.Equal(t, tc.allowed, out.Allowed)
		})
	}
}
