package entitlements

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v84"
	"gopkg.in/yaml.v3"
)

// GetUpdatedFields checks for updates to billing information in the properties and returns a stripe.CustomerParams object with the updated information
// and a boolean indicating whether there are updates
func GetUpdatedFields(props map[string]any, orgCustomer *OrganizationCustomer) (params *stripe.CustomerUpdateParams) {
	// Initialize params to avoid nil dereference
	params = &stripe.CustomerUpdateParams{}

	// if its in the properties, it has been updated
	// use the current value from orgCustomer
	if _, exists := props["billing_email"]; exists {
		WithUpdateCustomerEmail(orgCustomer.Email)(params)
	}

	if _, exists := props["billing_phone"]; exists {
		WithUpdateCustomerPhone(orgCustomer.Phone)(params)
	}

	if _, exists := props["billing_address"]; exists {
		WithUpdateCustomerAddress(&stripe.AddressParams{
			Line1:      orgCustomer.Line1,
			Line2:      orgCustomer.Line2,
			City:       orgCustomer.City,
			State:      orgCustomer.State,
			PostalCode: orgCustomer.PostalCode,
			Country:    orgCustomer.Country,
		})(params)
	}

	return params
}

// WritePlansToYAML writes the []Product information into a YAML file.
func WritePlansToYAML(product []Product, filename string) error {
	// Marshal the []Product information into YAML
	data, err := yaml.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal plans to YAML: %w", err)
	}

	// Write the YAML data to a file
	err = os.WriteFile(filename, data, 0600) // nolint:mnd
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}

// Seq2IsEmpty checks if a stripe.Seq2 is empty.
//
// Parameters:
//   - seq: a stripe.Seq2 iterator of type K and error type V.
//
// Returns:
//   - bool: true if the sequence is empty, false otherwise.
func Seq2IsEmpty[K any, V error](seq stripe.Seq2[K, V]) bool {
	isEmpty := true

	seq(func(_ K, _ V) bool {
		isEmpty = false
		return false // stop after first element
	})

	return isEmpty
}
