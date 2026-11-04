package entitlements

import (
	"testing"

	"github.com/stripe/stripe-go/v86"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestProductUpdateOptions(t *testing.T) {
	params := &stripe.ProductUpdateParams{}
	params = (&StripeClient{}).UpdateProductWithOptions(params, WithUpdateProductName("NewName"), WithUpdateProductDescription("NewDesc"), WithUpdateProductMetadata(map[string]string{"baz": "qux"}))
	assert.Check(t, is.Equal("NewName", *params.Name))
	assert.Check(t, is.Equal("NewDesc", *params.Description))
	assert.Check(t, is.Equal("qux", params.Metadata["baz"]))
}
