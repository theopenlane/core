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
var directorySyncOperation = types.OperationRefOf[DirectorySync]().HandlesRequest(runDirectorySync)

// DirectorySync is the SCIM directory sync operation configuration
type DirectorySync struct {
	// Disable switches the directory sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users and groups from SCIM"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
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
