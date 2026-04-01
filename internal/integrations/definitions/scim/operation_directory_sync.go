package scim

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// directorySyncAckMessage is the static message returned when the SCIM directory sync is invoked
const directorySyncAckMessage = "scim is push-based; sync is triggered by the external identity provider"

// DirectorySync is the SCIM directory sync operation configuration
type DirectorySync struct{}

// directorySyncResult is the operation result returned for push-based SCIM sync requests
type directorySyncResult struct {
	// Message describes the push-based sync state when the operation is invoked without payloads
	Message string `json:"message,omitempty"`
}

// Run returns the SCIM push-based sync acknowledgement
func (DirectorySync) Run() (json.RawMessage, error) {
	return providerkit.EncodeResult(directorySyncResult{
		Message: directorySyncAckMessage,
	}, ErrResultEncode)
}
