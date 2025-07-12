package catalog

import "errors"

var (
	// ErrCatalogValidationFailed is returned when the catalog fails validation
	ErrCatalogValidationFailed = errors.New("catalog validation failed")
	// ErrProductMissingFeature is returned when a product is missing a required feature
	ErrProductMissingFeature = errors.New("product missing required feature")
	// ErrYamlToJSONConversion = errors.New("failed to convert YAML to JSON for catalog validation")
	ErrYamlToJSONConversion = errors.New("failed to convert YAML to JSON for catalog validation")
	// ErrMatchingPriceNotFound = errors.New("matching price not found for feature")
	ErrMatchingPriceNotFound = errors.New("matching price not found for feature")
	// ErrFailedToCreateProduct = errors.New("failed to create product in Stripe")
	ErrFailedToCreateProduct = errors.New("failed to create product in Stripe")
	// ErrFailedToCreatePrice = errors.New("failed to create price in Stripe")
	ErrFailedToCreatePrice = errors.New("failed to create price in Stripe")
	// ErrContextandClientRequired = errors.New("context and client are required for catalog operations"
	ErrContextandClientRequired = errors.New("context and client are required for catalog operations")
	// ErrLookupKeyConflict indicates a price lookup key already exists in Stripe
	ErrLookupKeyConflict = errors.New("lookup key conflict")
)
