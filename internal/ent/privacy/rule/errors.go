package rule

import "errors"

var (
	ErrRequiredScopeNotSet = errors.New("the provided token does not have the required scopes for the request")
	// ErrFeaturesNotEnabled is returned when a mutation targets a schema whose modules the organization does not have
	ErrFeaturesNotEnabled = errors.New("features are not enabled")
)
