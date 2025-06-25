package main

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
)

// helper to create a stripe client returning provided products
func setupProductClient(products []*stripe.Product, err error) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if v, ok := args.Get(4).(*stripe.ProductList); ok && err == nil {
			*v = stripe.ProductList{Data: products, ListMeta: stripe.ListMeta{HasMore: false}}
		}
	}).Return(err)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

// helper to create a stripe client returning provided prices
func setupPriceClient(price *stripe.Price, list []*stripe.Price) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if v, ok := args.Get(4).(*stripe.Price); ok {
			if price != nil {
				*v = *price
			}
		}
	}).Return(nil)

	backend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if v, ok := args.Get(4).(*stripe.PriceList); ok {
			*v = stripe.PriceList{Data: list, ListMeta: stripe.ListMeta{HasMore: false}}
		}
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPriceMatchesStripe(t *testing.T) {
	prod := &stripe.Product{ID: "prod_1"}
	sp := &stripe.Price{
		ID:         "price_1",
		UnitAmount: 1000,
		Recurring:  &stripe.PriceRecurring{Interval: "month"},
		Nickname:   "basic",
		LookupKey:  "basic",
		Metadata:   map[string]string{"managed_by": "module-manager"},
		Product:    prod,
	}

	cp := catalog.Price{
		UnitAmount: 1000,
		Interval:   "month",
		Nickname:   "basic",
		LookupKey:  "basic",
		Metadata:   map[string]string{"managed_by": "module-manager"},
	}

	assert.True(t, priceMatchesStripe(sp, cp, prod.ID))

	cp.UnitAmount = 2000
	assert.False(t, priceMatchesStripe(sp, cp, prod.ID))
}

func TestBuildProductMap(t *testing.T) {
	p1 := &stripe.Product{ID: "prod_1", Name: "Basic"}
	p2 := &stripe.Product{ID: "prod_2", Name: "Pro"}
	sc, _ := setupProductClient([]*stripe.Product{p1, p2}, nil)

	m, err := buildProductMap(context.Background(), sc)
	require.NoError(t, err)
	require.Len(t, m, 2)
	assert.Equal(t, p1, m["Basic"])
	assert.Equal(t, p2, m["Pro"])
}

func TestBuildProductMapError(t *testing.T) {
	sc, _ := setupProductClient(nil, assert.AnError)
	m, err := buildProductMap(context.Background(), sc)
	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestUpdateFeaturePrices(t *testing.T) {
	prod := &stripe.Product{ID: "prod_1"}
	priceA := &stripe.Price{
		ID:         "priceA",
		UnitAmount: 1000,
		Recurring:  &stripe.PriceRecurring{Interval: "month"},
		Product:    prod,
		Metadata:   map[string]string{},
	}
	priceB := &stripe.Price{
		ID:         "priceB",
		UnitAmount: 2000,
		Recurring:  &stripe.PriceRecurring{Interval: "month"},
		Product:    prod,
		Metadata:   map[string]string{catalog.ManagedByKey: catalog.ManagedByValue},
	}

	sc, _ := setupPriceClient(priceA, []*stripe.Price{priceB})

	feat := catalog.Feature{
		DisplayName: "FeatureA",
		Billing: catalog.Billing{Prices: []catalog.Price{
			{UnitAmount: 1000, Interval: "month", PriceID: "priceA"},
			{UnitAmount: 2000, Interval: "month"},
		}},
	}

	updated, takeovers, missing := updateFeaturePrices(context.Background(), sc, prod, "f1", feat)
	require.Equal(t, 0, missing)
	assert.Len(t, takeovers, 1)
	assert.Equal(t, "priceA", updated.Billing.Prices[0].PriceID)
	assert.Equal(t, "priceB", updated.Billing.Prices[1].PriceID)
}

func TestProcessFeatureSet(t *testing.T) {
	prod := &stripe.Product{ID: "prod_1"}
	price := &stripe.Price{
		ID:         "price1",
		UnitAmount: 1000,
		Recurring:  &stripe.PriceRecurring{Interval: "month"},
		Product:    prod,
		Metadata:   map[string]string{},
	}
	sc, _ := setupPriceClient(price, []*stripe.Price{price})

	fs := catalog.FeatureSet{
		"f1": {DisplayName: "prod_1", Billing: catalog.Billing{Prices: []catalog.Price{{UnitAmount: 1000, Interval: "month", PriceID: "price1"}}}},
		"f2": {DisplayName: "missing", Billing: catalog.Billing{Prices: []catalog.Price{{UnitAmount: 500, Interval: "month"}}}},
	}

	prodMap := map[string]*stripe.Product{"prod_1": prod}

	takeovers, missing, reports := processFeatureSet(context.Background(), sc, prodMap, "module", fs)
	assert.True(t, missing)
	assert.Len(t, takeovers, 1)
	assert.Len(t, reports, 2)
	assert.Equal(t, "f1", reports[0].name)
}

func TestHandleTakeovers(t *testing.T) {
	prod := &stripe.Product{ID: "prod"}
	price := &stripe.Price{ID: "price1", Product: prod, Metadata: map[string]string{}}
	sc, backend := setupPriceClient(price, nil)

	take := []takeoverInfo{{feature: "feat", price: catalog.Price{}, stripe: price}}
	takeover := true
	updated := false

	backend.ExpectedCalls = nil
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		updated = true
		resp := args.Get(4).(*stripe.Price)
		*resp = *price
	}).Return(nil)

	err := handleTakeovers(context.Background(), sc, take, &takeover)
	require.NoError(t, err)
	assert.True(t, updated)
}

func TestPrintFeatureReports(t *testing.T) {
	reports := []featureReport{{kind: "module", name: "m1", product: true, missingPrices: 0}}
	out := captureOutput(func() { printFeatureReports(reports) })
	assert.Contains(t, out, "module")
	assert.Contains(t, out, "m1")
}

func TestPromptAndCreateMissing(t *testing.T) {
	sc, _ := setupProductClient(nil, nil)
	cat := &catalog.Catalog{}

	old := os.Stdin
	r, w, _ := os.Pipe()
	w.WriteString("y\n")
	w.Close()
	os.Stdin = r
	defer func() { os.Stdin = old }()

	err := promptAndCreateMissing(context.Background(), cat, sc)
	require.NoError(t, err)
}
