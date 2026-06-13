package onedrive

import (
	"time"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the OneDrive integration definition
	definitionID = types.NewDefinitionRef("def_01K0ONEDRIVE00000000000001")
	// installation is the typed installation metadata handle for the OneDrive definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// _, oneDriveCredential is the credential slot for OneDrive OAuth credentials
	_, oneDriveCredential = providerkit.CredentialSchema[oneDriveCred]()
	// oneDriveClient is the client ref for the wrapped OneDrive graph client
	oneDriveClient = types.NewClientRef[*DriveClient]()
	// healthCheckSchema is the operation schema for the health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// documentExportSchema is the operation schema for the document export operation
	documentExportSchema, documentExportOperation = providerkit.OperationSchema[operations.DocumentExport]()
	// folderSyncSchema is the operation schema for the folder sync operation
	folderSyncSchema, folderSyncOperation = providerkit.OperationSchema[FolderSync]()
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
	// Cfg is the operator-level configuration, carried so export operations can access
	// optional services like Azure Document Intelligence
	Cfg Config
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Primary marks this installation as the authoritative OneDrive source for live document exports
	Primary bool `json:"primary,omitempty" jsonschema:"title=Primary"`
	// FolderID is the folder path relative to the drive root (e.g. "Policies"); leave empty to sync the root
	FolderID string `json:"folderId,omitempty" jsonschema:"title=Folder Path,description=Folder path relative to drive root (e.g. Policies). Leave empty to sync the entire drive root."`
	// FilterExpr is an optional CEL expression to filter which documents in the folder are eligible
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to filter documents before creating policies"`
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
