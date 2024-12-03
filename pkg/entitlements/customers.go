package entitlements

import "github.com/stripe/stripe-go/v81"

// CreateCustomer creates a new customer
func (sc *StripeClient) CreateCustomer(email string) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.New(&stripe.CustomerParams{Email: &email})
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (sc *StripeClient) CreateNewCustomer(c *OrganizationCustomer) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.New(&stripe.CustomerParams{
		Email: &c.BillingEmail,
		Name:  &c.OrganizationID,
		Phone: &c.BillingPhone,
		Metadata: map[string]string{
			"organization_id":          c.OrganizationID,
			"organization_settings_id": c.OrganizationSettingsID,
			"organization_name":        c.OrganizationName,
		},
	})
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomerByID gets a customer by ID
func (sc *StripeClient) GetCustomerByStripeID(id string) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// UpdateCustomer updates a customer
func (sc *StripeClient) UpdateCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.Update(id, params)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// DeleteCustomer deletes a customer
func (sc *StripeClient) DeleteCustomer(id string) error {
	_, err := sc.Client.Customers.Del(id, nil)
	if err != nil {
		return err
	}

	return nil
}
