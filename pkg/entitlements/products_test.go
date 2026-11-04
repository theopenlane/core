package entitlements_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v86"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
)

// helper to stage product list responses
func setupProductClient(products []*stripe.Product, err error) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			out := args.Get(4) // this is *v1SearchPage[*stripe.Product] now, but unexported

			// Build a payload that matches Stripe search response shape
			payload := map[string]any{
				"object":   "search_result",
				"data":     products, // products := []*stripe.Product{...}
				"has_more": false,
			}

			b, _ := json.Marshal(payload)
			_ = json.Unmarshal(b, out)
		}).Return(err)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

func TestListProducts(t *testing.T) {
	ctx := context.Background()
	product := &stripe.Product{ID: "prod_1", Name: "Basic"}
	sc, _ := setupProductClient([]*stripe.Product{product}, nil)

	products, err := sc.ListProducts(ctx)
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual([]*stripe.Product{product}, products))
}

func TestListProductsError(t *testing.T) {
	sc, _ := setupProductClient(nil, errMock)
	products, err := sc.ListProducts(context.Background())
	assert.Check(t, err != nil)
	assert.Check(t, is.Nil(products))
}
