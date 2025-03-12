package entitlements

import (
	"errors"
)

var (
	// ErrFoundMultipleCustomers is returned when multiple customers are found
	ErrFoundMultipleCustomers = errors.New("found multiple customers with the same name")
	// ErrCustomerNotFound is returned when a customer is not found
	ErrCustomerNotFound = errors.New("customer not found")
	// ErrCustomerLookupFailed is returned when a customer lookup fails
	ErrCustomerLookupFailed = errors.New("failed to lookup customer")
	// ErrCustomerIDRequired is returned when a customer ID is required
	ErrCustomerIDRequired = errors.New("customer ID is required")
	// ErrMissingAPIKey is returned when the API key is missing
	ErrMissingAPIKey = errors.New("missing API key")
)
