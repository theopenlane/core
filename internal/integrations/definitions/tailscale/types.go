package tailscale

import (
	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Tailscale integration definition
	definitionID = types.NewDefinitionRef("def_01K0TAILSCALE0000000000001")
	// installation is the typed installation metadata handle for the Tailscale definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// tailscaleSchema is the credential schema for the Tailscale integration definition
	tailscaleSchema, tailscaleCredential = providerkit.CredentialSchema[CredentialSchema]()
	// tailscaleClient is the client ref for the Tailscale API client used by this definition
	tailscaleClient = types.NewClientRef[*tsclient.Client]()
	// healthCheckSchema is the operation schema for the health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation schema for the directory sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
	// assetSyncSchema is the operation schema for the asset sync operation
	assetSyncSchema, assetSyncOperation = providerkit.OperationSchema[AssetSync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// DirectorySync includes the configuration for syncing users, groups, and memberships from Tailscale
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Sync"`
	// AssetSync includes the configuration for syncing Tailscale devices as assets
	AssetSync AssetSync `json:"assetSync,omitempty" jsonschema:"title=Asset Sync"`
}

// DirectorySync holds configuration for the Tailscale directory sync operation
type DirectorySync struct {
	// Disable stops the directory sync operation from running
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users and groups from Tailscale"`
	// DisableGroupSync skips group and membership sync, importing only users
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from Tailscale; disable role-based group sync"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.status == 'active'"`
}

// AssetSync holds configuration for the Tailscale asset sync operation
type AssetSync struct {
	// Disable stops the asset sync operation from running
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of Tailscale devices as assets"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting"`
}

// CredentialSchema holds the Tailscale OAuth credentials for one installation
type CredentialSchema struct {
	// ClientID is the OAuth client ID generated in the Tailscale admin console
	ClientID string `json:"clientId" jsonschema:"required,title=Client ID,description=OAuth client ID from the Tailscale admin console."`
	// ClientSecret is the OAuth client secret paired with ClientID
	ClientSecret string `json:"clientSecret" jsonschema:"required,title=Client Secret,description=OAuth client secret from the Tailscale admin console.,secret=true"`
}

// InstallationMetadata holds the stable Tailscale identity for one installation
type InstallationMetadata struct {
	// ClientID is the OAuth client ID used to connect this installation
	ClientID string `json:"clientId,omitempty" jsonschema:"title=Client ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.ClientID,
	}
}
