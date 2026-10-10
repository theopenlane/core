package authentik

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Authentik integration definition
	definitionID = types.NewDefinitionRef("def_01K0AUTHENTIK000000000000001")
	// installation is the typed installation metadata handle for the Authentik definition
	installation = types.InstallationOf[InstallationMetadata]()
	// authentikConnection is the typed connection handle for the Authentik API token
	authentikConnection = types.ConnectionOf[CredentialSchema]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
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
	// PrimaryDirectory marks this installation as the authoritative source for identity holder sync
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory,description=Mark this as the authoritative source for identity holder enrichment and lifecycle"`
}

// InstallationMetadata holds the stable Authentik instance identity for one installation
type InstallationMetadata struct {
	// Brand is the Authentik instance brand name
	Brand string `json:"brand,omitempty"`
	// BrandID is the immutable uuid of the Authentik instance's default brand
	BrandID string `json:"brandId,omitempty"`
	// Host is the HTTP host of the Authentik instance
	Host string `json:"host,omitempty"`
	// BaseURL is the base URL of the Authentik instance
	BaseURL string `json:"baseUrl,omitempty"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.Brand,
		ExternalID:   m.BrandID,
	}
}
