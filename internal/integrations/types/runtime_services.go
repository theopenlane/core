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
	// ExecuteRuntimeOperation runs a system-initiated operation inline, untracked
	ExecuteRuntimeOperation(ctx context.Context, definitionID, operationName string, config json.RawMessage) (json.RawMessage, error)
	// Dispatch enqueues one integration operation through the runtime-managed dispatcher
	Dispatch(ctx context.Context, req DispatchRequest) (DispatchResult, error)
	// ReapExpiredInstallation soft-deletes an expired never-connected installation
	ReapExpiredInstallation(ctx context.Context, integrationID string) (bool, error)
}
