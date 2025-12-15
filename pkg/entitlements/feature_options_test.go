package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v84"
)

func TestEntitlementsFeatureCreateOptions(t *testing.T) {
	params := &stripe.EntitlementsFeatureCreateParams{}
	WithFeatureName("TestFeature")(params)
	assert.Equal(t, "TestFeature", *params.Name)

	WithFeatureLookupKey("test-feature")(params)
	assert.Equal(t, "test-feature", *params.LookupKey)
}

func TestProductFeatureCreateOption(t *testing.T) {
	params := &stripe.ProductFeatureCreateParams{}
	WithProductFeatureProductID("prod_123")(params)
	assert.Equal(t, "prod_123", *params.Product)

	WithProductFeatureEntitlementFeatureID("feat_456")(params)
	assert.Equal(t, "feat_456", *params.EntitlementFeature)

	params2 := CreateProductFeatureAssociationWithOptions(&stripe.ProductFeatureCreateParams{},
		WithProductFeatureProductID("prod_abc"),
		WithProductFeatureEntitlementFeatureID("feat_xyz"),
	)
	assert.Equal(t, "prod_abc", *params2.Product)
	assert.Equal(t, "feat_xyz", *params2.EntitlementFeature)
}
