package targetresolver

import (
	"context"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Resolver resolves installed integrations from explicit criteria
type Resolver struct {
	source IntegrationSource
}

// EntSource resolves installed integrations from ent persistence
type EntSource struct {
	client *entgen.Client
}

// NewResolver constructs a target resolver with required dependencies
func NewResolver(source IntegrationSource) (*Resolver, error) {
	if source == nil {
		return nil, ErrResolverSourceRequired
	}

	return &Resolver{
		source: source,
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

// Resolve resolves an installed integration from explicit criteria
func (r *Resolver) Resolve(ctx context.Context, criteria ResolveCriteria) (ResolveResult, error) {
	if err := validateResolveCriteria(criteria); err != nil {
		return ResolveResult{}, err
	}

	provider, integrationRecord, err := r.resolveProviderAndIntegration(ctx, criteria)
	if err != nil {
		return ResolveResult{}, err
	}

	return ResolveResult{
		Integration: integrationRecord,
		Provider:    provider,
	}, nil
}

// validateResolveCriteria validates required fields and explicit constraints
func validateResolveCriteria(criteria ResolveCriteria) error {
	if criteria.OwnerID == "" {
		return ErrResolverOwnerIDRequired
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
