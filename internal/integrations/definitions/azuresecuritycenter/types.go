package azuresecuritycenter

import "github.com/theopenlane/core/internal/integrations/types"

var (
	// DefinitionID is the stable identifier for the Azure Security Center integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AZSECC000000000000000001")
	// SecurityCenterClient is the client ref for the Azure management API client used by this definition
	SecurityCenterClient = types.NewClientRef[*azurePricingsClient]()
	// HealthDefaultOperation is the operation ref for the Azure Security Center health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	// SecurityPricingOverviewOperation is the operation ref for the Defender pricing overview operation
	SecurityPricingOverviewOperation = types.NewOperationRef[SecurityPricingOverview]("security.pricing_overview")
)

// Slug is the unique identifier for the Azure Security Center integration
const Slug = "azure_security_center"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	FilterExpr    string `json:"filterExpr,omitempty"    jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	ResourceGroup string `json:"resourceGroup,omitempty" jsonschema:"title=Resource Group"`
	WorkspaceID   string `json:"workspaceId,omitempty"   jsonschema:"title=Log Analytics Workspace ID"`
}

// CredentialSchema holds the Azure Security Center client credentials for one installation
type CredentialSchema struct {
	TenantID       string `json:"tenantId"                jsonschema:"required,title=Tenant ID"`
	ClientID       string `json:"clientId"                jsonschema:"required,title=Client ID"`
	ClientSecret   string `json:"clientSecret"            jsonschema:"required,title=Client Secret"`
	SubscriptionID string `json:"subscriptionId"          jsonschema:"required,title=Subscription ID"`
	ResourceGroup  string `json:"resourceGroup,omitempty" jsonschema:"title=Resource Group"`
	WorkspaceID    string `json:"workspaceId,omitempty"   jsonschema:"title=Log Analytics Workspace ID"`
}
