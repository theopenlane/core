package types //nolint:revive

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/pkg/gala"
)

// DefinitionLookup resolves registered definitions by id
type DefinitionLookup interface {
	// Definition returns the registered definition with the id and whether it exists
	Definition(id string) (Definition, bool)
}

// RuntimeServices exposes the runtime capabilities available to definition-declared gala listeners and operations
type RuntimeServices interface {
	// Gala returns the runtime-managed event runtime
	Gala() *gala.Gala
	// Definitions returns the registered definitions
	Definitions() DefinitionLookup
	// ExecuteRuntimeOperation runs one system-initiated operation inline with no installation or run tracking
	ExecuteRuntimeOperation(ctx context.Context, definitionID, operationName string, config json.RawMessage) (json.RawMessage, error)
	// Dispatch enqueues one integration operation through the runtime-managed dispatcher
	Dispatch(ctx context.Context, req DispatchRequest) (DispatchResult, error)
	// ReapExpiredInstallation soft-deletes one expired never-connected installation and its credentials, reporting whether it was reaped
	ReapExpiredInstallation(ctx context.Context, integrationID string) (bool, error)
	// ResolveOwnerIntegration returns the operational installation id for the definition and owner
	ResolveOwnerIntegration(ctx context.Context, definitionID, ownerID string, prefer ...func(*Installation) bool) (string, error)
}
