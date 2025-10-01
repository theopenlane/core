package entitlements

import (
	"errors"
)

var (
	// ErrFoundMultipleCustomers is returned when multiple customers are found
	ErrFoundMultipleCustomers = errors.New("found multiple customers with the same name")
	// ErrMultipleSubscriptions = errors.New("multiple subscriptions found")
	ErrMultipleSubscriptions = errors.New("multiple subscriptions found")
	// ErrNoSubscriptions is returned when no subscriptions are found
	ErrNoSubscriptions = errors.New("no subscriptions found")
	// ErrCustomerNotFound is returned when a customer is not found
	ErrCustomerNotFound = errors.New("customer not found")
	// ErrCustomerLookupFailed is returned when a customer lookup fails
	ErrCustomerLookupFailed = errors.New("failed to lookup customer")
	// ErrCustomerSearchFailed is returned when a customer search fails
	ErrCustomerSearchFailed = errors.New("failed to search for customer")
	// ErrProductListFailed is returned when a product list fails
	ErrProductListFailed = errors.New("failed to list products")
	// ErrCustomerIDRequired is returned when a customer ID is required
	ErrCustomerIDRequired = errors.New("customer ID is required")
	// ErrMissingAPIKey is returned when the API key is missing
	ErrMissingAPIKey = errors.New("missing API key")
	// ErrNoSubscriptionItems is returned when no subscription items are found
	ErrNoSubscriptionItems = errors.New("no subscription items found to create subscription")
)
