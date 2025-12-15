package entitlements

import (
	"github.com/stripe/stripe-go/v84"
)

// ProductUpdateOption allows customizing ProductUpdateParams
type ProductUpdateOption func(params *stripe.ProductUpdateParams)

func WithUpdateProductName(name string) ProductUpdateOption {
	return func(params *stripe.ProductUpdateParams) {
		params.Name = stripe.String(name)
	}
}

func WithUpdateProductDescription(desc string) ProductUpdateOption {
	return func(params *stripe.ProductUpdateParams) {
		params.Description = stripe.String(desc)
	}
}

func WithUpdateProductMetadata(metadata map[string]string) ProductUpdateOption {
	return func(params *stripe.ProductUpdateParams) {
		params.Metadata = metadata
	}
}

// UpdateProductWithOptions creates update params with functional options
func (sc *StripeClient) UpdateProductWithOptions(baseParams *stripe.ProductUpdateParams, opts ...ProductUpdateOption) *stripe.ProductUpdateParams {
	params := baseParams
	for _, opt := range opts {
		opt(params)
	}

	return params
}

// --- Example Usage ---
// params := NewProductCreate(WithProductID("prod_123")).WithName("My Product").WithDescription("desc").Params
// product, err := sc.CreateProductWithParams(ctx, params)
//
// updateParams := &stripe.ProductUpdateParams{}
// updateParams = sc.UpdateProductWithOptions(updateParams, WithUpdateProductName("New Name"))
// updated, err := sc.UpdateProductWithParams(ctx, productID, updateParams)
