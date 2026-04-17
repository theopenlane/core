package googleworkspace

import (
	"time"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Google Workspace integration definition
	definitionID = types.NewDefinitionRef("def_01K0GWKSP000000000000000001")
	// installation is the typed installation metadata handle for the Google Workspace definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// workspaceCredential is the credential slot for Google Workspace OAuth credentials
	_, workspaceCredential = providerkit.CredentialSchema[googleWorkspaceCred]()
	// workspaceClient is the client ref for the Google Workspace Admin SDK
	workspaceClient = types.NewClientRef[*admin.Service]()
	// healthCheckSchema is the operation ref for the health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation ref for the directory sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// googleWorkspaceCred holds the provider-owned credential material for a Google Workspace installation
type googleWorkspaceCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
	// PrimaryDirectory marks this installation as the authoritative directory source for identity holder enrichment and lifecycle derivation
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}

// InstallationMetadata holds the stable Google Workspace directory target selected for one installation
type InstallationMetadata struct {
	// CustomerID is the Google Workspace customer identifier
	CustomerID string `json:"customerId,omitempty" jsonschema:"title=Customer ID"`
	// Domain is the primary domain of the Google Workspace customer
	Domain string `json:"domain,omitempty" jsonschema:"title=Domain"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.Domain,
		ExternalID:   m.CustomerID,
	}
}
