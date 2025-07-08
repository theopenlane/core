package catalog_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/core/pkg/catalog"
)

type fakeLookup struct {
	price   *stripe.Price
	feature *stripe.EntitlementsFeature
	product *stripe.Product
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
	return nil, nil
}

func TestLookupKeyConflicts(t *testing.T) {
	cat := &catalog.Catalog{
		Modules: catalog.FeatureSet{
			"m1": {
				LookupKey: "dup_feat",
				Billing:   catalog.Billing{Prices: []catalog.Price{{LookupKey: "dup"}}},
			},
		},
	}

	price := &stripe.Price{ID: "price_x", LookupKey: "dup"}
	feat := &stripe.EntitlementsFeature{ID: "feat_x", LookupKey: "dup_feat"}
	prod := &stripe.Product{ID: "m1"}

	conflicts, err := cat.LookupKeyConflicts(context.Background(), &fakeLookup{price: price, feature: feat, product: prod})
	require.NoError(t, err)
	require.Len(t, conflicts, 3)
}
