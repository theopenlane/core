package azuresecuritycenter

import "github.com/theopenlane/core/internal/integrations/types"

var (
	// DefinitionID is the stable identifier for the Azure Security Center integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AZSECC000000000000000001")
	// SecurityCenterClient is the client ref for the Azure management API client used by this definition
	SecurityCenterClient = types.NewClientRef[*azurePricingsClient]()
	// HealthDefaultOperation is the operation ref for the Azure Security Center health check
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	// SecurityPricingOverviewOperation is the operation ref for the Defender pricing overview operation
	SecurityPricingOverviewOperation = types.NewOperationRef[struct{}]("security.pricing_overview")
)

// Slug is the unique identifier for the Azure Security Center integration
const Slug = "azure_security_center"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty"    jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// ResourceGroup optionally scopes operations to a specific Azure resource group
	ResourceGroup string `json:"resourceGroup,omitempty" jsonschema:"title=Resource Group"`
	// WorkspaceID is the optional Log Analytics workspace identifier
	WorkspaceID string `json:"workspaceId,omitempty"   jsonschema:"title=Log Analytics Workspace ID"`
}

// CredentialSchema holds the Azure Security Center client credentials for one installation
type CredentialSchema struct {
	// TenantID is the Azure Active Directory tenant identifier
	TenantID string `json:"tenantId"                jsonschema:"required,title=Tenant ID"`
	// ClientID is the Azure application (service principal) client identifier
	ClientID string `json:"clientId"                jsonschema:"required,title=Client ID"`
	// ClientSecret is the Azure application client secret
	ClientSecret string `json:"clientSecret"            jsonschema:"required,title=Client Secret"`
	// SubscriptionID is the Azure subscription identifier scoping Security Center resources
	SubscriptionID string `json:"subscriptionId"          jsonschema:"required,title=Subscription ID"`
	// ResourceGroup optionally scopes operations to a specific Azure resource group
	ResourceGroup string `json:"resourceGroup,omitempty" jsonschema:"title=Resource Group"`
	// WorkspaceID is the optional Log Analytics workspace identifier
	WorkspaceID string `json:"workspaceId,omitempty"   jsonschema:"title=Log Analytics Workspace ID"`
}
