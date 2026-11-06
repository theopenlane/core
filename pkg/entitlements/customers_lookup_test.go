package entitlements_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v87"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
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
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(expected, cust))
}

func TestGetCustomerByStripeIDMissingID(t *testing.T) {
	sc, _ := setupCustomerClient(nil, nil)
	cust, err := sc.GetCustomerByStripeID(context.Background(), "")
	assert.Check(t, is.ErrorIs(err, entitlements.ErrCustomerIDRequired))
	assert.Check(t, is.Nil(cust))
}

func TestGetCustomerByStripeIDNotFound(t *testing.T) {
	sc, _ := setupCustomerClient(nil, &stripe.Error{Code: stripe.ErrorCodeMissing})
	cust, err := sc.GetCustomerByStripeID(context.Background(), "cus_404")
	assert.Check(t, is.ErrorIs(err, entitlements.ErrCustomerNotFound))
	assert.Check(t, is.Nil(cust))
}

func TestGetCustomerByStripeIDLookupErr(t *testing.T) {
	sc, _ := setupCustomerClient(nil, errMock)
	cust, err := sc.GetCustomerByStripeID(context.Background(), "cus_1")
	assert.Check(t, is.ErrorIs(err, entitlements.ErrCustomerLookupFailed))
	assert.Check(t, is.Nil(cust))
}
