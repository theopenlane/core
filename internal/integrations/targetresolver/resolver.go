package targetresolver

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
)

// Resolver resolves installed integrations and provider operation descriptors from explicit criteria
type Resolver struct {
	source   IntegrationSource
	registry OperationRegistry
}

// EntSource resolves installed integrations from ent persistence
type EntSource struct {
	client *entgen.Client
}

// NewResolver constructs a target resolver with required dependencies
func NewResolver(source IntegrationSource, registry OperationRegistry) (*Resolver, error) {
	if source == nil {
		return nil, ErrResolverSourceRequired
	}
	if registry == nil {
		return nil, ErrResolverRegistryRequired
	}

	return &Resolver{
		source:   source,
		registry: registry,
	}, nil
}

// NewEntSource constructs an ent-backed integration source
func NewEntSource(client *entgen.Client) (*EntSource, error) {
	if client == nil {
		return nil, ErrResolverDBClientRequired
	}

	return &EntSource{client: client}, nil
}

// IntegrationByID returns one installed integration record by owner and id
func (s *EntSource) IntegrationByID(ctx context.Context, ownerID string, integrationID string) (*entgen.Integration, error) {
	record, err := s.client.Integration.Query().
		Where(
			integration.IDEQ(integrationID),
			integration.OwnerIDEQ(ownerID),
		).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return record, nil
}

// IntegrationsByProvider returns installed integration records by owner and provider
func (s *EntSource) IntegrationsByProvider(ctx context.Context, ownerID string, provider types.ProviderType) ([]*entgen.Integration, error) {
	return s.client.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.KindEQ(string(provider)),
		).
		All(ctx)
}

// Resolve resolves an installed integration and operation descriptor from explicit criteria
func (r *Resolver) Resolve(ctx context.Context, criteria ResolveCriteria) (ResolveResult, error) {
	if err := validateResolveCriteria(criteria); err != nil {
		return ResolveResult{}, err
	}

	provider, integrationRecord, err := r.resolveProviderAndIntegration(ctx, criteria)
	if err != nil {
		return ResolveResult{}, err
	}

	operationDescriptor, err := r.resolveOperationDescriptor(provider, criteria)
	if err != nil {
		return ResolveResult{}, err
	}

	return ResolveResult{
		Integration: integrationRecord,
		Provider:    provider,
		Operation:   operationDescriptor,
	}, nil
}

// validateResolveCriteria validates required fields and explicit constraints
func validateResolveCriteria(criteria ResolveCriteria) error {
	if criteria.OwnerID == "" {
		return ErrResolverOwnerIDRequired
	}
	if !criteria.OperationName.IsPresent() && !criteria.OperationKind.IsPresent() {
		return ErrResolverOperationCriteriaRequired
	}

	if integrationID, ok := criteria.IntegrationID.Get(); ok && integrationID == "" {
		return ErrResolverIntegrationIDRequired
	}

	if provider, ok := criteria.Provider.Get(); ok && provider == types.ProviderUnknown {
		return ErrResolverProviderRequired
	}

	return nil
}

// resolveProviderAndIntegration resolves provider and installed integration from criteria
func (r *Resolver) resolveProviderAndIntegration(ctx context.Context, criteria ResolveCriteria) (types.ProviderType, *entgen.Integration, error) {
	if integrationID, ok := criteria.IntegrationID.Get(); ok {
		return r.resolveByIntegrationID(ctx, criteria, integrationID)
	}

	provider, ok := criteria.Provider.Get()
	if !ok {
		return types.ProviderUnknown, nil, ErrResolverProviderRequired
	}

	records, err := r.source.IntegrationsByProvider(ctx, criteria.OwnerID, provider)
	if err != nil {
		return types.ProviderUnknown, nil, err
	}

	switch len(records) {
	case 0:
		return types.ProviderUnknown, nil, ErrResolverIntegrationNotFound
	case 1:
		return provider, records[0], nil
	default:
		return types.ProviderUnknown, nil, ErrResolverIntegrationAmbiguous
	}
}

// resolveByIntegrationID resolves provider and integration using explicit integration id input
func (r *Resolver) resolveByIntegrationID(ctx context.Context, criteria ResolveCriteria, integrationID string) (types.ProviderType, *entgen.Integration, error) {
	record, err := r.source.IntegrationByID(ctx, criteria.OwnerID, integrationID)
	if err != nil {
		return types.ProviderUnknown, nil, err
	}
	if record == nil {
		return types.ProviderUnknown, nil, ErrResolverIntegrationNotFound
	}

	resolvedProvider := types.ProviderTypeFromString(record.Kind)
	if resolvedProvider == types.ProviderUnknown {
		return types.ProviderUnknown, nil, ErrResolverProviderUnknown
	}

	if provider, ok := criteria.Provider.Get(); ok && provider != resolvedProvider {
		return types.ProviderUnknown, nil, ErrResolverProviderMismatch
	}

	return resolvedProvider, record, nil
}

// ResolveOperation resolves a single operation descriptor for a known provider without a database
// lookup. Use this when the integration record is already in hand (e.g. after EnsureIntegration)
// to avoid a redundant round-trip.
func (r *Resolver) ResolveOperation(provider types.ProviderType, criteria ResolveCriteria) (types.OperationDescriptor, error) {
	if provider == types.ProviderUnknown {
		return types.OperationDescriptor{}, ErrResolverProviderRequired
	}
	if !criteria.OperationName.IsPresent() && !criteria.OperationKind.IsPresent() {
		return types.OperationDescriptor{}, ErrResolverOperationCriteriaRequired
	}

	return r.resolveOperationDescriptor(provider, criteria)
}

// resolveOperationDescriptor resolves one operation descriptor from provider and operation criteria
func (r *Resolver) resolveOperationDescriptor(provider types.ProviderType, criteria ResolveCriteria) (types.OperationDescriptor, error) {
	descriptors := r.registry.OperationDescriptors(provider)
	if len(descriptors) == 0 {
		return types.OperationDescriptor{}, ErrResolverOperationNotRegistered
	}

	if operationName, ok := criteria.OperationName.Get(); ok {
		return resolveOperationDescriptorByName(descriptors, operationName, criteria)
	}

	operationKind, _ := criteria.OperationKind.Get()
	matches := lo.Filter(descriptors, func(descriptor types.OperationDescriptor, _ int) bool {
		return descriptor.Kind == operationKind
	})

	switch len(matches) {
	case 0:
		return types.OperationDescriptor{}, ErrResolverOperationNotRegistered
	case 1:
		return matches[0], nil
	default:
		return types.OperationDescriptor{}, ErrResolverOperationDescriptorAmbiguous
	}
}

// resolveOperationDescriptorByName resolves one operation descriptor by exact operation name
func resolveOperationDescriptorByName(descriptors []types.OperationDescriptor, operationName types.OperationName, criteria ResolveCriteria) (types.OperationDescriptor, error) {
	matches := lo.Filter(descriptors, func(descriptor types.OperationDescriptor, _ int) bool {
		return descriptor.Name == operationName
	})

	switch len(matches) {
	case 0:
		return types.OperationDescriptor{}, ErrResolverOperationNotRegistered
	case 1:
		if operationKind, ok := criteria.OperationKind.Get(); ok && matches[0].Kind != operationKind {
			return types.OperationDescriptor{}, ErrResolverOperationKindMismatch
		}

		return matches[0], nil
	default:
		return types.OperationDescriptor{}, ErrResolverOperationDescriptorAmbiguous
	}
}
