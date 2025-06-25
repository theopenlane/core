package entitlements

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

func TestCustomerCreateOptions(t *testing.T) {
	params := &stripe.CustomerCreateParams{}
	params = (&StripeClient{}).CreateCustomerWithOptions(params, WithCustomerEmail("foo@bar.com"), WithCustomerName("Acme"), WithCustomerMetadata(map[string]string{"foo": "bar"}))
	require.Equal(t, "foo@bar.com", *params.Email)
	require.Equal(t, "Acme", *params.Name)
	require.Equal(t, "bar", params.Metadata["foo"])
}

func TestCustomerUpdateOptions(t *testing.T) {
	params := &stripe.CustomerUpdateParams{}
	params = (&StripeClient{}).UpdateCustomerWithOptions(params, WithUpdateCustomerEmail("new@bar.com"), WithUpdateCustomerMetadata(map[string]string{"baz": "qux"}))
	require.Equal(t, "new@bar.com", *params.Email)
	require.Equal(t, "qux", params.Metadata["baz"])
}
