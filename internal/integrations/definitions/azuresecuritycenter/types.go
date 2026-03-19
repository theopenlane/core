package azuresecuritycenter

import "github.com/theopenlane/core/internal/integrations/types"

var (
	// DefinitionID is the stable identifier for the Azure Security Center integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AZSECC000000000000000001")
	// SecurityCenterClient is the client ref for the Azure Security Center client used by this definition
	SecurityCenterClient = types.NewClientRef[*azureSecurityClient]()
	// HealthDefaultOperation is the operation ref for the Azure Security Center health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// AssessmentsCollectOperation is the operation ref for collecting security assessment findings
	AssessmentsCollectOperation = types.NewOperationRef[AssessmentsCollect]("assessments.collect")
	// SubAssessmentsCollectOperation is the operation ref for collecting sub-assessment vulnerability findings
	SubAssessmentsCollectOperation = types.NewOperationRef[SubAssessmentsCollect]("subassessments.collect")
)

// Slug is the unique identifier for the Azure Security Center integration
const Slug = "azure_security_center"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// CredentialSchema holds the Azure service principal credentials for one installation
type CredentialSchema struct {
	// TenantID is the Azure Active Directory tenant identifier
	TenantID string `json:"tenantId"       jsonschema:"required,title=Tenant ID"`
	// ClientID is the Azure application (service principal) client identifier
	ClientID string `json:"clientId"       jsonschema:"required,title=Client ID"`
	// ClientSecret is the Azure application client secret
	ClientSecret string `json:"clientSecret"   jsonschema:"required,title=Client Secret"`
	// SubscriptionID is the Azure subscription identifier scoping Security Center resources
	SubscriptionID string `json:"subscriptionId" jsonschema:"required,title=Subscription ID"`
}
