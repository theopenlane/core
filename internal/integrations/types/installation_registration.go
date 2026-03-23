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
	// Connection is the resolved connection mode for the installation when available
	Connection ConnectionRegistration
	// Credential is the primary credential bundle for convenience when one slot is used
	Credential CredentialSet
	// Credentials lists all resolved credential bundles participating in the connection mode
	Credentials CredentialBindings
	// Config is the installation-scoped configuration payload
	Config IntegrationConfig
	// Input is provider-defined raw input used to derive installation metadata
	Input json.RawMessage
}

// InstallationFunc derives, validates, and marshals installation metadata for one connection-backed installation
// The bool return indicates whether metadata was produced; false with a nil error means the connection
// does not yield metadata for this installation
type InstallationFunc func(ctx context.Context, req InstallationRequest) (IntegrationInstallationMetadata, bool, error)

// InstallationRegistration describes how one connection mode derives installation metadata
type InstallationRegistration struct {
	// Resolve derives installation metadata for the connection mode
	Resolve InstallationFunc `json:"-"`
}
