package registry

import "errors"

var (
	// ErrNoProviderSpecs indicates no provider specifications were supplied during registry creation
	ErrNoProviderSpecs = errors.New("integrations/registry: no provider specs supplied")
	// ErrRegistryNil indicates the registry is nil
	ErrRegistryNil = errors.New("integrations/registry: registry is nil")
	// ErrProviderTypeRequired indicates the provider type is missing
	ErrProviderTypeRequired = errors.New("integrations/registry: provider type required")
	// ErrBuilderMismatch indicates the builder is missing or mismatched
	ErrBuilderMismatch = errors.New("integrations/registry: builder missing or type mismatch")
	// ErrProviderBuildFailed indicates provider construction failed
	ErrProviderBuildFailed = errors.New("integrations/registry: build provider failed")
	// ErrProviderNil indicates a provider build returned a nil instance
	ErrProviderNil = errors.New("integrations/registry: provider is nil")
)
