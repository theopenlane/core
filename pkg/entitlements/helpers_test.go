package entitlements_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v84"
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

func TestWriteYAMLHelpers(t *testing.T) {
	plans := []entitlements.Product{{Name: "Test"}}

	dir := t.TempDir()
	plansFile := filepath.Join(dir, "plans.yaml")

	assert.NoError(t, entitlements.WritePlansToYAML(plans, plansFile))

	pData, err := os.ReadFile(plansFile)
	assert.NoError(t, err)
	assert.Contains(t, string(pData), "Test")

}
