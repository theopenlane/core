package types //nolint:revive

import (
	"encoding/json"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
)

// InstallationRequest bundles the inputs used by installation upgrade and validation hooks
type InstallationRequest struct {
	// Integration is the target installation record
	Integration *generated.Integration
	// Credentials lists every stored credential bundle by connection name
	Credentials map[string]CredentialSet
	// UserInput is the stored installation-scoped user input document
	UserInput json.RawMessage
}

// InstallationRegistration is the definition's installation metadata layout, declared once
type InstallationRegistration struct {
	// InputRegistration is the layout name, schema, upgrade, and validation of the metadata type
	InputRegistration
	// Identifiable reports whether the metadata type derives its own identity; otherwise the installation keeps its own id
	Identifiable bool `json:"-"`
	// Identify recomputes the display identity from a stored metadata document
	Identify func(json.RawMessage) (IntegrationInstallationIdentity, error) `json:"-"`
}
