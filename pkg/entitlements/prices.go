package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v83"
)

// MigrateToKey is the metadata key used to indicate the price that a subscription should migrate to
const MigrateToKey = "migrate_to"

// UpsellToKey is the metadata key used to denote the price that another price should upsell into
const UpsellToKey = "upsell_to"

// CreatePrice a price for a product in stripe
func (sc *StripeClient) CreatePrice(ctx context.Context, productID string, unitAmount int64, currency, interval, nickname, lookupKey string, metadata map[string]string) (*stripe.Price, error) {
	params := &stripe.PriceCreateParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
		Recurring: &stripe.PriceCreateRecurringParams{
			Interval: stripe.String(interval),
		},
		Nickname:  stripe.String(nickname),
		LookupKey: stripe.String(lookupKey),
	}

	if len(metadata) > 0 {
		params.Metadata = metadata
	}

	return sc.Client.V1Prices.Create(ctx, params)
}

// CreatePriceWithParams creates a price with the given parameters
func (sc *StripeClient) CreatePriceWithParams(ctx context.Context, params *stripe.PriceCreateParams) (*stripe.Price, error) {
	return sc.Client.V1Prices.Create(ctx, params)
}

// GetPrice retrieves a price from stripe by its ID
func (sc *StripeClient) GetPrice(ctx context.Context, priceID string) (*stripe.Price, error) {
	return sc.Client.V1Prices.Retrieve(ctx, priceID, nil)
}

// ListPrices retrieves all prices from stripe
func (sc *StripeClient) ListPrices(ctx context.Context) ([]*stripe.Price, error) {
	var prices []*stripe.Price

	params := &stripe.PriceListParams{}
	result := sc.Client.V1Prices.List(ctx, params)

	for price, err := range result {
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}

	return prices, nil
}

// GetPricesMapped retrieves all prices from stripe which are active and maps them into a []Price struct
func (sc *StripeClient) GetPricesMapped(ctx context.Context) (prices []Price) {
	priceParams := &stripe.PriceListParams{}

	result := sc.Client.V1Prices.List(ctx, priceParams)

	for priceData, err := range result {
		if err != nil || priceData.Product == nil {
			continue
		}

		prices = append(prices, Price{
			ID:        priceData.ID,
			Price:     float64(priceData.UnitAmount) / 100, // nolint:mnd
			ProductID: priceData.Product.ID,
			Interval:  string(priceData.Recurring.Interval),
		})
	}

	return prices
}

// ListPricesForProduct returns all prices for a given product
func (sc *StripeClient) ListPricesForProduct(ctx context.Context, productID string) ([]*stripe.Price, error) {
	params := &stripe.PriceListParams{Product: stripe.String(productID)}

	var prices []*stripe.Price

	it := sc.Client.V1Prices.List(ctx, params)

	for price, err := range it {
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}

	return prices, nil
}

// UpdatePriceMetadata updates the metadata for a price
func (sc *StripeClient) UpdatePriceMetadata(ctx context.Context, priceID string, metadata map[string]string) (*stripe.Price, error) {
	params := &stripe.PriceUpdateParams{Metadata: metadata}

	return sc.Client.V1Prices.Update(ctx, priceID, params)
}

// GetPriceByLookupKey returns the first price matching the provided lookup key.
func (sc *StripeClient) GetPriceByLookupKey(ctx context.Context, lookupKey string) (*stripe.Price, error) {
	params := &stripe.PriceListParams{
		LookupKeys: stripe.StringSlice([]string{lookupKey}),
	}
	params.Limit = stripe.Int64(1)

	it := sc.Client.V1Prices.List(ctx, params)

	for price, err := range it {
		if err != nil {
			return nil, err
		}

		return price, nil
	}

	return nil, nil
}

// FindPriceForProduct searches existing prices for a matching interval, amount
// and optional metadata. It prefers locating by price ID or lookup key before
// falling back to attribute matching within the specified product.
func (sc *StripeClient) FindPriceForProduct(ctx context.Context, productID, priceID string, unitAmount int64, currency, interval, nickname, lookupKey string, metadata map[string]string) (*stripe.Price, error) {
	if priceID != "" {
		price, err := sc.GetPrice(ctx, priceID)
		if err != nil {
			return nil, err
		}

		if price != nil {
			if productID == "" || (price.Product != nil && price.Product.ID == productID) {
				return price, nil
			}
		}
	}

	if lookupKey != "" {
		price, err := sc.GetPriceByLookupKey(ctx, lookupKey)
		if err != nil {
			return nil, err
		}

		if price != nil {
			if productID == "" || (price.Product != nil && price.Product.ID == productID) {
				return price, nil
			}
		}
	}

	prices, err := sc.ListPricesForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	for _, p := range prices {
		if lookupKey != "" {
			if p.LookupKey == lookupKey {
				return p, nil
			}

			continue
		}

		if nickname != "" && p.Nickname != nickname {
			continue
		}

		if interval != "" {
			if p.Recurring == nil || string(p.Recurring.Interval) != interval {
				continue
			}
		}

		if unitAmount != 0 && p.UnitAmount != unitAmount {
			continue
		}

		if currency != "" && string(p.Currency) != currency {
			continue
		}

		match := true

		for k, v := range metadata {
			if p.Metadata == nil || p.Metadata[k] != v {
				match = false
				break
			}
		}

		if match {
			return p, nil
		}
	}

	return nil, nil
}

// TagPriceMigration annotates a price in preparation for migration to a new price
func (sc *StripeClient) TagPriceMigration(ctx context.Context, fromPriceID, toPriceID string) error {
	md := map[string]string{MigrateToKey: toPriceID}

	_, err := sc.UpdatePriceMetadata(ctx, fromPriceID, md)

	return err
}

// TagPriceUpsell annotates a price so that consumers can upsell to a new price.
func (sc *StripeClient) TagPriceUpsell(ctx context.Context, fromPriceID, toPriceID string) error {
	md := map[string]string{UpsellToKey: toPriceID}

	_, err := sc.UpdatePriceMetadata(ctx, fromPriceID, md)

	return err
}
