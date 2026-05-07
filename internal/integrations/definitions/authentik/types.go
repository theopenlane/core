package authentik

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Authentik integration definition
	definitionID = types.NewDefinitionRef("def_01K0AUTHENTIK000000000000001")
	// integration is the typed installation metadata handle for the Authentik definition
	integration = types.NewInstallationRef(resolveInstallationMetadata)
	// authentikCredentialSchema is the JSON schema for the Authentik credential
	// authentikCredential is the typed runtime ref for resolving the credential
	authentikCredentialSchema, authentikCredential = providerkit.CredentialSchema[CredentialSchema]()
	// authentikClient is the client ref for the Authentik API client
	authentikClient = types.NewClientRef[*Client]()
	// healthCheckSchema, healthCheckOperation is the operation ref for the health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema, directorySyncOperation is the operation ref for directory sync
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// CredentialSchema holds the Authentik instance credentials for one installation
type CredentialSchema struct {
	// BaseURL is the base URL of the Authentik instance
	BaseURL string `json:"baseUrl" jsonschema:"required,title=Base URL"`
	// Token is the Authentik API token
	Token string `json:"token" jsonschema:"required,title=API Token"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// DisableGroupSync when true only syncs users, skipping groups and memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users disable group and membership sync operations"`
	// PrimaryDirectory marks this installation as the authoritative source for identity holder sync
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory,description=Mark this as the authoritative source for identity holder enrichment and lifecycle"`
	// FilterExpr limits imported records to envelopes matching a CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting"`
}

// InstallationMetadata holds the stable Authentik instance identity for one installation
type InstallationMetadata struct {
	// Domain is the primary domain of the Authentik instance
	Domain string `json:"domain,omitempty"`
	// TenantID is the UUID of the Authentik tenant
	TenantID string `json:"tenantId,omitempty"`
	// BaseURL is the base URL of the Authentik instance
	BaseURL string `json:"baseUrl,omitempty"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.Domain,
		ExternalID:   m.TenantID,
	}
}
