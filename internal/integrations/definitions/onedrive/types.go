package onedrive

import (
	"time"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the OneDrive integration definition
	definitionID = types.NewDefinitionRef("def_01K0ONEDRIVE00000000000001")
	// installation is the typed installation metadata handle for the OneDrive definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// oneDriveCredential is the auth-managed credential slot for OneDrive OAuth credentials
	oneDriveCredential = types.CredentialRefOf[oneDriveCred]()
	// oneDriveClient is the client ref for the wrapped OneDrive graph client
	oneDriveClient = types.ClientRefOf[*DriveClient]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// oneDriveCred holds the provider-owned credential material for a OneDrive installation
type oneDriveCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// DriveClient wraps the Graph client for OneDrive operations
type DriveClient struct {
	// Graph is the authenticated Microsoft Graph service client
	Graph *msgraphsdk.GraphServiceClient
	// TS is the OAuth2 token source used to obtain access tokens for plain HTTP requests
	TS oauth2.TokenSource
	// Cfg is the operator-level configuration for export operations
	Cfg Config
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Primary marks this installation as the authoritative OneDrive source for live document exports
	Primary bool `json:"primary,omitempty" jsonschema:"title=Primary"`
}

// InstallationMetadata holds the stable OneDrive target selected for one installation
type InstallationMetadata struct {
	// TenantID is the Microsoft Entra tenant identifier
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
	// Domain is the primary domain of the Microsoft tenant
	Domain string `json:"domain,omitempty" jsonschema:"title=Domain"`
}

// DefinitionID returns the stable definition identifier for the OneDrive integration
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
		ExternalID:   m.TenantID,
	}
}
