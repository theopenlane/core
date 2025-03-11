package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v81"
)

func TestGetUpdatedFields(t *testing.T) {
	tests := []struct {
		name           string
		props          map[string]interface{}
		stripeCustomer *OrganizationCustomer
		expectedParams *stripe.CustomerParams
	}{
		{
			name:  "No updates",
			props: map[string]interface{}{},
			stripeCustomer: &OrganizationCustomer{
				ContactInfo: ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerParams{},
		},
		{
			name: "Update email",
			props: map[string]interface{}{
				"billing_email": "new@example.com",
			},
			stripeCustomer: &OrganizationCustomer{
				ContactInfo: ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerParams{
				Email: stripe.String("test@example.com"),
			},
		},
		{
			name: "Update phone",
			props: map[string]interface{}{
				"billing_phone": "1234567890",
			},
			stripeCustomer: &OrganizationCustomer{
				ContactInfo: ContactInfo{
					Email: "test@example.com",
					Phone: "1234567890",
				},
			},
			expectedParams: &stripe.CustomerParams{
				Phone: stripe.String("1234567890"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := GetUpdatedFields(tt.props, tt.stripeCustomer)
			assert.Equal(t, tt.expectedParams, params)
		})
	}
}
