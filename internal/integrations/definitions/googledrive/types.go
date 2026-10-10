package googledrive

import (
	"time"

	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Google Drive integration definition
	definitionID = types.NewDefinitionRef("def_01K0GDRIVE00000000000000001")
	// oauthConnection is the Google Drive OAuth connection
	oauthConnection = types.ConnectionOf[googleDriveCred]()
	// installation is the installation metadata layout
	installation = types.InstallationOf[InstallationMetadata]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// DriveClient wraps the client for Google operations
type DriveClient struct {
	// Svc is the authenticated Google Drive service client
	Svc *drive.Service
}

// googleDriveCred holds the provider-owned credential material for a Google Drive installation
type googleDriveCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Primary marks this installation as the authoritative Drive source for live document exports
	Primary bool `json:"primary,omitempty" jsonschema:"title=Primary"`
}

// InstallationMetadata holds the stable Google Drive target selected for one installation
type InstallationMetadata struct {
	// AccountID is the stable id of the connected Google account as reported by the Drive About API
	AccountID string `json:"accountId,omitempty" jsonschema:"title=Account ID"`
	// Domain is the primary domain of the Google Workspace account
	Domain string `json:"domain,omitempty" jsonschema:"title=Domain"`
}

// DefinitionID returns the stable definition identifier for the Google Drive integration
func DefinitionID() string {
	return definitionID.ID()
}

// ExportOperationName returns the registered operation name for the document export operation
func ExportOperationName() string {
	return documentExportOperation.Name()
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.Domain,
		ExternalID:   m.AccountID,
	}
}
