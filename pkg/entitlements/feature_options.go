package entitlements

import "github.com/stripe/stripe-go/v82"

// EntitlementsFeatureCreateOption allows customizing EntitlementsFeatureCreateParams
type EntitlementsFeatureCreateOption func(params *stripe.EntitlementsFeatureCreateParams)

// WithFeatureName allows setting the name of the feature
func WithFeatureName(name string) EntitlementsFeatureCreateOption {
	return func(params *stripe.EntitlementsFeatureCreateParams) {
		params.Name = stripe.String(name)
	}
}

// WithFeatureLookupKey allows setting the lookup key for a feature
func WithFeatureLookupKey(lookupKey string) EntitlementsFeatureCreateOption {
	return func(params *stripe.EntitlementsFeatureCreateParams) {
		params.LookupKey = stripe.String(lookupKey)
	}
}

// EntitlementsFeatureCreateWithOptions creates create params with functional options
type EntitlementsFeatureUpdateOption func(params *stripe.EntitlementsFeatureUpdateParams)

// WithUpdateFeatureName allows setting the name of the feature in update params
func WithUpdateFeatureName(name string) EntitlementsFeatureUpdateOption {
	return func(params *stripe.EntitlementsFeatureUpdateParams) {
		params.Name = stripe.String(name)
	}
}

// UpdateFeatureWithOptions creates update params with functional options
func (sc *StripeClient) UpdateFeatureWithOptions(baseParams *stripe.EntitlementsFeatureUpdateParams, opts ...EntitlementsFeatureUpdateOption) *stripe.EntitlementsFeatureUpdateParams {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return params
}

// AttachFeatureToProductWithOptions creates params for attaching a feature to a product
type AttachFeatureToProductWithOptions func(params *stripe.ProductFeatureCreateParams)

// ProductFeatureCreateOption allows customizing ProductFeatureCreateParams
type ProductFeatureCreateOption func(params *stripe.ProductFeatureCreateParams)

// WithProductFeatureProductID allows setting the product ID for a product feature
func WithProductFeatureProductID(productID string) ProductFeatureCreateOption {
	return func(params *stripe.ProductFeatureCreateParams) {
		params.Product = stripe.String(productID)
	}
}

// WithProductFeatureEntitlementFeatureID allows setting the entitlement feature ID for a product feature
func WithProductFeatureEntitlementFeatureID(featureID string) ProductFeatureCreateOption {
	return func(params *stripe.ProductFeatureCreateParams) {
		params.EntitlementFeature = stripe.String(featureID)
	}
}

// CreateProductFeatureAssociationWithOptions creates params for associating a feature to a product
func CreateProductFeatureAssociationWithOptions(baseParams *stripe.ProductFeatureCreateParams, opts ...ProductFeatureCreateOption) *stripe.ProductFeatureCreateParams {
	params := baseParams
	for _, opt := range opts {
		opt(params)
	}
	return params
}
