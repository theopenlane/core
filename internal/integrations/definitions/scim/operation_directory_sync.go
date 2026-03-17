package scim

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// directorySyncAckMessage is the static message returned when the SCIM directory sync is invoked
const directorySyncAckMessage = "scim is push-based; sync is triggered by the external identity provider"

// DirectorySync is a no-op placeholder result for push-based SCIM directory sync
type DirectorySync struct {
	// Message describes the push-based sync state
	Message string `json:"message"`
}

// Handle adapts SCIM directory sync to the generic operation registration boundary
func (d DirectorySync) Handle() types.OperationHandler {
	return func(_ context.Context, _ types.OperationRequest) (json.RawMessage, error) {
		return d.Run()
	}
}

// Run returns the SCIM push-based sync acknowledgement
func (DirectorySync) Run() (json.RawMessage, error) {
	return providerkit.EncodeResult(DirectorySync{
		Message: directorySyncAckMessage,
	}, ErrResultEncode)
}
