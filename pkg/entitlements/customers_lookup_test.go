package entitlements_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v84"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
)

// helper to create Stripe client for customer retrieval
func setupCustomerClient(cust *stripe.Customer, err error) (*entitlements.StripeClient, *mocks.MockStripeBackend) {
	backend := new(mocks.MockStripeBackend)
	backends := &stripe.Backends{API: backend, Connect: backend, Uploads: backend}

	backend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if v, ok := args.Get(4).(*stripe.Customer); ok && err == nil {
				*v = *cust
			}
		}).Return(err)

	sc := entitlements.StripeClient{Client: stripe.NewClient("sk_test", stripe.WithBackends(backends))}
	return &sc, backend
}

func TestGetCustomerByStripeID(t *testing.T) {
	ctx := context.Background()
	expected := &stripe.Customer{ID: "cus_123"}
	sc, _ := setupCustomerClient(expected, nil)

	cust, err := sc.GetCustomerByStripeID(ctx, "cus_123")
	assert.NoError(t, err)
	assert.Equal(t, expected, cust)
}

func TestGetCustomerByStripeIDMissingID(t *testing.T) {
	sc, _ := setupCustomerClient(nil, nil)
	cust, err := sc.GetCustomerByStripeID(context.Background(), "")
	assert.ErrorIs(t, err, entitlements.ErrCustomerIDRequired)
	assert.Nil(t, cust)
}

func TestGetCustomerByStripeIDNotFound(t *testing.T) {
	sc, _ := setupCustomerClient(nil, &stripe.Error{Code: stripe.ErrorCodeMissing})
	cust, err := sc.GetCustomerByStripeID(context.Background(), "cus_404")
	assert.ErrorIs(t, err, entitlements.ErrCustomerNotFound)
	assert.Nil(t, cust)
}

func TestGetCustomerByStripeIDLookupErr(t *testing.T) {
	sc, _ := setupCustomerClient(nil, assert.AnError)
	cust, err := sc.GetCustomerByStripeID(context.Background(), "cus_1")
	assert.ErrorIs(t, err, entitlements.ErrCustomerLookupFailed)
	assert.Nil(t, cust)
}
