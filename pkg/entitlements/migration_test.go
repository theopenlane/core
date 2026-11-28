package entitlements_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v87"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
)

var errMock = errors.New("mock error")

func TestTagPriceMigration(t *testing.T) {
	ctx := context.Background()

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	meta := map[string]string{entitlements.MigrateToKey: "price_new"}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		params := args.Get(3).(*stripe.PriceUpdateParams)
		assert.Check(t, is.DeepEqual(meta, params.Metadata))
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	err := sc.TagPriceMigration(ctx, "price_old", "price_new")
	assert.Check(t, err)
	backend.AssertExpectations(t)
}

func TestMigrateSubscriptionPrice(t *testing.T) {
	ctx := context.Background()

	sub := &stripe.Subscription{
		ID: "sub_123",
		Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
			{ID: "si_1", Price: &stripe.Price{ID: "price_old"}},
			{ID: "si_2", Price: &stripe.Price{ID: "other"}},
		}},
	}

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		params := args.Get(3).(*stripe.SubscriptionUpdateParams)
		assert.Check(t, is.Len(params.Items, 1))
		assert.Check(t, is.Equal("si_1", *params.Items[0].ID))
		assert.Check(t, is.Equal("price_new", *params.Items[0].Price))
		resp := args.Get(4).(*stripe.Subscription)
		*resp = *sub
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	updated, err := sc.MigrateSubscriptionPrice(ctx, sub, "price_old", "price_new")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(sub, updated))
	backend.AssertExpectations(t)
}

func TestMigrateSubscriptionPriceNoMatch(t *testing.T) {
	ctx := context.Background()

	sub := &stripe.Subscription{ID: "sub_123", Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
		{ID: "si_1", Price: &stripe.Price{ID: "other"}},
	}}}

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}
	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	updated, err := sc.MigrateSubscriptionPrice(ctx, sub, "price_old", "price_new")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(sub, updated))
	backend.AssertExpectations(t)
}

func TestTagPriceMigrationError(t *testing.T) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errMock)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	err := sc.TagPriceMigration(context.Background(), "price_old", "price_new")
	assert.Check(t, err != nil)
}

func TestMigrateSubscriptionPriceError(t *testing.T) {
	sub := &stripe.Subscription{ID: "sub_123", Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
		{ID: "si_1", Price: &stripe.Price{ID: "price_old"}},
	}}}

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errMock)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	updated, err := sc.MigrateSubscriptionPrice(context.Background(), sub, "price_old", "price_new")
	assert.Check(t, err != nil)
	assert.Check(t, is.Nil(updated))
}

func TestMigrateSubscriptionPriceNil(t *testing.T) {
	sc := entitlements.StripeClient{}
	updated, err := sc.MigrateSubscriptionPrice(context.Background(), nil, "old", "new")
	assert.Check(t, err)
	assert.Check(t, is.Nil(updated))
}

func TestTagPriceUpsellError(t *testing.T) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errMock)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	err := sc.TagPriceUpsell(context.Background(), "price_old", "price_new")
	assert.Check(t, err != nil)
}

func TestTagPriceUpsell(t *testing.T) {
	ctx := context.Background()

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	meta := map[string]string{entitlements.UpsellToKey: "price_new"}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		params := args.Get(3).(*stripe.PriceUpdateParams)
		assert.Check(t, is.DeepEqual(meta, params.Metadata))
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	err := sc.TagPriceUpsell(ctx, "price_old", "price_new")
	assert.Check(t, err)
	backend.AssertExpectations(t)
}
