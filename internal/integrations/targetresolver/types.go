package targetresolver

import (
	"context"

	"github.com/samber/mo"

	"github.com/theopenlane/core/common/integrations/types"
	entgen "github.com/theopenlane/core/internal/ent/generated"
)

// OperationRegistry exposes provider operation descriptors to resolver consumers
type OperationRegistry interface {
	// OperationDescriptors returns all operation descriptors for a provider
	OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor
}

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
	// OperationName optionally constrains operation descriptor selection by exact name
	OperationName mo.Option[types.OperationName]
	// OperationKind optionally constrains operation descriptor selection by operation kind
	OperationKind mo.Option[types.OperationKind]
}

// ResolveResult captures the final integration and operation selected for execution
type ResolveResult struct {
	// Integration is the selected installed integration
	Integration *entgen.Integration
	// Provider is the selected provider kind
	Provider types.ProviderType
	// Operation is the selected operation descriptor
	Operation types.OperationDescriptor
}
