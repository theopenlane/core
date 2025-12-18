package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v84"
)

// GetFeatureByLookupKey retrieves the first entitlements feature matching the lookup key.
func (sc *StripeClient) GetFeatureByLookupKey(ctx context.Context, lookupKey string) (*stripe.EntitlementsFeature, error) {
	params := &stripe.EntitlementsFeatureListParams{
		LookupKey: stripe.String(lookupKey),
	}

	params.Limit = stripe.Int64(1)
	it := sc.Client.V1EntitlementsFeatures.List(ctx, params)

	for feature, err := range it {
		if err != nil {
			return nil, err
		}

		return feature, nil
	}

	return nil, nil
}
