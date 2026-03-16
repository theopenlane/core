package types

import (
	"encoding/json"

	"github.com/theopenlane/core/common/models"
)

// CredentialSet is the persisted credential bundle used by integrations
type CredentialSet = models.CredentialSet

// CredentialRegistration declares how a definition accepts credentials
type CredentialRegistration struct {
	// Schema is the JSON schema used to collect credentials
	Schema json.RawMessage `json:"schema,omitempty"`
}
