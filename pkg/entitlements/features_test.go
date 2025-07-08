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

// helper to create a Stripe client returning provided features from ListFeatures
func setupFeatureClient(features []*stripe.EntitlementsFeature, err error) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if v, ok := args.Get(4).(*stripe.EntitlementsFeatureList); ok && err == nil {
			*v = stripe.EntitlementsFeatureList{Data: features, ListMeta: stripe.ListMeta{HasMore: false}}
		}
	}).Return(err)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

func TestGetFeatureByLookupKey(t *testing.T) {
	ctx := context.Background()
	feature := &stripe.EntitlementsFeature{ID: "feat_1", LookupKey: "basic"}
	sc, _ := setupFeatureClient([]*stripe.EntitlementsFeature{feature}, nil)

	found, err := sc.GetFeatureByLookupKey(ctx, "basic")
	require.NoError(t, err)
	assert.Equal(t, feature, found)
}

func TestGetFeatureByLookupKeyNotFound(t *testing.T) {
	sc, _ := setupFeatureClient(nil, nil)
	feat, err := sc.GetFeatureByLookupKey(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, feat)
}

func TestGetFeatureByLookupKeyError(t *testing.T) {
	sc, _ := setupFeatureClient(nil, assert.AnError)
	feat, err := sc.GetFeatureByLookupKey(context.Background(), "err")
	assert.Error(t, err)
	assert.Nil(t, feat)
}
