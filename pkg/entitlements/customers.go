package entitlements

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
)

func (sc *StripeClient) CreateCustomer(c *OrganizationCustomer) (*stripe.Customer, error) {
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

// FindorCreateCustomer finds or creates a customer
func (sc *StripeClient) FindorCreateCustomer(ctx context.Context, o *OrganizationCustomer) (*OrganizationCustomer, error) {
	//	qb := NewQueryBuilder(WithKeys(map[string]string{"organization_id": o.OrganizationID}))
	withName := fmt.Sprintf("name: '%s'", o.OrganizationID)

	customers, err := sc.SearchCustomers(ctx, withName)
	if err != nil {
		return nil, err
	}

	switch len(customers) {
	case 0:
		customer, err := sc.CreateCustomer(o)
		if err != nil {
			return nil, err
		}

		o.StripeCustomerID = customer.ID

		return o, nil
	case 1:
		o.StripeCustomerID = customers[0].ID

		return o, nil
	default:
		return nil, ErrFoundMultipleCustomers
	}

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
			if stripeErr.Code == stripe.ErrorCodeMissing {
				return nil, ErrCustomerNotFound
			}
		}

		return nil, ErrCustomerLookupFailed
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
