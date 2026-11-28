package entitlements_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v87"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
)

// helper to create a Stripe client returning provided features from ListFeatures
func setupFeatureClient(features []*stripe.EntitlementsFeature, err error) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// This is actually *v1Page[*stripe.EntitlementsFeature], but it's unexported.
		out := args.Get(4)

		// Minimal list-shaped payload Stripe expects
		payload := map[string]any{
			"object":   "list",
			"data":     features, // []*stripe.EntitlementsFeature
			"has_more": false,
		}

		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, out)
	}).Return(err)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

func TestGetFeatureByLookupKey(t *testing.T) {
	ctx := context.Background()
	feature := &stripe.EntitlementsFeature{ID: "feat_1", LookupKey: "basic"}
	sc, _ := setupFeatureClient([]*stripe.EntitlementsFeature{feature}, nil)

	found, err := sc.GetFeatureByLookupKey(ctx, "basic")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(feature, found))
}

func TestGetFeatureByLookupKeyNotFound(t *testing.T) {
	sc, _ := setupFeatureClient(nil, nil)
	feat, err := sc.GetFeatureByLookupKey(context.Background(), "missing")
	assert.Check(t, err)
	assert.Check(t, is.Nil(feat))
}

func TestGetFeatureByLookupKeyError(t *testing.T) {
	sc, _ := setupFeatureClient(nil, errMock)
	feat, err := sc.GetFeatureByLookupKey(context.Background(), "err")
	assert.Check(t, err != nil)
	assert.Check(t, is.Nil(feat))
}
