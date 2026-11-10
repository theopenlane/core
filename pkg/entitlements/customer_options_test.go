package entitlements

import (
	"testing"

	"github.com/stripe/stripe-go/v86"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestCustomerCreateOptions(t *testing.T) {
	params := &stripe.CustomerCreateParams{}
	params = (&StripeClient{}).CreateCustomerWithOptions(params, WithCustomerEmail("foo@bar.com"), WithCustomerName("Acme"), WithCustomerMetadata(map[string]string{"foo": "bar"}))
	assert.Check(t, is.Equal("foo@bar.com", *params.Email))
	assert.Check(t, is.Equal("Acme", *params.Name))
	assert.Check(t, is.Equal("bar", params.Metadata["foo"]))
}

func TestCustomerUpdateOptions(t *testing.T) {
	params := &stripe.CustomerUpdateParams{}
	params = (&StripeClient{}).UpdateCustomerWithOptions(params, WithUpdateCustomerEmail("new@bar.com"), WithUpdateCustomerMetadata(map[string]string{"baz": "qux"}))
	assert.Check(t, is.Equal("new@bar.com", *params.Email))
	assert.Check(t, is.Equal("qux", params.Metadata["baz"]))
}
