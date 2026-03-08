package types //nolint:revive

import "errors"

var (
	// ErrProviderTypeRequired indicates a required provider identifier was missing.
	ErrProviderTypeRequired = errors.New("integrations/types: provider type required")

	// ErrCredentialSetRequired signals that no credential set was provided where required.
	ErrCredentialSetRequired = errors.New("integrations/types: credential set required")

	// ErrCredentialMetadataInvalid indicates provider metadata could not be marshaled
	// or unmarshaled into canonical JSON/map form.
	ErrCredentialMetadataInvalid = errors.New("integrations/types: credential metadata invalid")
)
