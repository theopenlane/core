package entitlements_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v84"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
)

// helper to create a Stripe client returning provided prices from ListPrices
func setupPriceClient(prices []*stripe.Price, err error) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		out := args.Get(4) // this is *v1SearchPage[*stripe.Price] now, but unexported

		// Build a payload that matches Stripe search response shape
		payload := map[string]any{
			"object":   "search_result",
			"data":     prices, // prices := []*stripe.Price{...}
			"has_more": false,
		}

		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, out)
	}).Return(err)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

func TestFindPriceForProductByLookupKey(t *testing.T) {
	ctx := context.Background()
	price := &stripe.Price{ID: "price_1", LookupKey: "basic", Product: &stripe.Product{ID: "prod_1"}}
	sc, _ := setupPriceClient([]*stripe.Price{price}, nil)

	found, err := sc.FindPriceForProduct(ctx, "prod_1", "", 0, "", "", "", "basic", nil)
	assert.NoError(t, err)
	assert.Equal(t, price, found)
}

func TestFindPriceForProductByAttributes(t *testing.T) {
	ctx := context.Background()
	price := &stripe.Price{
		ID:         "price_1",
		UnitAmount: 1000,
		Currency:   "usd",
		Nickname:   "basic",
		LookupKey:  "basic",
		Recurring:  &stripe.PriceRecurring{Interval: "month"},
		Product:    &stripe.Product{ID: "prod_1"},
		Metadata:   map[string]string{"managed_by": "openlane"},
	}
	sc, _ := setupPriceClient([]*stripe.Price{price}, nil)

	found, err := sc.FindPriceForProduct(ctx, "prod_1", "", 1000, "usd", "month", "basic", "", map[string]string{"managed_by": "openlane"})
	assert.NoError(t, err)
	assert.Equal(t, price, found)
}

func TestFindPriceForProductNoMatch(t *testing.T) {
	ctx := context.Background()
	price := &stripe.Price{ID: "price_1", LookupKey: "basic", Product: &stripe.Product{ID: "prod_1"}}
	sc, _ := setupPriceClient([]*stripe.Price{price}, nil)

	found, err := sc.FindPriceForProduct(ctx, "prod_1", "", 2000, "usd", "month", "", "", nil)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindPriceForProductListError(t *testing.T) {
	sc, _ := setupPriceClient(nil, assert.AnError)
	found, err := sc.FindPriceForProduct(context.Background(), "prod_1", "", 0, "", "", "", "", nil)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUpdatePriceMetadata(t *testing.T) {
	ctx := context.Background()
	expected := &stripe.Price{ID: "price_1"}

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	metadata := map[string]string{"managed_by": "openlane"}

	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		params := args.Get(3).(*stripe.PriceUpdateParams)
		assert.Equal(t, metadata, params.Metadata)
		resp := args.Get(4).(*stripe.Price)
		*resp = *expected
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	p, err := sc.UpdatePriceMetadata(ctx, "price_1", metadata)
	assert.NoError(t, err)
	assert.Equal(t, expected, p)
}
