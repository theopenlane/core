package entitlements_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stripe/stripe-go/v86"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
			assert.Check(t, is.DeepEqual(tt.expectedParams.Email, params.Email))
			assert.Check(t, is.DeepEqual(tt.expectedParams.Phone, params.Phone))
			assert.Check(t, is.DeepEqual(tt.expectedParams.Address, params.Address))
		})
	}
}

func TestWriteYAMLHelpers(t *testing.T) {
	plans := []entitlements.Product{{Name: "Test"}}

	dir := t.TempDir()
	plansFile := filepath.Join(dir, "plans.yaml")

	assert.Check(t, entitlements.WritePlansToYAML(plans, plansFile))

	pData, err := os.ReadFile(plansFile)
	assert.Check(t, err)
	assert.Check(t, is.Contains(string(pData), "Test"))

}
