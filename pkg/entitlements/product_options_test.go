package entitlements

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

func TestProductUpdateOptions(t *testing.T) {
	params := &stripe.ProductUpdateParams{}
	params = (&StripeClient{}).UpdateProductWithOptions(params, WithUpdateProductName("NewName"), WithUpdateProductDescription("NewDesc"), WithUpdateProductMetadata(map[string]string{"baz": "qux"}))
	require.Equal(t, "NewName", *params.Name)
	require.Equal(t, "NewDesc", *params.Description)
	require.Equal(t, "qux", params.Metadata["baz"])
}
