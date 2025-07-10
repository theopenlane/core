package entitlements

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

func TestPriceCreateOptions(t *testing.T) {
	params := &stripe.PriceCreateParams{}
	params = (&StripeClient{}).CreatePriceWithOptions(params, WithPriceProduct("prod_123"), WithPriceAmount(1000), WithPriceCurrency("usd"), WithPriceRecurring("month"), WithPriceMetadata(map[string]string{"foo": "bar"}))
	require.Equal(t, "prod_123", *params.Product)
	require.Equal(t, int64(1000), *params.UnitAmount)
	require.Equal(t, "usd", *params.Currency)
	require.Equal(t, "month", *params.Recurring.Interval)
	require.Equal(t, "bar", params.Metadata["foo"])
}

func TestPriceUpdateOptions(t *testing.T) {
	params := &stripe.PriceUpdateParams{}
	params = (&StripeClient{}).UpdatePriceWithOptions(params, WithUpdatePriceMetadata(map[string]string{"baz": "qux"}))
	require.Equal(t, "qux", params.Metadata["baz"])
}

func TestSeq2IsEmpty(t *testing.T) {
	empty := stripe.Seq2[string, error](func(yield func(string, error) bool) {})
	require.True(t, Seq2IsEmpty(empty))

	nonEmpty := stripe.Seq2[string, error](func(yield func(string, error) bool) {
		yield("v", nil)
	})

	require.True(t, Seq2IsEmpty(nonEmpty))
}
