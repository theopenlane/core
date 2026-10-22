package types //nolint:revive

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/iam/tokens"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
)

// Connection is one way a user connects a definition
type Connection struct {
	// Name is the user-facing connection name
	Name string
	// Description explains what the connection does
	Description string
	// Recommended indicates the connection that is recommended when there are multiple options
	Recommended bool
	// Meta is additional data the user might need to set up the connection
	Meta map[string]MetaInfo
	// Credential is the stored credential layout; Credential.Name is the persisted slot and connection name
	Credential InputRegistration
	// Form is the user-facing credential form schema, empty when Auth obtains the credential
	Form json.RawMessage
	// Replaces lists the retired connection names whose stored payloads move onto this connection, sorted
	Replaces []string
	// Auth describes how the connection performs auth when supported
	Auth *AuthRegistration
	// Disconnect describes how the connection tears down an installation
	Disconnect *DisconnectRegistration
	// Clients builds each client the connection provides, keyed by client name
	Clients map[string]ClientBuilderFunc
	// Verify probes the connection under the named client and derives the installation metadata
	Verify VerifyRegistration
}

// VerifyRegistration is the connection's verification
type VerifyRegistration struct {
	// ClientRef is the client the verification runs against; it must be one of Clients
	ClientRef string
	// Installation is the layout name of the metadata type the verification returns
	Installation string
	// Handle runs the verification and returns the marshalled metadata
	Handle func(context.Context, ConnectionInput) (IntegrationInstallationMetadata, error)
}

// Connector yields one connection
type Connector interface {
	// Connection returns the connection
	Connection() Connection
}

// Connection returns the connection itself, so a materialized connection is a Connector
func (c Connection) Connection() Connection {
	return c
}

// ConnectionRequest is what a client builder and the verification receive
type ConnectionRequest[T any] struct {
	// Integration is the target installation record
	Integration *generated.Integration
	// Credential is the decoded stored credential
	Credential T
	// TokenManager signs assertions for providers that authenticate via identity federation
	TokenManager *tokens.TokenManager
}

// DisconnectRequest is what a connection's disconnect hook receives
type DisconnectRequest[T any] struct {
	// Integration is the installation record being disconnected
	Integration *generated.Integration
	// Credential is the decoded stored credential
	Credential T
	// UserInput is the stored installation-scoped user input document
	UserInput json.RawMessage
}
