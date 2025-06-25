package entitlements

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

func TestEntitlementsFeatureCreateOptions(t *testing.T) {
	params := &stripe.EntitlementsFeatureCreateParams{}
	WithFeatureName("TestFeature")(params)
	require.Equal(t, "TestFeature", *params.Name)

	WithFeatureLookupKey("test-feature")(params)
	require.Equal(t, "test-feature", *params.LookupKey)
}

func TestProductFeatureCreateOption(t *testing.T) {
	params := &stripe.ProductFeatureCreateParams{}
	WithProductFeatureProductID("prod_123")(params)
	require.Equal(t, "prod_123", *params.Product)

	WithProductFeatureEntitlementFeatureID("feat_456")(params)
	require.Equal(t, "feat_456", *params.EntitlementFeature)

	params2 := CreateProductFeatureAssociationWithOptions(&stripe.ProductFeatureCreateParams{},
		WithProductFeatureProductID("prod_abc"),
		WithProductFeatureEntitlementFeatureID("feat_xyz"),
	)
	require.Equal(t, "prod_abc", *params2.Product)
	require.Equal(t, "feat_xyz", *params2.EntitlementFeature)
}
