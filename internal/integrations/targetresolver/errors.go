package targetresolver

import "errors"

var (
	// ErrResolverSourceRequired indicates an integration source dependency is required
	ErrResolverSourceRequired = errors.New("targetresolver: source required")
	// ErrResolverDBClientRequired indicates a database client dependency is required
	ErrResolverDBClientRequired = errors.New("targetresolver: db client required")
	// ErrResolverOwnerIDRequired indicates owner id is required to resolve targets
	ErrResolverOwnerIDRequired = errors.New("targetresolver: owner id required")
	// ErrResolverIntegrationIDRequired indicates integration id cannot be empty when specified
	ErrResolverIntegrationIDRequired = errors.New("targetresolver: integration id required")
	// ErrResolverProviderRequired indicates provider is required when integration id is not specified
	ErrResolverProviderRequired = errors.New("targetresolver: provider required")
	// ErrResolverProviderUnknown indicates the resolved provider cannot be mapped from integration kind
	ErrResolverProviderUnknown = errors.New("targetresolver: provider unknown")
	// ErrResolverProviderMismatch indicates provider input conflicts with resolved integration provider
	ErrResolverProviderMismatch = errors.New("targetresolver: provider mismatch")
	// ErrResolverIntegrationNotFound indicates no installed integration matched the criteria
	ErrResolverIntegrationNotFound = errors.New("targetresolver: integration not found")
	// ErrResolverIntegrationAmbiguous indicates multiple installed integrations matched the criteria
	ErrResolverIntegrationAmbiguous = errors.New("targetresolver: integration ambiguous")
)
