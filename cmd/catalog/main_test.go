//go:build ignore

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"

	"github.com/theopenlane/shared/catalog"
	"github.com/theopenlane/shared/entitlements"
	"github.com/theopenlane/shared/entitlements/mocks"
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
	oldOut := outWriter
	outWriter = w
	f()
	w.Close()
	outWriter = oldOut
	out, _ := io.ReadAll(r)
	return string(out)
}

type fakeClient struct {
	products []*stripe.Product
	prices   map[string]*stripe.Price
	features map[string]*stripe.EntitlementsFeature
	updated  []string
}

func (f *fakeClient) ListProducts(ctx context.Context) ([]*stripe.Product, error) {
	return f.products, nil
}

func (f *fakeClient) GetPrice(ctx context.Context, id string) (*stripe.Price, error) {
	return f.prices[id], nil
}

func (f *fakeClient) GetPriceByLookupKey(ctx context.Context, lookupKey string) (*stripe.Price, error) {
	for _, p := range f.prices {
		if p.LookupKey == lookupKey {
			return p, nil
		}
	}

	return nil, nil
}

func (f *fakeClient) GetFeatureByLookupKey(ctx context.Context, lookupKey string) (*stripe.EntitlementsFeature, error) {
	for _, feat := range f.features {
		if feat.LookupKey == lookupKey {
			return feat, nil
		}
	}

	return nil, nil
}

func (f *fakeClient) GetProduct(ctx context.Context, id string) (*stripe.Product, error) {
	for _, p := range f.products {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, nil
}

func (f *fakeClient) FindPriceForProduct(ctx context.Context, productID string, currency string, unitAmount int64, interval, nickname, lookupKey, metadata string, meta map[string]string) (*stripe.Price, error) {
	for _, p := range f.prices {
		if p.Product != nil && p.Product.ID == productID {
			if p.LookupKey == metadata &&
				p.UnitAmount == unitAmount &&
				(p.Nickname == lookupKey || (lookupKey == "" && p.Nickname == "")) &&
				(p.Recurring != nil && string(p.Recurring.Interval) == nickname) {
				return p, nil
			}
		}
	}

	return nil, nil
}

func (f *fakeClient) UpdatePriceMetadata(ctx context.Context, id string, md map[string]string) (*stripe.Price, error) {
	f.updated = append(f.updated, id)
	p := f.prices[id]
	if p != nil {
		p.Metadata = md
	}
	return p, nil
}

func (f *fakeClient) UpdateProductWithParams(ctx context.Context, productID string, params *stripe.ProductUpdateParams) (*stripe.Product, error) {
	for _, p := range f.products {
		if p.ID == productID {
			return p, nil
		}
	}
	return nil, nil
}

func (f *fakeClient) UpdateProductWithOptions(baseParams *stripe.ProductUpdateParams, opts ...entitlements.ProductUpdateOption) *stripe.ProductUpdateParams {
	return baseParams
}

func TestPriceMatchesStripe(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	p1 := &stripe.Product{ID: "prod_1", Name: "Basic"}
	p2 := &stripe.Product{ID: "prod_2", Name: "Pro"}
	sc, _ := setupProductClient([]*stripe.Product{p1, p2}, nil)

	m, err := buildProductMap(context.Background(), sc)
	require.NoError(t, err)
	require.Len(t, m, 4)
	assert.Equal(t, p1, m["Basic"])
	assert.Equal(t, p2, m["Pro"])
	assert.Equal(t, p1, m[p1.ID])
	assert.Equal(t, p2, m[p2.ID])
}

func TestBuildProductMapError(t *testing.T) {
	t.Parallel()

	sc, _ := setupProductClient(nil, assert.AnError)
	m, err := buildProductMap(context.Background(), sc)
	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestUpdateFeaturePrices(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

	slices.SortFunc(reports, func(a, b featureReport) int {
		return strings.Compare(a.name, b.name)
	})

	assert.Equal(t, "f1", reports[0].name)
	assert.Equal(t, "f2", reports[1].name)
}

func TestHandleTakeovers(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	reports := []featureReport{{kind: "module", name: "m1", product: true, missingPrices: 0}}
	out := captureOutput(func() { printFeatureReports(reports) })
	assert.Contains(t, out, "module")
	assert.Contains(t, out, "m1")
}

func TestPromptAndCreateMissing(t *testing.T) {
	t.Parallel()

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

func writeTempCatalogFile(t *testing.T, data string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, fmt.Sprintf("catalog-%s.yaml", ulid.Make().String()))
	require.NoError(t, os.WriteFile(p, []byte(data), 0o644))
	return p
}

func TestCatalogAppNoop(t *testing.T) {
	catYAML := `version: v1
sha: bad
modules:
  mod1:
    display_name: Prod1
    lookup_key: mod1
    description: D
    audience: public
    billing:
      prices:
        - interval: month
          unit_amount: 100
          nickname: p1_month
          lookup_key: p1_month
          price_id: price_1
addons: {}`

	path := writeTempCatalogFile(t, catYAML)

	prod := &stripe.Product{ID: "prod1", Name: "Prod1", Metadata: map[string]string{"module": "mod1"}}
	price := &stripe.Price{ID: "price_1", UnitAmount: 100, Recurring: &stripe.PriceRecurring{Interval: "month"}, LookupKey: "p1_month", Nickname: "p1_month", Product: prod, Metadata: map[string]string{catalog.ManagedByKey: catalog.ManagedByValue}}

	client := &fakeClient{products: []*stripe.Product{prod}, prices: map[string]*stripe.Price{"price_1": price}, features: map[string]*stripe.EntitlementsFeature{}}
	newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) { return client, nil }
	defer func() {
		newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
			return entitlements.NewStripeClient(opts...)
		}
	}()

	buf := &bytes.Buffer{}
	outWriter = buf
	defer func() { outWriter = os.Stdout }()

	app := catalogApp()
	err := app.Run(context.Background(), []string{"catalog", "--catalog", path, "--stripe-key", "sk", "--takeover"})
	require.NoError(t, err)
	require.Empty(t, client.updated)
}

func TestCatalogAppTakeover(t *testing.T) {
	catYAML := `version: v1
sha: bad
modules:
  mod1:
    display_name: Prod1
    lookup_key: mod1
    description: D
    audience: public
    billing:
      prices:
        - interval: month
          unit_amount: 100
          nickname: p1_month
          lookup_key: p1_month
          price_id: price_1
addons: {}`

	path := writeTempCatalogFile(t, catYAML)

	prod := &stripe.Product{ID: "prod1", Name: "Prod1", Metadata: map[string]string{"module": "mod1"}}
	price := &stripe.Price{ID: "price_1", UnitAmount: 100, Recurring: &stripe.PriceRecurring{Interval: "month"}, LookupKey: "p1_month", Nickname: "p1_month", Product: prod, Metadata: map[string]string{}}

	client := &fakeClient{products: []*stripe.Product{prod}, prices: map[string]*stripe.Price{"price_1": price}}
	newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) { return client, nil }
	defer func() {
		newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
			return entitlements.NewStripeClient(opts...)
		}
	}()

	buf := &bytes.Buffer{}
	outWriter = buf
	defer func() { outWriter = os.Stdout }()

	app := catalogApp()
	err := app.Run(context.Background(), []string{"catalog", "--catalog", path, "--stripe-key", "sk", "--takeover"})
	require.NoError(t, err)
	assert.Equal(t, []string{"price_1"}, client.updated)
}

func TestCatalogAppLookupKeyConflict(t *testing.T) {
	catYAML := `version: v1
sha: bad
modules:
  mod1:
    display_name: Prod1
    lookup_key: mod1
    description: D
    audience: public
    billing:
      prices:
        - interval: month
          unit_amount: 100
          nickname: p1_month
          lookup_key: p1_month
addons: {}`

	path := writeTempCatalogFile(t, catYAML)

	prod := &stripe.Product{ID: "prod1", Name: "Prod1", Metadata: map[string]string{"module": "mod1"}}
	price := &stripe.Price{ID: "price_conflict", UnitAmount: 100, Recurring: &stripe.PriceRecurring{Interval: "month"}, LookupKey: "p1_month", Nickname: "p1_month", Product: prod, Metadata: map[string]string{}}

	client := &fakeClient{products: []*stripe.Product{prod}, prices: map[string]*stripe.Price{"price_conflict": price}, features: map[string]*stripe.EntitlementsFeature{}}
	newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) { return client, nil }
	defer func() {
		newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
			return entitlements.NewStripeClient(opts...)
		}
	}()

	app := catalogApp()
	err := app.Run(context.Background(), []string{"catalog", "--catalog", path, "--stripe-key", "sk", "--takeover"})
	require.NoError(t, err)
	assert.Equal(t, []string{"price_conflict"}, client.updated)
}
