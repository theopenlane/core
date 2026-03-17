package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Azure Entra ID integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AZENTRA0000000000000001")
	// EntraClient is the client ref for the Microsoft Graph API client used by this definition
	EntraClient = types.NewClientRef[*providerkit.AuthenticatedClient]()
	// HealthDefaultOperation is the operation ref for the Azure Entra ID health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	// DirectoryInspectOperation is the operation ref for the Azure Entra ID directory inspect operation
	DirectoryInspectOperation = types.NewOperationRef[DirectoryInspect]("directory.inspect")
)

// Slug is the unique identifier for the Azure Entra ID integration
const Slug = "azure_entra_id"

const (
	// azureAuthURL is the Microsoft identity platform authorization endpoint for the Azure Entra ID OAuth flow
	azureAuthURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	// azureTokenURL is the Microsoft identity platform token endpoint for the Azure Entra ID OAuth flow
	azureTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

// azureEntraScopes lists the Microsoft Graph OAuth scopes requested for the Azure Entra ID flow
var azureEntraScopes = []string{
	"https://graph.microsoft.com/.default",
	"offline_access",
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// TenantID is the Azure Entra ID tenant identifier
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}
