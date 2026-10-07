package types //nolint:revive

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/common/models"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
)

// CredentialSet is the persisted credential bundle used by integrations
type CredentialSet = models.CredentialSet

// ConnectionInput is the erased input every connection closure receives; Client is set for verification, UserInput for disconnect
type ConnectionInput struct {
	// Integration is the target installation record
	Integration *generated.Integration
	// Credential is the persisted credential bundle for the connection
	Credential CredentialSet
	// TokenManager signs assertions for providers that authenticate via identity federation
	TokenManager *tokens.TokenManager
	// Client is the built client the verification runs against
	Client any
	// UserInput is the stored installation-scoped user input document consumed by disconnect
	UserInput json.RawMessage
}

// ClientBuilderFunc builds one client from the erased input
type ClientBuilderFunc func(ctx context.Context, input ConnectionInput) (any, error)
