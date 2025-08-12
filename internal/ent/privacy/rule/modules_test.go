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

func createExportMutation(t *testing.T) ent.Mutation {
	t.Helper()
	return generated.NewClient().Export.Create().Mutation()
}

func setupContext(t *testing.T, org string, feats []models.OrgModule) context.Context {
	t.Helper()
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	_ = cache.SetFeatures(ctx, org, feats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestHasFeature(t *testing.T) {
	ctx := setupContext(t, "org1", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, err := rule.HasFeature(ctx, "base_module")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasFeature(ctx, "nonexistent_module")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAnyFeature(t *testing.T) {
	ctx := setupContext(t, "org2", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, _, err := rule.HasAnyFeature(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAnyFeature(ctx, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAllFeatures(t *testing.T) {
	ctx := setupContext(t, "org3", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule})

	ok, _, err := rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogComplianceModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogComplianceModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogTrustCenterModule)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestDenyIfMissingAllModules(t *testing.T) {
	tests := []struct {
		title            string
		createMutationFn func() ent.Mutation
		modules          []models.OrgModule
		shouldSkip       bool
		shouldDeny       bool
		expectedError    string
	}{
		{
			title: "Export features present should skip",
			createMutationFn: func() ent.Mutation {
				return createExportMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogBaseModule},
			shouldSkip: true,
		},
		{
			title: "Export features missing should deny",
			createMutationFn: func() ent.Mutation {
				return createExportMutation(t)
			},
			modules:       []models.OrgModule{models.CatalogComplianceModule}, // missing base module
			shouldDeny:    true,
			expectedError: "features are not enabled",
		},
		{
			title: "features present should skip",
			createMutationFn: func() ent.Mutation {
				return createExportMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogBaseModule},
			shouldSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {

			ctx := setupContext(t, "test-org", tt.modules)

			err := rule.DenyIfMissingAllModules().EvalMutation(ctx, tt.createMutationFn())

			if tt.shouldSkip {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
				return
			}

			if tt.shouldDeny {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.NotContains(t, err.Error(), "skip rule")
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDenyIfMissingAllModules_BypassScenarios(t *testing.T) {
	baseCtx := setupContext(t, "test-org", []models.OrgModule{})

	featureRule := rule.DenyIfMissingAllModules()

	testMutation := createExportMutation(t)

	t.Run("bypass with privacy decision context", func(t *testing.T) {
		ctx := privacy.DecisionContext(baseCtx, privacy.Allow)
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with OrgSubscriptionContextKey", func(t *testing.T) {
		ctx := contextx.With(baseCtx, auth.OrgSubscriptionContextKey{})
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with OrganizationCreationContextKey", func(t *testing.T) {
		ctx := contextx.With(baseCtx, auth.OrganizationCreationContextKey{})
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with EmailSignUpToken", func(t *testing.T) {
		ctx := token.NewContextWithSignUpToken(baseCtx, "test@example.com")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with ResetToken", func(t *testing.T) {
		ctx := token.NewContextWithResetToken(baseCtx, "test-reset-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with VerifyToken", func(t *testing.T) {
		ctx := token.NewContextWithVerifyToken(baseCtx, "test-verify-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with JobRunnerRegistrationToken", func(t *testing.T) {
		ctx := token.NewContextWithJobRunnerRegistrationToken(baseCtx, "test-job-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})
}
