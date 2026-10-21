package googledrive

import (
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// oauthRevision is the revision used for the latest oauth connection
	oauthRevision = 2
)

var (
	// definitionID is the stable identifier for the Google Drive integration definition
	definitionID = types.NewDefinitionRef("def_01K0GDRIVE00000000000000001")
	// oauthConnection is the Google Drive OAuth connection scoped to drive.file
	oauthConnection = types.ConnectionOf[googleDriveCred]().Revision(oauthRevision)
	// legacyOauthConnection is the retired drive.readonly OAuth connection installations created before the drive.file migration still use
	legacyOauthConnection = types.ConnectionOf[googleDriveCred]()
	// installation is the installation metadata layout
	installation = types.InstallationOf[InstallationMetadata]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// DriveClient wraps the client for Google operations
type DriveClient struct {
	// Svc is the authenticated Google Drive service client
	Svc *drive.Service
	// TokenSource yields the refreshed OAuth token backing Svc
	TokenSource oauth2.TokenSource
	// Legacy reports the client was built for the retired drive.readonly connection
	Legacy bool
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
	// SelectedDocuments lists the Google Docs selected in the Picker, set by the selection operation rather than the settings form
	SelectedDocuments []types.SelectedResource `json:"selectedDocuments,omitempty" jsonschema:"title=Selected Documents" jsonschema_extras:"readOnly=true"`
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
