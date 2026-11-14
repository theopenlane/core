package entitlements

import (
	"testing"

	"github.com/stripe/stripe-go/v87"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestEntitlementsFeatureCreateOptions(t *testing.T) {
	params := &stripe.EntitlementsFeatureCreateParams{}
	WithFeatureName("TestFeature")(params)
	assert.Check(t, is.Equal("TestFeature", *params.Name))

	WithFeatureLookupKey("test-feature")(params)
	assert.Check(t, is.Equal("test-feature", *params.LookupKey))
}

func TestProductFeatureCreateOption(t *testing.T) {
	params := &stripe.ProductFeatureCreateParams{}
	WithProductFeatureProductID("prod_123")(params)
	assert.Check(t, is.Equal("prod_123", *params.Product))

	WithProductFeatureEntitlementFeatureID("feat_456")(params)
	assert.Check(t, is.Equal("feat_456", *params.EntitlementFeature))

	params2 := CreateProductFeatureAssociationWithOptions(&stripe.ProductFeatureCreateParams{},
		WithProductFeatureProductID("prod_abc"),
		WithProductFeatureEntitlementFeatureID("feat_xyz"),
	)
	assert.Check(t, is.Equal("prod_abc", *params2.Product))
	assert.Check(t, is.Equal("feat_xyz", *params2.EntitlementFeature))
}
