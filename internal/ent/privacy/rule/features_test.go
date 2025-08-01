package rule_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/core/pkg/testutils"
)

func setupContext(org string, feats []string) context.Context {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))
	_ = cache.SetFeatures(ctx, org, feats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestHasFeature(t *testing.T) {
	ctx := setupContext("org1", []string{"base_module", "compliance_module"})

	ok, err := rule.HasFeature(ctx, "base_module")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasFeature(ctx, "nonexistent_module")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAnyFeature(t *testing.T) {
	ctx := setupContext("org2", []string{"base_module", "compliance_module"})

	ok, err := rule.HasAnyFeature(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasAnyFeature(ctx, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAllFeatures(t *testing.T) {
	ctx := setupContext("org3", []string{"base_module", "compliance_module", "entity-management_module"})

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
	ctx := setupContext("org4", []string{"compliance_module", "trust_center_module"})

	featureRule := rule.DenyIfMissingAllFeatures("test_schema", models.CatalogComplianceModule, models.CatalogTrustCenterModule)
	err := featureRule.EvalQuery(ctx, nil)

	// When all features are present, the rule should return privacy.Skip to continue to next rule
	// In the context of testing, this appears as an error with "skip rule" message
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skip rule")

	ctx = setupContext("org5", []string{"compliance_module"}) // missing trust_center_module

	err = featureRule.EvalQuery(ctx, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "features are not enabled")
}
