package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	apimodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestProductCatalogHandler() {
	t := suite.T()

	// add handler
	// Create operation for ProductCatalogHandler
	operation := suite.createImpersonationOperation("ProductCatalogHandler", "Get product catalog")
	suite.registerTestHandler(http.MethodGet, "products", operation, suite.h.ProductCatalogHandler)

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

			var out *apimodels.ProductCatalogReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.True(t, out.Success)
			assert.NotNil(t, out.Catalog)

			// we should always have public products in the catalog
			assert.NotEqual(t, 0, len(out.Catalog.Modules))

			assert.NotEmpty(t, out.Version)
			assert.NotEmpty(t, out.SHA)

			for _, v := range out.Catalog.Modules {
				switch v.Audience {
				case handlers.PublicAudience:
					// always included
				case handlers.BetaAudience:
					assert.True(t, tc.expectBetaFeatures, "unexpected beta feature found: %v", v)
				case handlers.PrivateAudience:
					assert.True(t, tc.expectPrivateFeatures, "unexpected private feature found: %v", v)
				default:
					// ensure we don't have new audiences we aren't handling
					t.Errorf("unknown audience: %v", v.Audience)
				}
			}

			for _, v := range out.Catalog.Addons {
				switch v.Audience {
				case handlers.PublicAudience:
					// always included
				case handlers.BetaAudience:
					assert.True(t, tc.expectBetaFeatures, "unexpected beta feature found: %v", v)
				case handlers.PrivateAudience:
					assert.True(t, tc.expectPrivateFeatures, "unexpected private feature found: %v", v)
				default:
					// ensure we don't have new audiences we aren't handling
					t.Errorf("unknown audience: %v", v.Audience)
				}
			}
		})
	}
}
