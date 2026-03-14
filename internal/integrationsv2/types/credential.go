package types

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/models"
	generated "github.com/theopenlane/core/internal/ent/generated"
)

// CredentialSet is the persisted credential bundle used by integrations
type CredentialSet = models.CredentialSet

// CredentialResolver loads installation-scoped credentials from an external credential service
type CredentialResolver interface {
	// LoadCredential resolves the credential for one installation and reports whether one exists
	LoadCredential(ctx context.Context, integration *generated.Integration) (CredentialSet, bool, error)
}

// CredentialPersistMode controls how one installation's credentials are stored
type CredentialPersistMode string

const (
	// CredentialPersistModeKeystore stores credentials in the external keystore
	CredentialPersistModeKeystore CredentialPersistMode = "keystore"
	// CredentialPersistModeEphemeral resolves credentials from environment without persisting them
	CredentialPersistModeEphemeral CredentialPersistMode = "ephemeral"
	// CredentialPersistModeNone indicates the definition requires no persisted credentials
	CredentialPersistModeNone CredentialPersistMode = "none"
)

// CredentialBuilderFunc normalizes or validates credentials for one integration record
type CredentialBuilderFunc func(ctx context.Context, integration *generated.Integration, value CredentialSet) (CredentialSet, error)

// CredentialRegistration declares how a definition accepts credentials
type CredentialRegistration struct {
	// Schema is the JSON schema used to collect credentials
	Schema json.RawMessage `json:"schema,omitempty"`
	// Normalize canonicalizes raw credential input
	Normalize CredentialBuilderFunc `json:"-"`
	// Validate verifies that the credential input is usable
	Validate CredentialBuilderFunc `json:"-"`
	// Persist controls how credentials are stored for this definition
	Persist CredentialPersistMode `json:"persist,omitempty"`
}
