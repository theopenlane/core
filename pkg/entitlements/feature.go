package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

// CreateProductFeatureWithOptions creates a product feature using functional options
func (sc *StripeClient) CreateProductFeatureWithOptions(ctx context.Context, baseParams *stripe.EntitlementsFeatureCreateParams, opts ...EntitlementsFeatureCreateOption) (*stripe.EntitlementsFeature, error) {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return sc.Client.V1EntitlementsFeatures.Create(ctx, params)
}

// ListProductFeatures lists all product features for a product
func (sc *StripeClient) ListProductFeatures(ctx context.Context, productID string) ([]*stripe.ProductFeature, error) {
	params := &stripe.ProductFeatureListParams{
		Product: stripe.String(productID),
	}

	result := sc.Client.V1ProductFeatures.List(ctx, params)

	var features []*stripe.ProductFeature

	for feature, err := range result {
		if err != nil {
			continue
		}

		features = append(features, feature)
	}

	return features, nil
}

// UpdateProductFeatureWithOptions updates a product feature using functional options
func (sc *StripeClient) UpdateProductFeatureWithOptions(ctx context.Context, featureID string, baseParams *stripe.EntitlementsFeatureUpdateParams, opts ...EntitlementsFeatureUpdateOption) (*stripe.EntitlementsFeature, error) {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return sc.Client.V1EntitlementsFeatures.Update(ctx, featureID, params)
}

// TODO: Implement AttachFeatureToProductWithOptions when/if the relevant types are available.

//import (
//	"github.com/stripe/stripe-go/v82"
//)

//func (sc *StripeClient) CreateProductFeature(params *stripe.EntitlementsFeatureParams) (*stripe.EntitlementsFeature, error) {
//	params := &stripe.EntitlementsFeatureParams{
//		Name:      stripe.String("My super awesome feature"),
//		LookupKey: stripe.String("my-super-awesome-feature"),
//	}
//	result, err := sc.CreateProductFeature(params)
//
//	return nil, nil
//}

// https://docs.stripe.com/api/product-feature/attach
