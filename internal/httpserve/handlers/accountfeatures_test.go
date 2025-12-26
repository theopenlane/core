package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/common/models"
	apimodels "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestAccountFeaturesHandler() {
	t := suite.T()

	// add handler
	// Create operation for AccountFeaturesHandler
	operation := suite.createImpersonationOperation("AccountFeaturesHandler", "Get account features")
	suite.registerTestHandler("POST", "account/features", operation, suite.h.AccountFeaturesHandler)

	// add modules for the user
	modulesEnabled := []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule}

	featuresExpected := []string{}
	for _, m := range modulesEnabled {
		featuresExpected = append(featuresExpected, m.String())
	}

	suite.enableModules(testUser1.ID, testUser1.OrganizationID, modulesEnabled)

	testCases := []struct {
		name             string
		request          apimodels.AccountFeaturesRequest
		expectedFeatures []string
		errMsg           string
	}{
		{
			name: "happy path, feature access",
			request: apimodels.AccountFeaturesRequest{
				ID: testUser1.OrganizationID,
			},
			expectedFeatures: featuresExpected,
		},
		{
			name:             "no id provided, get from context",
			request:          apimodels.AccountFeaturesRequest{},
			expectedFeatures: featuresExpected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := "/account/features"

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

			var out *apimodels.AccountFeaturesReply

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
			assert.ElementsMatch(t, tc.expectedFeatures, out.Features)
		})
	}
}
