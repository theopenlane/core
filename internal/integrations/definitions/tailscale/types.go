package tailscale

import (
	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Tailscale integration definition
	definitionID = types.NewDefinitionRef("def_01K0TAILSCALE0000000000001")
	// installation is the typed installation metadata handle for the Tailscale definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// tailscaleCredential is the credential slot for the Tailscale integration definition
	tailscaleCredential = types.CredentialRefOf[CredentialSchema]()
	// tailscaleClient is the client ref for the Tailscale API client used by this definition
	tailscaleClient = types.ClientRefOf[*tsclient.Client]().Using(tailscaleCredential)
	// tailscaleConnection is the connection mode selected by the Tailscale OAuth client credential
	tailscaleConnection = types.NewConnectionRef(tailscaleCredential).Enables(tailscaleClient)
	// userInput is the installation user input layout for the Tailscale definition
	userInput = types.NewUserInputRef[UserInput]("tailscale")
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
	// Disable switches the directory sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users and groups from Tailscale"`
	// DisableGroupSync skips group and membership sync, importing only users
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from Tailscale; disable role-based group sync"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.status == 'active'"`
}

// AssetSync holds configuration for the Tailscale asset sync operation
type AssetSync struct {
	// Disable switches the asset sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of devices from Tailscale"`
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
	// Tailnet is the name of the tailnet the OAuth client is scoped to
	Tailnet string `json:"tailnet,omitempty" jsonschema:"title=Tailnet"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID:   m.Tailnet,
		ExternalName: m.Tailnet,
	}
}
