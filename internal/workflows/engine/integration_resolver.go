package engine

import (
	"context"

	"github.com/samber/mo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	types "github.com/theopenlane/core/internal/integrations/types"
)

// installationResolveCriteria specifies the lookup parameters for resolving an installation record
type installationResolveCriteria struct {
	// OwnerID is the organization that owns the installation
	OwnerID string
	// InstallationID narrows resolution to a specific installation record when set
	InstallationID mo.Option[string]
	// DefinitionID narrows resolution to a specific definition when set
	DefinitionID mo.Option[types.DefinitionID]
}

// installationResolution holds the result of resolving an installation record
type installationResolution struct {
	// Installation is the resolved installation record
	Installation *ent.Integration
	// DefinitionID is the definition identifier derived from the installation
	DefinitionID types.DefinitionID
}

// installationResolver resolves installation records from the ent database
type installationResolver struct {
	client *ent.Client
}

// newInstallationResolver creates an installationResolver backed by the given ent client
func newInstallationResolver(client *ent.Client) *installationResolver {
	return &installationResolver{client: client}
}

// Resolve queries the Integration table using the provided criteria and returns the matching record
// and its derived definition identifier. When InstallationID is set, it queries by ID (filtered
// by OwnerID). When DefinitionID is set, it queries by owner and definition.
// ErrInstallationNotFound is returned when no record matches.
func (r *installationResolver) Resolve(ctx context.Context, criteria installationResolveCriteria) (installationResolution, error) {
	q := r.client.Integration.Query().
		Where(integration.OwnerIDEQ(criteria.OwnerID))

	if id, ok := criteria.InstallationID.Get(); ok {
		q = q.Where(integration.IDEQ(id))
	}

	if defID, ok := criteria.DefinitionID.Get(); ok && string(defID) != "" {
		q = q.Where(integration.DefinitionIDEQ(string(defID)))
	}

	record, err := q.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return installationResolution{}, ErrInstallationNotFound
		}

		return installationResolution{}, err
	}

	return installationResolution{
		Installation: record,
		DefinitionID: types.DefinitionID(record.DefinitionID),
	}, nil
}
