package types //nolint:revive

import "errors"

var (
	// ErrProviderTypeRequired indicates a builder or option pipeline did not
	// receive a provider identifier. Every credential payload must be scoped to
	// a provider to avoid ambiguous persistence/lookups.
	ErrProviderTypeRequired = errors.New("integrations/types: provider type required")

	// ErrCredentialSetRequired signals that no credential set was provided while
	// building a payload.
	ErrCredentialSetRequired = errors.New("integrations/types: credential set required")
)
