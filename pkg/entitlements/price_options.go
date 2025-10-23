package entitlements

import "github.com/stripe/stripe-go/v83"

// PriceCreateOption allows customizing PriceCreateParams
type PriceCreateOption func(params *stripe.PriceCreateParams)

// WithPriceProduct sets the product ID for the price
func WithPriceProduct(productID string) PriceCreateOption {
	return func(params *stripe.PriceCreateParams) {
		params.Product = stripe.String(productID)
	}
}

// WithPriceAmount sets the unit amount for the price
func WithPriceAmount(amount int64) PriceCreateOption {
	return func(params *stripe.PriceCreateParams) {
		params.UnitAmount = stripe.Int64(amount)
	}
}

// WithPriceCurrency sets the currency for the price
func WithPriceCurrency(currency string) PriceCreateOption {
	return func(params *stripe.PriceCreateParams) {
		params.Currency = stripe.String(currency)
	}
}

// WithPriceRecurring sets the recurring interval for the price
func WithPriceRecurring(interval string) PriceCreateOption {
	return func(params *stripe.PriceCreateParams) {
		params.Recurring = &stripe.PriceCreateRecurringParams{Interval: stripe.String(interval)}
	}
}

// WithPriceMetadata sets metadata for the price
func WithPriceMetadata(metadata map[string]string) PriceCreateOption {
	return func(params *stripe.PriceCreateParams) {
		params.Metadata = metadata
	}
}

// CreatePriceWithOptions creates a price with functional options
func (sc *StripeClient) CreatePriceWithOptions(baseParams *stripe.PriceCreateParams, opts ...PriceCreateOption) *stripe.PriceCreateParams {
	params := baseParams
	for _, opt := range opts {
		opt(params)
	}

	return params
}

// PriceUpdateOption allows customizing PriceUpdateParams
type PriceUpdateOption func(params *stripe.PriceUpdateParams)

// WithUpdatePriceMetadata sets metadata for the price update
func WithUpdatePriceMetadata(metadata map[string]string) PriceUpdateOption {
	return func(params *stripe.PriceUpdateParams) {
		params.Metadata = metadata
	}
}

// UpdatePriceWithOptions creates update params with functional options
func (sc *StripeClient) UpdatePriceWithOptions(baseParams *stripe.PriceUpdateParams, opts ...PriceUpdateOption) *stripe.PriceUpdateParams {
	params := baseParams
	for _, opt := range opts {
		opt(params)
	}

	return params
}

// --- Example Usage ---
// params := &stripe.PriceCreateParams{}
// params = sc.CreatePriceWithOptions(params, WithPriceProduct("prod_123"), WithPriceAmount(1000), WithPriceCurrency("usd"), WithPriceRecurring("month"))
// price, err := sc.CreatePriceWithParams(ctx, params)
//
// updateParams := &stripe.PriceUpdateParams{}
// updateParams = sc.UpdatePriceWithOptions(updateParams, WithUpdatePriceMetadata(map[string]string{"foo": "bar"}))
// updated, err := sc.UpdatePriceWithParams(ctx, priceID, updateParams)
