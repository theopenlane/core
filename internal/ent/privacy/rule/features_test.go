package rule_test

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/core/pkg/testutils"
)

func setupContext(org string, feats []models.OrgModule) context.Context {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	featStrs := make([]string, len(feats))
	for i, feat := range feats {
		featStrs[i] = string(feat)
	}

	_ = cache.SetFeatures(ctx, org, featStrs)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestHasFeature(t *testing.T) {
	ctx := setupContext("org1", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, err := rule.HasFeature(ctx, "base_module")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasFeature(ctx, "nonexistent_module")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAnyFeature(t *testing.T) {
	ctx := setupContext("org2", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, err := rule.HasAnyFeature(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasAnyFeature(ctx, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAllFeatures(t *testing.T) {
	ctx := setupContext("org3", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule})

	ok, err := rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogComplianceModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasAllFeatures(ctx, models.CatalogComplianceModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasAllFeatures(ctx, models.CatalogTrustCenterModule)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestDenyIfMissingAllFeatures(t *testing.T) {
	tests := []struct {
		name          string
		features      []models.OrgModule
		enabledFeats  []models.OrgModule
		expectedSkip  bool
		expectedDeny  bool
		expectedError string
	}{
		{
			name:         "all features present should skip",
			features:     []models.OrgModule{models.CatalogComplianceModule, models.CatalogTrustCenterModule},
			enabledFeats: []models.OrgModule{models.CatalogComplianceModule, models.CatalogTrustCenterModule},
			expectedSkip: true,
		},
		{
			name:          "missing features should deny",
			features:      []models.OrgModule{models.CatalogComplianceModule, models.CatalogTrustCenterModule},
			enabledFeats:  []models.OrgModule{models.CatalogComplianceModule}, // missing trust_center_module
			expectedDeny:  true,
			expectedError: "features are not enabled",
		},
		{
			name:         "single feature present should skip",
			features:     []models.OrgModule{models.CatalogComplianceModule},
			enabledFeats: []models.OrgModule{models.CatalogComplianceModule},
			expectedSkip: true,
		},
		{
			name:          "single feature missing should deny",
			features:      []models.OrgModule{models.CatalogComplianceModule},
			enabledFeats:  []models.OrgModule{models.CatalogBaseModule},
			expectedDeny:  true,
			expectedError: "features are not enabled",
		},
		{
			name:         "empty features list should skip",
			features:     []models.OrgModule{},
			enabledFeats: []models.OrgModule{models.CatalogComplianceModule},
			expectedSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupContext("test-org", tt.enabledFeats)
			featureRule := rule.DenyIfMissingAllFeatures("test_schema", tt.features...)

			mockMutation := &generated.OrganizationMutation{}
			mockMutation.SetOp(ent.OpCreate)

			err := featureRule.EvalMutation(ctx, mockMutation)

			if tt.expectedSkip {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
				return
			}

			if tt.expectedDeny {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.NotContains(t, err.Error(), "skip rule")
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDenyIfMissingAllFeatures_BypassScenarios(t *testing.T) {
	baseCtx := setupContext("test-org", []models.OrgModule{})
	featureRule := rule.DenyIfMissingAllFeatures("test_schema", models.CatalogComplianceModule)

	mockMutation := &generated.OrganizationMutation{}
	mockMutation.SetOp(ent.OpCreate)

	t.Run("bypass with privacy decision context", func(t *testing.T) {
		ctx := privacy.DecisionContext(baseCtx, privacy.Allow)
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with OrgSubscriptionContextKey", func(t *testing.T) {
		ctx := contextx.With(baseCtx, auth.OrgSubscriptionContextKey{})
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with OrganizationCreationContextKey", func(t *testing.T) {
		ctx := contextx.With(baseCtx, auth.OrganizationCreationContextKey{})
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with EmailSignUpToken", func(t *testing.T) {
		ctx := token.NewContextWithSignUpToken(baseCtx, "test@example.com")
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with ResetToken", func(t *testing.T) {
		ctx := token.NewContextWithResetToken(baseCtx, "test-reset-token")
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with VerifyToken", func(t *testing.T) {
		ctx := token.NewContextWithVerifyToken(baseCtx, "test-verify-token")
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with JobRunnerRegistrationToken", func(t *testing.T) {
		ctx := token.NewContextWithJobRunnerRegistrationToken(baseCtx, "test-job-token")
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})
}

func TestDenyQueryIfMissingAllFeatures(t *testing.T) {
	tests := []struct {
		name          string
		features      []models.OrgModule
		enabledFeats  []models.OrgModule
		expectedSkip  bool
		expectedDeny  bool
		expectedError string
	}{
		{
			name:         "all features present should skip",
			features:     []models.OrgModule{models.CatalogComplianceModule, models.CatalogTrustCenterModule},
			enabledFeats: []models.OrgModule{models.CatalogComplianceModule, models.CatalogTrustCenterModule},
			expectedSkip: true,
		},
		{
			name:          "missing features should deny",
			features:      []models.OrgModule{models.CatalogComplianceModule, models.CatalogTrustCenterModule},
			enabledFeats:  []models.OrgModule{models.CatalogComplianceModule}, // missing trust_center_module
			expectedDeny:  true,
			expectedError: "features are not enabled",
		},
		{
			name:         "single feature present should skip",
			features:     []models.OrgModule{models.CatalogComplianceModule},
			enabledFeats: []models.OrgModule{models.CatalogComplianceModule},
			expectedSkip: true,
		},
		{
			name:          "single feature missing should deny",
			features:      []models.OrgModule{models.CatalogComplianceModule},
			enabledFeats:  []models.OrgModule{models.CatalogBaseModule},
			expectedDeny:  true,
			expectedError: "features are not enabled",
		},
		{
			name:         "empty features list should skip",
			features:     []models.OrgModule{},
			enabledFeats: []models.OrgModule{models.CatalogComplianceModule},
			expectedSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupContext("test-org", tt.enabledFeats)
			featureRule := rule.DenyQueryIfMissingAllFeatures("test_schema", tt.features...)

			mockQuery := &generated.OrganizationQuery{}

			err := featureRule.EvalQuery(ctx, mockQuery)

			if tt.expectedSkip {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
				return
			}

			if tt.expectedDeny {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.NotContains(t, err.Error(), "skip rule")
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDenyIfMissingAllFeatures_EdgeCases(t *testing.T) {
	mockMutation := &generated.OrganizationMutation{}
	mockMutation.SetOp(ent.OpCreate)

	t.Run("no authenticated user should skip", func(t *testing.T) {
		ctx := context.Background()
		r := testutils.NewRedisClient()
		cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))
		ctx = permissioncache.WithCache(ctx, cache)

		// no authenticated user in context

		featureRule := rule.DenyIfMissingAllFeatures("test_schema", models.CatalogComplianceModule)
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("empty organization ID should skip", func(t *testing.T) {
		ctx := context.Background()
		r := testutils.NewRedisClient()
		cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))
		ctx = permissioncache.WithCache(ctx, cache)
		ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: ""})

		featureRule := rule.DenyIfMissingAllFeatures("test_schema", models.CatalogComplianceModule)
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("no features to check should skip", func(t *testing.T) {
		ctx := setupContext("test-org", []models.OrgModule{})
		featureRule := rule.DenyIfMissingAllFeatures("test_schema")
		err := featureRule.EvalMutation(ctx, mockMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})
}
