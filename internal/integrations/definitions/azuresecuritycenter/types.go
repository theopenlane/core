package azuresecuritycenter

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Azure Security Center integration definition
	definitionID = types.NewDefinitionRef("def_01K0AZSECC000000000000000001")
	// installation is the typed installation metadata handle for the Azure Security Center definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// securityCenterSchema is the credential schema for the Azure Security Center integration definition
	securityCenterSchema, securityCenterCredential = providerkit.CredentialSchema[CredentialSchema]()
	// securityCenterClient is the client ref for the Azure Security Center client
	securityCenterClient = types.NewClientRef[*azureSecurityClient]()
	// healthDefaultOperation is the operation ref for the Azure Security Center health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// assessmentsCollectSchema is the operation ref for the Azure Security Center assessments collect operation
	assessmentsCollectSchema, assessmentsCollectOperation = providerkit.OperationSchema[AssessmentsCollect]()
	// subAssessmentsCollectSchema is the operation ref for the Azure Security Center sub-assessments collect operation
	subAssessmentsCollectSchema, subAssessmentsCollectOperation = providerkit.OperationSchema[SubAssessmentsCollect]()
)

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

// InstallationMetadata holds the stable Azure subscription identity for one installation
type InstallationMetadata struct {
	// TenantID is the Azure Active Directory tenant identifier
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
	// ClientID is the Azure application identifier used for this connection
	ClientID string `json:"clientId,omitempty" jsonschema:"title=Client ID"`
	// SubscriptionID is the Azure subscription identifier scoped to this installation
	SubscriptionID string `json:"subscriptionId,omitempty" jsonschema:"title=Subscription ID"`
}
