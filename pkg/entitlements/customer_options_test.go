package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v83"
)

func TestCustomerCreateOptions(t *testing.T) {
	params := &stripe.CustomerCreateParams{}
	params = (&StripeClient{}).CreateCustomerWithOptions(params, WithCustomerEmail("foo@bar.com"), WithCustomerName("Acme"), WithCustomerMetadata(map[string]string{"foo": "bar"}))
	assert.Equal(t, "foo@bar.com", *params.Email)
	assert.Equal(t, "Acme", *params.Name)
	assert.Equal(t, "bar", params.Metadata["foo"])
}

func TestCustomerUpdateOptions(t *testing.T) {
	params := &stripe.CustomerUpdateParams{}
	params = (&StripeClient{}).UpdateCustomerWithOptions(params, WithUpdateCustomerEmail("new@bar.com"), WithUpdateCustomerMetadata(map[string]string{"baz": "qux"}))
	assert.Equal(t, "new@bar.com", *params.Email)
	assert.Equal(t, "qux", params.Metadata["baz"])
}
