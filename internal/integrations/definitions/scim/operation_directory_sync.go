package scim

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// directorySyncAckMessage is the static message returned when the SCIM directory sync is invoked
const directorySyncAckMessage = "scim is push-based; sync is triggered by the external identity provider"

// directorySyncOperation is the operation ref for the SCIM directory sync operation
var directorySyncOperation = types.OperationRefOf[DirectorySync]().
	HandlesRequest(runDirectorySync).
	Policy(types.ExecutionPolicy{Inline: true}).
	Ingest(providerkit.DirectoryIngestContracts()...).
	SkipDefaultLookback()

// DirectorySync is the SCIM directory sync operation configuration
type DirectorySync struct {
	types.OperationSettings
}

// directorySyncResult is the operation result returned for push-based SCIM sync requests
type directorySyncResult struct {
	// Message describes the push-based sync state when the operation is invoked without payloads
	Message string `json:"message,omitempty"`
}

// DirectorySyncOperationName returns the SCIM directory sync operation name
func DirectorySyncOperationName() string {
	return directorySyncOperation.Name()
}

// runDirectorySync returns the SCIM push-based sync acknowledgement
func runDirectorySync(_ context.Context, _ types.OperationRequest, _ DirectorySync) (json.RawMessage, error) {
	return providerkit.EncodeResult(directorySyncResult{
		Message: directorySyncAckMessage,
	}, ErrResultEncode)
}
