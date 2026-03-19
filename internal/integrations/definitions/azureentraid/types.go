package azureentraid

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Azure Entra ID integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AZENTRA0000000000000001")
	// entraTenantCredential is the credential slot shared by the Entra clients in this definition
	entraTenantCredential = types.NewCredentialRef(Slug)
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

// CredentialSchema holds the per-installation credential for one Entra ID tenant
type CredentialSchema struct {
	// TenantID is the Azure Active Directory tenant identifier for this installation
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}
