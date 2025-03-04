package entitlements

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
)

// CreateCustomer creates a customer leveraging the openlane organization ID
// as the organization name, and the email provided as the billing email
// we assume that the billing email will be changed, so lookups are performed by the organization ID
func (sc *StripeClient) CreateCustomer(c *OrganizationCustomer) (*stripe.Customer, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	customer, err := sc.Client.Customers.New(&stripe.CustomerParams{
		Email: &c.Email,
		Name:  &c.OrganizationID,
		Phone: &c.Phone,
		Address: &stripe.AddressParams{
			Line1:      c.Line1,
			City:       c.City,
			State:      c.State,
			PostalCode: c.PostalCode,
			Country:    c.Country,
		},
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
// leverage QueryBuilder to construct more complex queries, otherwise see:
// https://docs.stripe.com/search#search-query-language
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

// FindOrCreateCustomer attempts to lookup a customer by the organization ID which is set in both the
// name field attribute as well as in the object metadata field
func (sc *StripeClient) FindOrCreateCustomer(ctx context.Context, o *OrganizationCustomer) (*OrganizationCustomer, error) {
	log.Debug().Str("organization_id", o.OrganizationID).Msg("searching for customer")

	customers, err := sc.SearchCustomers(ctx, fmt.Sprintf("name: '%s'", o.OrganizationID))
	if err != nil {
		return nil, err
	}

	switch len(customers) {
	case 0:
		log.Debug().Str("organization_id", o.OrganizationID).Msg("no customer found, creating")

		customer, err := sc.CreateCustomer(o)
		if err != nil {
			return nil, err
		}

		log.Debug().Str("customer_id", customer.ID).Msg("customer created")

		if o.PersonalOrg {
			psubs, err := sc.CreatePersonalOrgFreeTierSubs(customer.ID)
			if err != nil {
				return nil, err
			}

			o.StripeCustomerID = customer.ID
			o.Subscription = *psubs

			newOrg, err := sc.retrieveFeatureLists(o)
			if err != nil {
				return nil, err
			}

			return newOrg, nil
		}

		subs, err := sc.CreateTrialSubscription(customer.ID)
		if err != nil {
			return nil, err
		}

		log.Debug().Str("customer_id", customer.ID).Str("subscription_id", subs.ID).Msg("trial subscription created")

		o.StripeCustomerID = customer.ID
		o.Subscription = *subs

		newOrg, err := sc.retrieveFeatureLists(o)
		if err != nil {
			return nil, err
		}

		return newOrg, nil
	case 1:
		log.Debug().Str("organization_id", o.OrganizationID).Str("customer_id", customers[0].ID).Msg("customer found, not creating a new one")
		o.StripeCustomerID = customers[0].ID

		feats, featNames, err := sc.retrieveActiveEntitlements(customers[0].ID)
		if err != nil {
			return nil, err
		}

		if len(feats) > 0 {
			// if we have feats, lets update the customer object in case things have changed or we missed something
			log.Debug().Str("organization_id", o.OrganizationID).Str("customer_id", customers[0].ID).Strs("features", feats).Msg("found features for customer")
			o.Features = feats
		}

		if len(featNames) > 0 {
			log.Debug().Str("organization_id", o.OrganizationID).Str("customer_id", customers[0].ID).Strs("features_names", featNames).Msg("found features for customer")

			o.FeatureNames = featNames
		}

		return o, nil
	default:
		log.Error().Err(ErrFoundMultipleCustomers).Str("organization_id", o.OrganizationID).Interface("customers", customers).Msg("found multiple customers, skipping all updates")

		return nil, ErrFoundMultipleCustomers
	}
}

func (sc *StripeClient) retrieveFeatureLists(o *OrganizationCustomer) (*OrganizationCustomer, error) {
	var feats, featNames []string

	const maxRetries = 5

	for i := range maxRetries {
		var err error

		feats, featNames, err = sc.retrieveActiveEntitlements(o.StripeCustomerID)
		if err != nil {
			return nil, err
		}

		// if we have features, break out of the loop
		if len(feats) > 0 {
			log.Debug().Str("organization_id", o.OrganizationID).Str("customer_id", o.StripeCustomerID).Msg("features found for customer")

			break
		}

		log.Debug().Str("organization_id", o.OrganizationID).Str("customer_id", o.StripeCustomerID).Msg("no features found for customer, retrying")

		time.Sleep(time.Duration(i+1) * time.Second) // backoff retry
	}

	if len(feats) == 0 {
		log.Warn().Str("customer_id", o.StripeCustomerID).Msg("no features found for customer")
	}

	log.Debug().Str("organization_id", o.OrganizationID).Strs("features", feats).Str("customer_id", o.StripeCustomerID).Msg("found features for customer")

	o.Features = feats
	o.FeatureNames = featNames

	return o, nil
}

// GetCustomerByStripeID gets a customer by ID
func (sc *StripeClient) GetCustomerByStripeID(ctx context.Context, customerID string) (*stripe.Customer, error) {
	if customerID == "" {
		return nil, ErrCustomerIDRequired
	}

	customer, err := sc.Client.Customers.Get(customerID, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{stripe.String("tax"), stripe.String("subscriptions")},
		},
	})

	// if the customer is not found, return a specific error, otherwise surface the failed lookup
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
