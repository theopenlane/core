package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v81"
)

func TestCheckForBillingUpdate(t *testing.T) {
	tests := []struct {
		name           string
		props          map[string]interface{}
		stripeCustomer *OrganizationCustomer
		expectedParams *stripe.CustomerParams
		expectedUpdate bool
	}{
		{
			name: "No updates",
			props: map[string]interface{}{
				"billing_email": "test@example.com",
				"billing_phone": "1234567890",
			},
			stripeCustomer: &OrganizationCustomer{
				ContactInfo: ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerParams{},
			expectedUpdate: false,
		},
		{
			name: "Update email",
			props: map[string]interface{}{
				"billing_email": "new@example.com",
				"billing_phone": "1234567890",
			},
			stripeCustomer: &OrganizationCustomer{
				ContactInfo: ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerParams{
				Email: stripe.String("new@example.com"),
			},
			expectedUpdate: true,
		},
		{
			name: "Update phone",
			props: map[string]interface{}{
				"billing_email": "test@example.com",
				"billing_phone": "0987654321",
			},
			stripeCustomer: &OrganizationCustomer{
				ContactInfo: ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerParams{
				Phone: stripe.String("0987654321"),
			},
			expectedUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, hasUpdate := CheckForBillingUpdate(tt.props, tt.stripeCustomer)
			assert.Equal(t, tt.expectedUpdate, hasUpdate)
			assert.Equal(t, tt.expectedParams, params)
		})
	}
}
