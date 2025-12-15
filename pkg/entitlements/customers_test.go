package entitlements_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/core/pkg/entitlements"
)

func TestMapStripeCustomer(t *testing.T) {
	tests := []struct {
		name           string
		stripeCustomer *stripe.Customer
		expected       *entitlements.OrganizationCustomer
	}{
		{
			name: "All fields",
			stripeCustomer: &stripe.Customer{
				ID: "cus_123",
				Metadata: map[string]string{
					"organization_id":              "org_123",
					"organization_settings_id":     "settings_123",
					"organization_subscription_id": "sub_123",
				},
				Sources: &stripe.PaymentSourceList{
					Data: []*stripe.PaymentSource{
						{
							ID: "src_123",
						},
					},
				}},
			expected: &entitlements.OrganizationCustomer{
				StripeCustomerID:           "cus_123",
				OrganizationID:             "org_123",
				OrganizationSettingsID:     "settings_123",
				OrganizationSubscriptionID: "sub_123",
				PaymentMethodAdded:         true,
			},
		},
		{
			name: "no payment method",
			stripeCustomer: &stripe.Customer{
				ID: "cus_123",
				Metadata: map[string]string{
					"organization_id":              "org_123",
					"organization_settings_id":     "settings_123",
					"organization_subscription_id": "sub_123",
				},
			},
			expected: &entitlements.OrganizationCustomer{
				StripeCustomerID:           "cus_123",
				OrganizationID:             "org_123",
				OrganizationSettingsID:     "settings_123",
				OrganizationSubscriptionID: "sub_123",
				PaymentMethodAdded:         false,
			},
		},
		{
			name: "all metadata fields",
			stripeCustomer: &stripe.Customer{
				ID: "cus_123",
			},
			expected: &entitlements.OrganizationCustomer{
				StripeCustomerID:           "cus_123",
				OrganizationID:             "",
				OrganizationSettingsID:     "",
				OrganizationSubscriptionID: "",
				PaymentMethodAdded:         false,
			},
		},
		{
			name:           "nil customer",
			stripeCustomer: nil,
			expected:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer := entitlements.MapStripeCustomer(tt.stripeCustomer)

			assert.Equal(t, tt.expected, customer)
		})
	}
}
