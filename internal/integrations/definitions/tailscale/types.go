package tailscale

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Tailscale integration definition
	definitionID = types.NewDefinitionRef("def_01K0TAILSCALE0000000000001")
	// installation is the typed installation metadata handle for the Tailscale definition
	installation = types.InstallationOf[InstallationMetadata]()
	// tailscaleConnection is the typed connection handle for the Tailscale OAuth client
	tailscaleConnection = types.ConnectionOf[CredentialSchema]()
)

// AssetSync holds configuration for the Tailscale asset sync operation
type AssetSync struct {
	types.OperationSettings
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
