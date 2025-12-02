package entitlements

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v83"
)

// CreateProductFeatureWithOptions creates a product feature using functional options
func (sc *StripeClient) CreateProductFeatureWithOptions(ctx context.Context, baseParams *stripe.EntitlementsFeatureCreateParams, opts ...FeatureCreateOption) (*stripe.EntitlementsFeature, error) {
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
			log.Error().Err(err).Msg("failed to list product features")
			continue
		}

		features = append(features, feature)
	}

	return features, nil
}

// UpdateProductFeatureWithOptions updates a product feature using functional options
func (sc *StripeClient) UpdateProductFeatureWithOptions(ctx context.Context, featureID string, baseParams *stripe.EntitlementsFeatureUpdateParams, opts ...FeatureUpdateOption) (*stripe.EntitlementsFeature, error) {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return sc.Client.V1EntitlementsFeatures.Update(ctx, featureID, params)
}

// AttachFeatureToProductWithOptions attaches an existing entitlements feature to a product using functional options.
func (sc *StripeClient) AttachFeatureToProductWithOptions(ctx context.Context, baseParams *stripe.ProductFeatureCreateParams, opts ...ProductFeatureCreateOption) (*stripe.ProductFeature, error) {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return sc.Client.V1ProductFeatures.Create(ctx, params)
}
