package entitlements

import (
	"testing"

	"github.com/stripe/stripe-go/v86"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestPriceCreateOptions(t *testing.T) {
	params := &stripe.PriceCreateParams{}
	params = (&StripeClient{}).CreatePriceWithOptions(params, WithPriceProduct("prod_123"), WithPriceAmount(1000), WithPriceCurrency("usd"), WithPriceRecurring("month"), WithPriceMetadata(map[string]string{"foo": "bar"}))
	assert.Check(t, is.Equal("prod_123", *params.Product))
	assert.Check(t, is.Equal(int64(1000), *params.UnitAmount))
	assert.Check(t, is.Equal("usd", *params.Currency))
	assert.Check(t, is.Equal("month", *params.Recurring.Interval))
	assert.Check(t, is.Equal("bar", params.Metadata["foo"]))
}

func TestPriceUpdateOptions(t *testing.T) {
	params := &stripe.PriceUpdateParams{}
	params = (&StripeClient{}).UpdatePriceWithOptions(params, WithUpdatePriceMetadata(map[string]string{"baz": "qux"}))
	assert.Check(t, is.Equal("qux", params.Metadata["baz"]))
}

func TestSeq2IsEmpty(t *testing.T) {
	empty := stripe.Seq2[string, error](func(yield func(string, error) bool) {})
	assert.Check(t, Seq2IsEmpty(empty))

	nonEmpty := stripe.Seq2[string, error](func(yield func(string, error) bool) {
		yield("v", nil)
	})

	assert.Check(t, !Seq2IsEmpty(nonEmpty))
}
