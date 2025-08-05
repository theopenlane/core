package interceptors_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/core/pkg/testutils"
)

type mockQuery struct {
	queryType string
}

func (m *mockQuery) WhereP(...func(*sql.Selector)) {}
func (m *mockQuery) Where(...func(*sql.Selector))  {}
func (m *mockQuery) Limit(int)                     {}
func (m *mockQuery) Offset(int)                    {}
func (m *mockQuery) Unique(bool)                   {}
func (m *mockQuery) Order(...func(*sql.Selector))  {}
func (m *mockQuery) Select(...string)              {}

type mockQuerier struct {
	result ent.Value
	err    error
}

func (m *mockQuerier) Query(ctx context.Context, query ent.Query) (ent.Value, error) {
	return m.result, m.err
}

func ctxWithFeatures(org string, feats []string) context.Context {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	_ = cache.SetFeatures(ctx, org, feats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestInterceptorFeatures_AllFeaturesEnabled(t *testing.T) {
	ctx := ctxWithFeatures("test-org", []string{
		string(models.CatalogBaseModule),
		string(models.CatalogComplianceModule),
		string(models.CatalogEntityManagementModule),
	})

	interceptor := interceptors.InterceptorFeatures(
		models.CatalogBaseModule,
		models.CatalogComplianceModule,
		models.CatalogEntityManagementModule,
	)

	expectedResult := "query_result"
	mockNext := &mockQuerier{result: expectedResult, err: nil}

	querier := interceptor.Intercept(mockNext)

	result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestInterceptorFeatures_MissingFeatures(t *testing.T) {
	ctx := ctxWithFeatures("test-org", []string{string(models.CatalogBaseModule)})

	interceptor := interceptors.InterceptorFeatures(
		models.CatalogBaseModule,
		models.CatalogComplianceModule,
		models.CatalogEntityManagementModule,
	)

	expectedResult := "query_result"
	mockNext := &mockQuerier{result: expectedResult, err: nil}

	querier := interceptor.Intercept(mockNext)

	result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestInterceptorFeatures_NoFeaturesRequired(t *testing.T) {
	ctx := ctxWithFeatures("test-org", []string{})

	interceptor := interceptors.InterceptorFeatures()

	expectedResult := "query_result"
	mockNext := &mockQuerier{result: expectedResult, err: nil}

	querier := interceptor.Intercept(mockNext)

	result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestInterceptorFeatures_SingleFeature(t *testing.T) {
	testCases := []struct {
		name            string
		enabledFeatures []string
		requiredFeature models.OrgModule
		expectSuccess   bool
	}{
		{
			name:            "single feature enabled",
			enabledFeatures: []string{string(models.CatalogBaseModule)},
			requiredFeature: models.CatalogBaseModule,
			expectSuccess:   true,
		},
		{
			name:            "single feature not enabled",
			enabledFeatures: []string{string(models.CatalogComplianceModule)},
			requiredFeature: models.CatalogBaseModule,
			expectSuccess:   false,
		},
		{
			name:            "trust center feature enabled",
			enabledFeatures: []string{string(models.CatalogTrustCenterModule)},
			requiredFeature: models.CatalogTrustCenterModule,
			expectSuccess:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ctxWithFeatures("test-org", tc.enabledFeatures)

			interceptor := interceptors.InterceptorFeatures(tc.requiredFeature)

			expectedResult := "query_result"
			mockNext := &mockQuerier{result: expectedResult, err: nil}

			querier := interceptor.Intercept(mockNext)

			result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

			require.NoError(t, err)
			if tc.expectSuccess {
				assert.Equal(t, expectedResult, result)
				return
			}

			assert.Nil(t, result)
		})
	}
}

func TestInterceptorFeatures_QueryError(t *testing.T) {
	ctx := ctxWithFeatures("test-org", []string{string(models.CatalogBaseModule)})

	interceptor := interceptors.InterceptorFeatures(models.CatalogBaseModule)

	expectedError := errors.New("query execution error")
	mockNext := &mockQuerier{result: nil, err: expectedError}

	querier := interceptor.Intercept(mockNext)

	result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestInterceptorFeatures_NoAuthenticatedUser(t *testing.T) {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))
	ctx = permissioncache.WithCache(ctx, cache)

	interceptor := interceptors.InterceptorFeatures(models.CatalogBaseModule)

	expectedResult := "query_result"
	mockNext := &mockQuerier{result: expectedResult, err: nil}

	querier := interceptor.Intercept(mockNext)

	result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestInterceptorFeatures_AllAvailableModules(t *testing.T) {
	allModuleNames := []string{
		string(models.CatalogBaseModule),
		string(models.CatalogComplianceModule),
		string(models.CatalogDomainScanningAddon),
		string(models.CatalogEntityManagementModule),
		string(models.CatalogExtraEvidenceStorageAddon),
		string(models.CatalogPolicyManagementAddon),
		string(models.CatalogRiskManagementAddon),
		string(models.CatalogTrustCenterModule),
		string(models.CatalogVulnerabilityManagementModule),
	}

	ctx := ctxWithFeatures("test-org", allModuleNames)

	interceptor := interceptors.InterceptorFeatures(models.AllOrgModules...)

	expectedResult := "query_result"
	mockNext := &mockQuerier{result: expectedResult, err: nil}

	querier := interceptor.Intercept(mockNext)

	result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestInterceptorFeatures_MissingFeaturesBlocked(t *testing.T) {
	t.Run("no_features_enabled_blocks", func(t *testing.T) {
		ctx := ctxWithFeatures("test-org", []string{})

		interceptor := interceptors.InterceptorFeatures(models.CatalogBaseModule)
		expectedResult := "query_proceeds"
		mockNext := &mockQuerier{result: expectedResult, err: nil}
		querier := interceptor.Intercept(mockNext)

		result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("features_enabled_allows", func(t *testing.T) {
		ctx := ctxWithFeatures("test-org", []string{string(models.CatalogBaseModule)})

		interceptor := interceptors.InterceptorFeatures(models.CatalogBaseModule)
		expectedResult := "query_proceeds"
		mockNext := &mockQuerier{result: expectedResult, err: nil}
		querier := interceptor.Intercept(mockNext)

		result, err := querier.Query(ctx, &mockQuery{queryType: "test"})

		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})
}
