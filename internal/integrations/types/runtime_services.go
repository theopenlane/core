package types //nolint:revive

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// RuntimeServices exposes the runtime capabilities available to definition-declared gala listeners
type RuntimeServices interface {
	// DB returns the runtime-managed ent client
	DB() *generated.Client
	// Gala returns the runtime-managed event runtime
	Gala() *gala.Gala
	// ExecuteRuntimeOperation runs one system-initiated operation inline with no installation or run tracking
	ExecuteRuntimeOperation(ctx context.Context, definitionID, operationName string, config json.RawMessage) (json.RawMessage, error)
	// Dispatch enqueues one integration operation through the runtime-managed dispatcher
	Dispatch(ctx context.Context, req DispatchRequest) (DispatchResult, error)
	// RemoveInstallation deletes one installation's credentials and record without provider teardown
	RemoveInstallation(ctx context.Context, integrationID string) error
	// ProbeIntegrationRecovery re-verifies an errored installation and clears it on success
	ProbeIntegrationRecovery(ctx context.Context, installation *generated.Integration) error
	// MarkIntegrationUnhealthy flags one installation as errored with a user-facing reason
	MarkIntegrationUnhealthy(ctx context.Context, installation *generated.Integration, reason string) error
}
