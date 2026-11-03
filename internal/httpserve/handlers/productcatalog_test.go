package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apimodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/httpsling"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestProductCatalogHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// add handler
	suite.registerTestHandler(http.MethodGet, "products", suite.h.ProductCatalogHandler)

	testCases := []struct {
		name                  string
		request               apimodels.ProductCatalogRequest
		expectBetaFeatures    bool
		expectPrivateFeatures bool
	}{
		{
			name:                  "happy path, only public products",
			request:               apimodels.ProductCatalogRequest{},
			expectBetaFeatures:    false,
			expectPrivateFeatures: false,
		},
		{
			name:                  "happy path, only public products",
			request:               apimodels.ProductCatalogRequest{},
			expectBetaFeatures:    true,
			expectPrivateFeatures: false,
		},
		{
			name:                  "happy path, all products",
			request:               apimodels.ProductCatalogRequest{},
			expectBetaFeatures:    true,
			expectPrivateFeatures: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := "/products"

			// setup query params
			if tc.expectBetaFeatures || tc.expectPrivateFeatures {
				if tc.expectBetaFeatures {
					target += "?include_beta=true"
				}

				if tc.expectPrivateFeatures {
					if tc.expectBetaFeatures {
						target += "&"
					} else {
						target += "?"
					}

					target += "include_private=true"
				}
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(context.Background()))

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.ProductCatalogResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
			assert.Check(t, out.Success)

			// we should always have public products in the catalog
			assert.Check(t, 0 != len(out.Catalog.Modules))

			assert.Check(t, len(out.Version) > 0)
			assert.Check(t, len(out.SHA) > 0)

			for _, v := range out.Catalog.Modules {
				switch v.Audience {
				case "public":
					// always included
				case "beta":
					assert.Check(t, tc.expectBetaFeatures, "unexpected beta feature found: %v", v)
				case "private":
					assert.Check(t, tc.expectPrivateFeatures, "unexpected private feature found: %v", v)
				default:
					// ensure we don't have new audiences we aren't handling
					t.Errorf("unknown audience: %v", v.Audience)
				}
			}

			for _, v := range out.Catalog.Addons {
				switch v.Audience {
				case "public":
					// always included
				case "beta":
					assert.Check(t, tc.expectBetaFeatures, "unexpected beta feature found: %v", v)
				case "private":
					assert.Check(t, tc.expectPrivateFeatures, "unexpected private feature found: %v", v)
				default:
					// ensure we don't have new audiences we aren't handling
					t.Errorf("unknown audience: %v", v.Audience)
				}
			}
		})
	}
}
