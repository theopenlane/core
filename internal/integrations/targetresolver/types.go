package targetresolver

import (
	"context"

	"github.com/samber/mo"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// IntegrationSource resolves installed integration records for owner and provider criteria
type IntegrationSource interface {
	// IntegrationByID returns an installed integration by owner and id
	IntegrationByID(ctx context.Context, ownerID string, integrationID string) (*entgen.Integration, error)
	// IntegrationsByProvider returns installed integrations for owner and provider
	IntegrationsByProvider(ctx context.Context, ownerID string, provider types.ProviderType) ([]*entgen.Integration, error)
}

// ResolveCriteria defines explicit constraints used to resolve an integration execution target
type ResolveCriteria struct {
	// OwnerID identifies the organization that owns the target integration
	OwnerID string
	// IntegrationID optionally constrains resolution to one installed integration id
	IntegrationID mo.Option[string]
	// Provider optionally constrains resolution to a provider kind
	Provider mo.Option[types.ProviderType]
}

// ResolveResult captures the final integration selected for execution.
type ResolveResult struct {
	// Integration is the selected installed integration
	Integration *entgen.Integration
	// Provider is the selected provider kind
	Provider types.ProviderType
}
