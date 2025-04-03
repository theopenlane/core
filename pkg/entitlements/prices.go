package entitlements

import "github.com/stripe/stripe-go/v81"

// CreatePrice a price for a product in stripe
func (sc *StripeClient) CreatePrice(productID string, unitAmount int64, currency, interval string) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(interval),
		},
	}

	return sc.Client.Prices.New(params)
}

// GetPrice retrieves a price from stripe by its ID
func (sc *StripeClient) GetPrice(priceID string) (*stripe.Price, error) {
	return sc.Client.Prices.Get(priceID, nil)
}

// ListPrices retrieves all prices from stripe
func (sc *StripeClient) ListPrices() ([]*stripe.Price, error) {
	var prices []*stripe.Price
	params := &stripe.PriceListParams{}
	i := sc.Client.Prices.List(params)

	for i.Next() {
		p := i.Price()
		prices = append(prices, p)
	}

	if err := i.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}

// GetPricesMapped retrieves all prices from stripe which are active and maps them into a []Price struct
func (sc *StripeClient) GetPricesMapped() []Price {
	priceParams := &stripe.PriceListParams{}

	iter := sc.Client.Prices.List(priceParams)
	prices := []Price{}

	for iter.Next() {
		priceData := iter.Price()
		if priceData.Product == nil {
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
