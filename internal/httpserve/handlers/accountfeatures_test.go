package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/theopenlane/core/common/models"
	apimodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/httpsling"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestAccountFeaturesHandler(t *testing.T) {
	suite.setupTest(t)

	// add handler
	suite.registerTestHandler("POST", "account/features", suite.h.AccountFeaturesHandler)

	// add modules for the user
	modulesEnabled := []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule, models.CatalogTrustCenterModule}

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
				assert.NilError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(testUser1.UserCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.AccountFeaturesResponse

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
			assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values(tc.expectedFeatures)), slices.Sorted(slices.Values(out.Features))))
		})
	}
}
