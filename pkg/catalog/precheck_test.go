package catalog_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v84"

	"github.com/theopenlane/core/pkg/catalog"
)

type fakeLookup struct {
	price    *stripe.Price
	feature  *stripe.EntitlementsFeature
	product  *stripe.Product
	products []*stripe.Product
}

func (f *fakeLookup) GetPriceByLookupKey(ctx context.Context, key string) (*stripe.Price, error) {
	if f.price != nil && f.price.LookupKey == key {
		return f.price, nil
	}
	return nil, nil
}

func (f *fakeLookup) GetFeatureByLookupKey(ctx context.Context, key string) (*stripe.EntitlementsFeature, error) {
	if f.feature != nil && f.feature.LookupKey == key {
		return f.feature, nil
	}
	return nil, nil
}

func (f *fakeLookup) GetProduct(ctx context.Context, id string) (*stripe.Product, error) {
	if f.product != nil && f.product.ID == id {
		return f.product, nil
	}
	return nil, &stripe.Error{HTTPStatusCode: http.StatusNotFound, Code: stripe.ErrorCodeResourceMissing}
}

func (f *fakeLookup) ListProducts(ctx context.Context) ([]*stripe.Product, error) {
	return f.products, nil
}

func TestLookupKeyConflicts(t *testing.T) {
	t.Parallel()
	cat := &catalog.Catalog{
		Modules: catalog.FeatureSet{
			"m1": {
				LookupKey: "dup_feat",
				Billing:   catalog.Billing{Prices: []catalog.Price{{LookupKey: "dup"}}},
				ProductID: "m1",
			},
		},
	}

	price := &stripe.Price{ID: "price_x", LookupKey: "dup"}
	feat := &stripe.EntitlementsFeature{ID: "feat_x", LookupKey: "dup_feat"}
	prod := &stripe.Product{ID: "m1", Name: "mod1"}

	conflicts, err := cat.LookupKeyConflicts(context.Background(), &fakeLookup{price: price, feature: feat, product: prod, products: []*stripe.Product{prod}})
	assert.NoError(t, err)
	assert.Len(t, conflicts, 3)
}

func TestLookupKeyConflictsByName(t *testing.T) {
	t.Parallel()
	cat := &catalog.Catalog{
		Modules: catalog.FeatureSet{
			"m2": {
				DisplayName: "mod2",
				Billing:     catalog.Billing{},
			},
		},
	}

	prod := &stripe.Product{ID: "prod_x", Name: "mod2"}

	conflicts, err := cat.LookupKeyConflicts(context.Background(), &fakeLookup{products: []*stripe.Product{prod}})
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
}
