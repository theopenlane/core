package entitlements

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v82"
	"gopkg.in/yaml.v3"
)

// GetUpdatedFields checks for updates to billing information in the properties and returns a stripe.CustomerParams object with the updated information
// and a boolean indicating whether there are updates
func GetUpdatedFields(props map[string]interface{}, orgCustomer *OrganizationCustomer) (params *stripe.CustomerUpdateParams) {
	// Initialize params to avoid nil dereference
	params = &stripe.CustomerUpdateParams{}

	// if its in the properties, it has been updated
	// use the current value from orgCustomer
	_, exists := props["billing_email"]
	if exists {
		params.Email = &orgCustomer.Email
	}

	_, exists = props["billing_phone"]
	if exists {
		params.Phone = &orgCustomer.Phone
	}

	_, exists = props["billing_address"]
	if exists {
		params.Address = &stripe.AddressParams{
			Line1:      orgCustomer.Line1,
			Line2:      orgCustomer.Line2,
			City:       orgCustomer.City,
			State:      orgCustomer.State,
			PostalCode: orgCustomer.PostalCode,
			Country:    orgCustomer.Country,
		}
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

// WriteTuplesToYaml writes the []TupleStruct information into a YAML file
func WriteTuplesToYaml(tuples []TupleStruct, filename string) error {
	// Marshal the []TupleStruct information into YAML
	data, err := yaml.Marshal(tuples)
	if err != nil {
		return fmt.Errorf("failed to marshal tuples to YAML: %w", err)
	}

	// Write the YAML data to a file
	err = os.WriteFile(filename, data, 0600) // nolint:mnd
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}

// seq2IsEmpty checks if a stripe.Seq2 is empty
func seq2IsEmpty[K any, V error](seq stripe.Seq2[K, V]) bool {
	seq(func(_ K, _ V) bool {
		return false
	})

	return true
}
