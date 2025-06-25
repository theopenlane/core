package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

// CreatePrice a price for a product in stripe
func (sc *StripeClient) CreatePrice(ctx context.Context, productID string, unitAmount int64, currency, interval string) (*stripe.Price, error) {
	params := &stripe.PriceCreateParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
		Recurring: &stripe.PriceCreateRecurringParams{
			Interval: stripe.String(interval),
		},
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
		// TODO: see if we need to actually handle errors here
		// note: we are ignoring any errors here, and returning prices that are valid
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
