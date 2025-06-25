package entitlements_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
)

func TestTagPriceMigration(t *testing.T) {
	ctx := context.Background()

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	meta := map[string]string{"migrate_to": "price_new"}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		params := args.Get(3).(*stripe.PriceUpdateParams)
		assert.Equal(t, meta, params.Metadata)
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	err := sc.TagPriceMigration(ctx, "price_old", "price_new")
	require.NoError(t, err)
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
		require.Len(t, params.Items, 1)
		assert.Equal(t, "si_1", *params.Items[0].ID)
		assert.Equal(t, "price_new", *params.Items[0].Price)
		resp := args.Get(4).(*stripe.Subscription)
		*resp = *sub
	}).Return(nil)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	updated, err := sc.MigrateSubscriptionPrice(ctx, sub, "price_old", "price_new")
	require.NoError(t, err)
	assert.Equal(t, sub, updated)
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
	require.NoError(t, err)
	assert.Equal(t, sub, updated)
	backend.AssertExpectations(t)
}

func TestTagPriceMigrationError(t *testing.T) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	err := sc.TagPriceMigration(context.Background(), "price_old", "price_new")
	assert.Error(t, err)
}

func TestMigrateSubscriptionPriceError(t *testing.T) {
	sub := &stripe.Subscription{ID: "sub_123", Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
		{ID: "si_1", Price: &stripe.Price{ID: "price_old"}},
	}}}

	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}
	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}

	updated, err := sc.MigrateSubscriptionPrice(context.Background(), sub, "price_old", "price_new")
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestMigrateSubscriptionPriceNil(t *testing.T) {
	sc := entitlements.StripeClient{}
	updated, err := sc.MigrateSubscriptionPrice(context.Background(), nil, "old", "new")
	assert.NoError(t, err)
	assert.Nil(t, updated)
}
