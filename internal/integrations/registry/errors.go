package registry

import "errors"

var (
	// ErrNoProviderSpecs indicates no provider specifications were supplied during registry creation
	ErrNoProviderSpecs = errors.New("integrations/registry: no provider specs supplied")
)
