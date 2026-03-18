package types

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
)

// InstallationRequest bundles the inputs used to resolve installation metadata
type InstallationRequest struct {
	// Installation is the target installation record
	Installation *generated.Integration
	// Credential is the installation-scoped credential bundle
	Credential CredentialSet
	// Config is the installation-scoped configuration payload
	Config IntegrationConfig
	// Input is provider-defined raw input used to derive installation metadata
	Input json.RawMessage
}

// InstallationFunc derives installation metadata for one definition installation
type InstallationFunc func(ctx context.Context, req InstallationRequest) (IntegrationInstallationMetadata, error)

// InstallationRegistration describes how a definition derives installation metadata
type InstallationRegistration struct {
	// Schema is the JSON schema used to validate installation metadata
	Schema json.RawMessage `json:"schema,omitempty"`
	// Resolve derives installation metadata for the definition
	Resolve InstallationFunc `json:"-"`
}
