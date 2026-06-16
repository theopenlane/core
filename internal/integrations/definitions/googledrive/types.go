package googledrive

import (
	"time"

	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Google Drive integration definition
	definitionID = types.NewDefinitionRef("def_01K0GDRIVE00000000000000001")
	// installation is the typed installation metadata handle for the Google Drive definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// driveCredential is the credential slot for Google Drive OAuth credentials
	_, driveCredential = providerkit.CredentialSchema[googleDriveCred]()
	// driveClient is the client ref for the Google Drive SDK
	driveClient = types.NewClientRef[DriveClient]()
	// healthCheckSchema is the operation ref for the health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// documentExportSchema is the operation ref for the document export operation
	documentExportSchema, documentExportOperation = providerkit.OperationSchema[operations.DocumentExport]()
	// folderSyncSchema is the operation ref for the folder sync operation
	folderSyncSchema, folderSyncOperation = providerkit.OperationSchema[FolderSync]()
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
	// FolderID is the Google Drive folder ID (or full URL) containing policy documents
	FolderID string `json:"folderId,omitempty" jsonschema:"title=Folder ID,description=Google Drive folder ID or URL containing policy documents"`
	// FilterExpr is an optional CEL expression to filter which documents in the folder are eligible
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to filter documents before creating policies"`
}

// InstallationMetadata holds the stable Google Drive target selected for one installation
type InstallationMetadata struct {
	// CustomerID is the Google Workspace customer identifier (empty for personal accounts)
	CustomerID string `json:"customerId,omitempty" jsonschema:"title=Customer ID"`
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
		ExternalID:   m.CustomerID,
	}
}
