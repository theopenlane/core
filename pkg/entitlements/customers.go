package entitlements

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
)

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

// SearchCustomers searches for customers with a structured stripe query as input
func (sc *StripeClient) SearchCustomers(ctx context.Context, query string) (customers []*stripe.Customer, err error) {
	params := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query:   query,
			Expand:  []*string{stripe.String("data.tax"), stripe.String("data.subscriptions")},
			Context: ctx,
		},
	}

	iter := sc.Client.Customers.Search(params)

	for iter.Next() {
		customers = append(customers, iter.Customer())
	}

	if iter.Err() != nil {
		log.Err(iter.Err()).Msg("failed to find customers")
		return nil, iter.Err()
	}

	return customers, nil
}

// GetCustomerByStripeID gets a customer by ID
func (sc *StripeClient) GetCustomerByStripeID(ctx context.Context, customerID string) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.Get(customerID, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{stripe.String("tax"), stripe.String("subscriptions")},
		},
	})

	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeMissing:
				return nil, fmt.Errorf("customer %s does not exist in stripe", customerID)
			}
		}

		return nil, fmt.Errorf("failed to get customer by customer ID %s", customerID)
	}

	return customer, nil
}

// GetCustomerByOrganizationID gets a customer by organization ID
func (sc *StripeClient) GetCustomerByOrganizationID(ctx context.Context, organizationID string) (*stripe.Customer, error) {
	customers, err := sc.SearchCustomers(ctx, fmt.Sprintf("metadata.organization_id:%s", organizationID))
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, fmt.Errorf("customer with organization ID %s not found", organizationID)
	}

	if len(customers) > 1 {
		return nil, fmt.Errorf("multiple customers found with organization ID %s", organizationID)
	}

	return customers[0], nil
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
