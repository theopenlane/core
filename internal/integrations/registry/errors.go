package registry

import "errors"

var (
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
	// ErrProviderNotFound indicates the requested provider is not registered
	ErrProviderNotFound = errors.New("integrations/registry: provider not found")
	// ErrOperationCriteriaRequired indicates operation name or kind is required
	ErrOperationCriteriaRequired = errors.New("integrations/registry: operation criteria required")
	// ErrOperationNotRegistered indicates no matching operation descriptor exists for the provider
	ErrOperationNotRegistered = errors.New("integrations/registry: operation not registered")
	// ErrOperationKindMismatch indicates operation name and kind constraints conflict
	ErrOperationKindMismatch = errors.New("integrations/registry: operation kind mismatch")
	// ErrOperationDescriptorAmbiguous indicates operation selection matched multiple descriptors
	ErrOperationDescriptorAmbiguous = errors.New("integrations/registry: operation descriptor ambiguous")
)
