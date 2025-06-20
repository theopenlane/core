package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

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

// ListPricesForProduct returns all prices for a given product.
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

// FindPriceForProduct searches existing prices for a matching interval, amount and optional metadata.
func (sc *StripeClient) FindPriceForProduct(ctx context.Context, productID string, unitAmount int64, currency, interval, nickname, lookupKey string, metadata map[string]string) (*stripe.Price, error) {
	prices, err := sc.ListPricesForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	for _, p := range prices {
		if lookupKey != "" && p.LookupKey != lookupKey {
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
			if p.Metadata[k] != v {
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
