package entitlements_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v82"
	"github.com/theopenlane/core/pkg/entitlements"
)

func TestGetUpdatedFields(t *testing.T) {
	tests := []struct {
		name           string
		props          map[string]interface{}
		stripeCustomer *entitlements.OrganizationCustomer
		expectedParams *stripe.CustomerUpdateParams
	}{
		{
			name:  "No updates",
			props: map[string]interface{}{},
			stripeCustomer: &entitlements.OrganizationCustomer{
				ContactInfo: entitlements.ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerUpdateParams{},
		},
		{
			name: "Update email",
			props: map[string]interface{}{
				"billing_email": "new@example.com",
			},
			stripeCustomer: &entitlements.OrganizationCustomer{
				ContactInfo: entitlements.ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerUpdateParams{
				Email: stripe.String("test@example.com"),
			},
		},
		{
			name: "Update phone",
			props: map[string]interface{}{
				"billing_phone": "1234567890",
			},
			stripeCustomer: &entitlements.OrganizationCustomer{
				ContactInfo: entitlements.ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerUpdateParams{
				Phone: stripe.String("1234567890"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := entitlements.GetUpdatedFields(tt.props, tt.stripeCustomer)
			assert.Equal(t, tt.expectedParams, params)
		})
	}
}
