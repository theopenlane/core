package entitlements

import "github.com/stripe/stripe-go/v84"

// CustomerCreateOption allows customizing CustomerCreateParams
type CustomerCreateOption func(params *stripe.CustomerCreateParams)

// WithCustomerEmail sets the email for the customer
func WithCustomerEmail(email string) CustomerCreateOption {
	return func(params *stripe.CustomerCreateParams) {
		params.Email = stripe.String(email)
	}
}

// WithCustomerName sets the name for the customer
func WithCustomerName(name string) CustomerCreateOption {
	return func(params *stripe.CustomerCreateParams) {
		params.Name = stripe.String(name)
	}
}

// WithCustomerMetadata sets metadata for the customer
func WithCustomerMetadata(metadata map[string]string) CustomerCreateOption {
	return func(params *stripe.CustomerCreateParams) {
		params.Metadata = metadata
	}
}

// WithCustomerPhone sets the phone for the customer
func WithCustomerPhone(phone string) CustomerCreateOption {
	return func(params *stripe.CustomerCreateParams) {
		params.Phone = stripe.String(phone)
	}
}

// WithCustomerAddress sets the address for the customer
func WithCustomerAddress(addr *stripe.AddressParams) CustomerCreateOption {
	return func(params *stripe.CustomerCreateParams) {
		params.Address = addr
	}
}

// CreateCustomerWithOptions creates a customer with functional options
func (sc *StripeClient) CreateCustomerWithOptions(baseParams *stripe.CustomerCreateParams, opts ...CustomerCreateOption) *stripe.CustomerCreateParams {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return params
}

// CustomerUpdateOption allows customizing CustomerUpdateParams
type CustomerUpdateOption func(params *stripe.CustomerUpdateParams)

// WithUpdateCustomerEmail sets the email for the customer update
func WithUpdateCustomerEmail(email string) CustomerUpdateOption {
	return func(params *stripe.CustomerUpdateParams) {
		params.Email = stripe.String(email)
	}
}

// WithUpdateCustomerMetadata sets metadata for the customer update
func WithUpdateCustomerMetadata(metadata map[string]string) CustomerUpdateOption {
	return func(params *stripe.CustomerUpdateParams) {
		params.Metadata = metadata
	}
}

// WithUpdateCustomerPhone sets the phone for the customer update
func WithUpdateCustomerPhone(phone string) CustomerUpdateOption {
	return func(params *stripe.CustomerUpdateParams) {
		params.Phone = stripe.String(phone)
	}
}

// WithUpdateCustomerAddress sets the address for the customer update
func WithUpdateCustomerAddress(addr *stripe.AddressParams) CustomerUpdateOption {
	return func(params *stripe.CustomerUpdateParams) {
		params.Address = addr
	}
}

// UpdateCustomerWithOptions creates update params with functional options
func (sc *StripeClient) UpdateCustomerWithOptions(baseParams *stripe.CustomerUpdateParams, opts ...CustomerUpdateOption) *stripe.CustomerUpdateParams {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return params
}

// --- Example Usage ---
// params := &stripe.CustomerCreateParams{}
// params = sc.CreateCustomerWithOptions(params, WithCustomerEmail("foo@bar.com"), WithCustomerName("Acme"))
// customer, err := sc.Client.V1Customers.Create(ctx, params)
//
// updateParams := &stripe.CustomerUpdateParams{}
// updateParams = sc.UpdateCustomerWithOptions(updateParams, WithUpdateCustomerEmail("new@bar.com"))
// updated, err := sc.UpdateCustomer(ctx, customerID, updateParams)
