package engine

import (
	"context"

	"github.com/samber/mo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
)

// integrationResolveCriteria specifies the lookup parameters for resolving an integration record
type integrationResolveCriteria struct {
	// OwnerID is the organization that owns the integration
	OwnerID string
	// IntegrationID narrows resolution to a specific integration record when set
	IntegrationID mo.Option[string]
	// Provider narrows resolution to a specific provider kind when set
	Provider mo.Option[types.ProviderType]
}

// integrationResolution holds the result of resolving an integration record
type integrationResolution struct {
	// Integration is the resolved integration record
	Integration *ent.Integration
	// Provider is the provider type derived from the integration record
	Provider types.ProviderType
}

// integrationEntResolver resolves integration records from the ent database
type integrationEntResolver struct {
	client *ent.Client
}

// newIntegrationEntResolver creates an integrationEntResolver backed by the given ent client
func newIntegrationEntResolver(client *ent.Client) *integrationEntResolver {
	return &integrationEntResolver{client: client}
}

// Resolve queries the Integration table using the provided criteria and returns the matching record
// and its derived provider type. When IntegrationID is set, it queries by ID (filtered by OwnerID).
// When only Provider is set, it queries by OwnerID and Kind. ErrIntegrationRecordMissing is returned
// when no record matches.
func (r *integrationEntResolver) Resolve(ctx context.Context, criteria integrationResolveCriteria) (integrationResolution, error) {
	q := r.client.Integration.Query().
		Where(integration.OwnerIDEQ(criteria.OwnerID))

	if id, ok := criteria.IntegrationID.Get(); ok {
		q = q.Where(integration.IDEQ(id))
	}

	if provider, ok := criteria.Provider.Get(); ok && provider != types.ProviderUnknown {
		q = q.Where(integration.KindEQ(string(provider)))
	}

	record, err := q.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return integrationResolution{}, ErrIntegrationRecordMissing
		}

		return integrationResolution{}, err
	}

	return integrationResolution{
		Integration: record,
		Provider:    types.ProviderTypeFromString(record.Kind),
	}, nil
}
