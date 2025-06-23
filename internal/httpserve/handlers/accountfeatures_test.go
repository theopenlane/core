package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/features"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/fgax"
)

func (suite *HandlerTestSuite) TestAccountFeaturesHandler() {
	t := suite.T()

	// Use a real StripeClient (empty struct is fine if not used in test logic)
	suite.h.Entitlements = &entitlements.StripeClient{}
	// Use a real features.Cache with a local/test Redis instance
	suite.h.FeatureCache = features.NewCache(redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Make sure Redis is running for tests
	}), 1*time.Minute)

	// Pre-populate the cache for the org
	_ = suite.h.FeatureCache.Set(context.Background(), testUser1.OrganizationID, dummyFeatures)

	// --- Ensure FGA is seeded for the test org ---
	ctx := context.Background()
	if suite.fga != nil {
		for _, feature := range dummyFeatures {
			req := fgax.TupleRequest{
				SubjectID:   testUser1.OrganizationID,
				SubjectType: "organization",
				ObjectID:    feature,
				ObjectType:  "feature",
				Relation:    "enabled",
			}
			_, err := suite.fga.WriteTupleKeys(ctx, []fgax.TupleKey{fgax.GetTupleKey(req)}, nil)
			require.NoError(t, err, "failed to seed FGA with feature: %s", feature)
		}
	}

	suite.e.POST("account/features", suite.h.AccountFeaturesHandler)

	testCases := []struct {
		name             string
		request          models.AccountFeaturesRequest
		expectedFeatures []string
		errMsg           string
		setup            func()
	}{
		{
			name: "happy path, feature access",
			request: models.AccountFeaturesRequest{
				ID: testUser1.OrganizationID,
			},
			expectedFeatures: dummyFeatures,
			setup: func() {
				suite.h.Entitlements = &entitlements.StripeClient{}
				suite.h.FeatureCache = features.NewCache(suite.h.RedisClient, 1*time.Minute)
				_ = suite.h.FeatureCache.Set(context.Background(), testUser1.OrganizationID, dummyFeatures)
			},
		},
		{
			name:             "no id provided, get from context",
			request:          models.AccountFeaturesRequest{},
			expectedFeatures: dummyFeatures,
			setup: func() {
				suite.h.Entitlements = &entitlements.StripeClient{}
				suite.h.FeatureCache = features.NewCache(suite.h.RedisClient, 1*time.Minute)
				_ = suite.h.FeatureCache.Set(context.Background(), testUser1.OrganizationID, dummyFeatures)
			},
		},
		{
			name:    "entitlements not configured",
			request: models.AccountFeaturesRequest{},
			errMsg:  "entitlements not configured",
			setup: func() {
				suite.h.Entitlements = nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			target := "/account/features"

			body, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			recorder := httptest.NewRecorder()
			suite.e.ServeHTTP(recorder, req.WithContext(testUser1.UserCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.AccountFeaturesReply
			err = json.NewDecoder(res.Body).Decode(&out)
			require.NoError(t, err)

			if tc.errMsg != "" {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assert.False(t, out.Success)
				assert.Contains(t, out.Error, tc.errMsg)
				return
			}

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.True(t, out.Success)
			assert.ElementsMatch(t, tc.expectedFeatures, out.Features)
		})
	}
}
