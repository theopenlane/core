package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v83"
)

func TestProductUpdateOptions(t *testing.T) {
	params := &stripe.ProductUpdateParams{}
	params = (&StripeClient{}).UpdateProductWithOptions(params, WithUpdateProductName("NewName"), WithUpdateProductDescription("NewDesc"), WithUpdateProductMetadata(map[string]string{"baz": "qux"}))
	assert.Equal(t, "NewName", *params.Name)
	assert.Equal(t, "NewDesc", *params.Description)
	assert.Equal(t, "qux", params.Metadata["baz"])
}
