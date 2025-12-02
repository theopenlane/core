package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v83"
)

func TestPriceCreateOptions(t *testing.T) {
	params := &stripe.PriceCreateParams{}
	params = (&StripeClient{}).CreatePriceWithOptions(params, WithPriceProduct("prod_123"), WithPriceAmount(1000), WithPriceCurrency("usd"), WithPriceRecurring("month"), WithPriceMetadata(map[string]string{"foo": "bar"}))
	assert.Equal(t, "prod_123", *params.Product)
	assert.Equal(t, int64(1000), *params.UnitAmount)
	assert.Equal(t, "usd", *params.Currency)
	assert.Equal(t, "month", *params.Recurring.Interval)
	assert.Equal(t, "bar", params.Metadata["foo"])
}

func TestPriceUpdateOptions(t *testing.T) {
	params := &stripe.PriceUpdateParams{}
	params = (&StripeClient{}).UpdatePriceWithOptions(params, WithUpdatePriceMetadata(map[string]string{"baz": "qux"}))
	assert.Equal(t, "qux", params.Metadata["baz"])
}

func TestSeq2IsEmpty(t *testing.T) {
	empty := stripe.Seq2[string, error](func(yield func(string, error) bool) {})
	assert.True(t, Seq2IsEmpty(empty))

	nonEmpty := stripe.Seq2[string, error](func(yield func(string, error) bool) {
		yield("v", nil)
	})

	assert.False(t, Seq2IsEmpty(nonEmpty))
}
