package entitlements

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v83"
)

// CreateCustomer creates a customer leveraging the openlane organization ID
// as the organization name, and the email provided as the billing email
// we assume that the billing email will be changed, so lookups are performed by the organization ID
func (sc *StripeClient) CreateCustomer(ctx context.Context, c *OrganizationCustomer) (*stripe.Customer, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	// Build metadata with default fields
	metadata := map[string]string{
		"organization_id":              c.OrganizationID,
		"organization_settings_id":     c.OrganizationSettingsID,
		"organization_name":            c.OrganizationName,
		"organization_subscription_id": c.OrganizationSubscriptionID,
		"personal_org":                 "",
	}

	// Merge any additional metadata from the struct
	maps.Copy(metadata, c.Metadata)

	params := sc.CreateCustomerWithOptions(
		&stripe.CustomerCreateParams{},
		WithCustomerEmail(c.Email),
		WithCustomerName(c.OrganizationID),
		WithCustomerPhone(c.Phone),
		WithCustomerAddress(&stripe.AddressParams{
			Line1:      c.Line1,
			City:       c.City,
			State:      c.State,
			PostalCode: c.PostalCode,
			Country:    c.Country,
		}),
		WithCustomerMetadata(metadata),
	)

	start := time.Now()
	customer, err := sc.Client.V1Customers.Create(ctx, params)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	stripeRequestCounter.WithLabelValues("customers", status).Inc()
	stripeRequestDuration.WithLabelValues("customers", status).Observe(duration)

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
		Expand: []*string{stripe.String("data.tax"), stripe.String("data.subscriptions")},
		SearchParams: stripe.SearchParams{
			Query:   query,
			Context: ctx,
		},
	}

	result := sc.Client.V1Customers.Search(ctx, params)

	for customer, err := range result {
		if err != nil {
			log.Err(err).Msg("failed to search customers")

			return nil, err
		}

		customers = append(customers, customer)
	}

	return customers, nil
}

// CreateCustomerAndSubscription handles the case where no customer exists by creating
// both the customer and their initial subscription
func (sc *StripeClient) CreateCustomerAndSubscription(ctx context.Context, o *OrganizationCustomer) error {
	customer, err := sc.CreateCustomer(ctx, o)
	if err != nil {
		return err
	}

	o.StripeCustomerID = customer.ID
	o.Metadata = customer.Metadata

	subscription, err := sc.CreateSubscriptionWithPrices(ctx, customer, o)
	if err != nil {
		return err
	}

	o.StripeSubscriptionID = subscription.ID
	o.Subscription = *subscription
	o.StripeSubscriptionScheduleID = subscription.StripeSubscriptionScheduleID

	log.Debug().Str("customer_id", customer.ID).Str("subscription_id", subscription.ID).Msg("subscription created")

	_, err = sc.Client.V1Customers.Update(ctx, customer.ID, sc.UpdateCustomerWithOptions(
		&stripe.CustomerUpdateParams{}, WithUpdateCustomerMetadata(map[string]string{"subscription_schedule_id": subscription.StripeSubscriptionScheduleID})))
	if err != nil {
		log.Err(err).Msg("Failed to update customer with subscription schedule ID")

		return err
	}

	return nil
}

// FindOrCreateCustomer attempts to lookup a customer by the organization ID which is set in both the
// name field attribute as well as in the object metadata field
func (sc *StripeClient) FindOrCreateCustomer(ctx context.Context, o *OrganizationCustomer) error {
	customers, err := sc.SearchCustomers(ctx, fmt.Sprintf("name: '%s'", o.OrganizationID))
	if err != nil {
		return err
	}

	switch len(customers) {
	case 0:
		return sc.CreateCustomerAndSubscription(ctx, o)
	case 1:
		if customers[0].Subscriptions == nil || len(customers[0].Subscriptions.Data) == 0 {
			return ErrNoSubscriptions
		}

		log.Debug().Str("organization_id", o.OrganizationID).Str("customer_id", customers[0].ID).Msg("found existing customer for organization")

		o.StripeCustomerID = customers[0].ID
		o.StripeSubscriptionID = customers[0].Subscriptions.Data[0].ID

		return nil
	default:
		log.Error().Err(ErrFoundMultipleCustomers).Str("organization_id", o.OrganizationID).Interface("customers", customers).Msg("found multiple customers, skipping all updates")
		return ErrFoundMultipleCustomers
	}
}

// GetCustomerByStripeID gets a customer by ID
func (sc *StripeClient) GetCustomerByStripeID(ctx context.Context, customerID string) (*stripe.Customer, error) {
	if customerID == "" {
		return nil, ErrCustomerIDRequired
	}

	start := time.Now()
	customer, err := sc.Client.V1Customers.Retrieve(ctx, customerID, &stripe.CustomerRetrieveParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{stripe.String("tax"), stripe.String("subscriptions")},
		},
	})
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	stripeRequestCounter.WithLabelValues("customers", status).Inc()
	stripeRequestDuration.WithLabelValues("customers", status).Observe(duration)

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

// UpdateCustomer updates a customer in stripe with the provided params and ID
func (sc *StripeClient) UpdateCustomer(ctx context.Context, id string, params *stripe.CustomerUpdateParams) (*stripe.Customer, error) {
	if id == "" || params == nil {
		return nil, ErrCustomerIDRequired
	}

	cust, err := sc.Client.V1Customers.Update(ctx, id, params)
	if err != nil {
		return nil, err
	}

	return cust, nil
}

// DeleteCustomer deletes a customer by ID from stripe
func (sc *StripeClient) DeleteCustomer(ctx context.Context, id string) error {
	_, err := sc.Client.V1Customers.Delete(ctx, id, nil)
	if err != nil {
		return err
	}

	return nil
}

// FindAndDeactivateCustomerSubscription finds a customer by the organization ID and deactivates their subscription
// this is used when an organization is deleted - we retain the customer record and keep a referenced to the deactivated subscription
// we do not delete the customer record in stripe for record / references
// we also do not delete the subscription record in stripe for record / references
// a cancelled active subscription will set to cancel at period end, a trialing subscription will be set to end immediately
func (sc *StripeClient) FindAndDeactivateCustomerSubscription(ctx context.Context, customerID string) error {
	customer, err := sc.GetCustomerByStripeID(ctx, customerID)
	if err != nil {
		return err
	}

	for _, sub := range customer.Subscriptions.Data {
		// skip subscriptions that are already inactive
		if sub.Status == stripe.SubscriptionStatusCanceled || sub.Status == stripe.SubscriptionStatusIncompleteExpired {
			log.Debug().Str("subscription_id", sub.ID).Msg("subscription already inactive, skipping")
			return nil
		}

		// when an organization is deleted, the subscription should be cancelled immediately, instead of at period end
		_, err := sc.Client.V1Subscriptions.Cancel(ctx, sub.ID,
			&stripe.SubscriptionCancelParams{
				CancellationDetails: &stripe.SubscriptionCancelCancellationDetailsParams{
					Comment: lo.ToPtr("system: organization was deleted - cancelling subscription"),
				},
			})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetOrganizationIDFromMetadata gets the organization ID from the metadata
// if it exists, otherwise returns an empty string
func GetOrganizationIDFromMetadata(metadata map[string]string) string {
	orgID, exists := metadata["organization_id"]
	if exists {
		return orgID
	}

	return ""
}

// GetOrganizationNameFromMetadata gets the organization name from the metadata
// if it exists, otherwise returns an empty string
func GetOrganizationNameFromMetadata(metadata map[string]string) string {
	orgName, exists := metadata["organization_name"]
	if exists {
		return orgName
	}

	return ""
}

// GetOrganizationSettingsIDFromMetadata gets the organization settings ID from the metadata
// if it exists, otherwise returns an empty string
func GetOrganizationSettingsIDFromMetadata(metadata map[string]string) string {
	orgSettingsID, exists := metadata["organization_settings_id"]
	if exists {
		return orgSettingsID
	}

	return ""
}

// GetOrganizationSubscriptionIDFromMetadata gets the organization subscription ID from the metadata
// if it exists, otherwise returns an empty string
func GetOrganizationSubscriptionIDFromMetadata(metadata map[string]string) string {
	orgSubID, exists := metadata["organization_subscription_id"]
	if exists {
		return orgSubID
	}

	return ""
}

// MapStripeCustomer maps a stripe customer to an organization customer
// this is used to convert the stripe customer object to our internal customer object
// we use the metadata to store the organization ID, settings ID, and subscription ID
func MapStripeCustomer(c *stripe.Customer) *OrganizationCustomer {
	if c == nil {
		return nil
	}

	paymentAdded := c.Sources != nil && c.Sources.Data != nil

	return &OrganizationCustomer{
		StripeCustomerID:           c.ID,
		OrganizationID:             GetOrganizationIDFromMetadata(c.Metadata),
		OrganizationSettingsID:     GetOrganizationSettingsIDFromMetadata(c.Metadata),
		OrganizationSubscriptionID: GetOrganizationSubscriptionIDFromMetadata(c.Metadata),
		OrganizationName:           GetOrganizationNameFromMetadata(c.Metadata),
		PaymentMethodAdded:         paymentAdded,
	}
}
