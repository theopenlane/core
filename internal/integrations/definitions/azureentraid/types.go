package azureentraid

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Azure Entra ID integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AZENTRA0000000000000001")
	// Installation is the typed installation metadata handle for the Azure Entra ID definition
	Installation = types.NewInstallationRef(resolveInstallationMetadata)
	// entraTenantCredential is the credential slot shared by the Entra clients in this definition
	entraTenantCredential = types.NewCredentialRef[entraIDCred](Slug)
	// EntraCredential is the client ref for the Azure token credential used by the health check
	EntraCredential = types.NewClientRef[azcore.TokenCredential]()
	// EntraClient is the client ref for the Microsoft Graph service client used by directory operations
	EntraClient = types.NewClientRef[*msgraphsdk.GraphServiceClient]()
	// HealthDefaultOperation is the operation ref for the Azure Entra ID health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// DirectoryInspectOperation is the operation ref for the Azure Entra ID directory inspect operation
	DirectoryInspectOperation = types.NewOperationRef[DirectoryInspect]("directory.inspect")
	// DirectorySyncOperation is the operation ref for the Azure Entra ID directory sync operation
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

// Slug is the unique identifier for the Azure Entra ID integration
const Slug = "azure_entra_id"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// EnableGroupSync controls whether group and membership records are collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
	// IncludeGuestUsers controls whether guest-type accounts are included in the sync
	IncludeGuestUsers bool `json:"includeGuestUsers,omitempty" jsonschema:"title=Include Guest Users"`
}

// entraIDCred holds the provider-owned credential material for an Azure Entra ID installation
type entraIDCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
	// TenantID is the Azure AD tenant identifier extracted from OIDC claims
	TenantID string `json:"tenantId"`
}

// CredentialSchema holds the per-installation credential for one Entra ID tenant
type CredentialSchema struct {
	// TenantID is the Azure Active Directory tenant identifier for this installation
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}

// InstallationMetadata holds the stable Azure Entra tenant identity for one installation
type InstallationMetadata struct {
	// TenantID is the Azure Active Directory tenant identifier selected during setup
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}
