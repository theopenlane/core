package microsoftteams

import (
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Microsoft Teams integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0MSTEAMS00000000000000001")
	// installation is the installation metadata layout
	installation = types.InstallationOf[InstallationMetadata]()
	// oauthConnection is the Microsoft Teams OAuth connection
	oauthConnection = types.ConnectionOf[teamsCred]()
	// userInput is the installation user input layout for the Microsoft Teams definition
	userInput = types.UserInputRefOf[UserInput]()
)

// teamsCred holds the provider-owned credential material for a Microsoft Teams installation
type teamsCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// DefaultMessaging marks this installation as the preferred tenant for messaging
	DefaultMessaging bool `json:"defaultMessaging,omitempty" jsonschema:"title=Default Messaging"`
}

// InstallationMetadata holds the stable Microsoft tenant identity for one Teams installation
type InstallationMetadata struct {
	// TenantID is the Microsoft Entra tenant identifier extracted from the access token when available
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.TenantID,
	}
}
