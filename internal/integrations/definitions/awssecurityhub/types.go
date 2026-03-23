package awssecurityhub

import (
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the AWS Security Hub integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AWSSECHUB0000000000001")
	// Installation is the typed installation metadata handle for the AWS Security Hub definition
	Installation = types.NewInstallationRef(resolveInstallationMetadata)

	// awsAssumeRoleSchema is the reflected JSON schema for the assume-role credential
	// awsAssumeRoleCredential is the assume-role credential slot shared by the AWS service clients in this definition
	awsAssumeRoleSchema, awsAssumeRoleCredential = providerkit.CredentialSchema[AssumeRoleCredentialSchema]()
	// awsSourceSchema is the reflected JSON schema for the optional source credential
	// awsSourceCredential is the optional static source credential slot used to assume the configured AWS role
	awsSourceSchema, awsSourceCredential = providerkit.CredentialSchema[SourceCredentialSchema]()

	// SecurityHubClient is the client ref for the AWS Security Hub client used by this definition
	SecurityHubClient = types.NewClientRef[*securityhub.Client]()
	// AuditManagerClient is the client ref for the AWS Audit Manager client used by this definition
	AuditManagerClient = types.NewClientRef[*auditmanager.Client]()

	// HealthDefaultOperation is the operation ref for the AWS Security Hub health check
	_, HealthDefaultOperation = providerkit.OperationSchema[HealthCheck]()
	// assessmentsCollectSchema is the reflected JSON schema for the assessments collect operation config
	// AssessmentsCollectOperation is the operation ref for the AWS Audit Manager assessments collection operation
	assessmentsCollectSchema, AssessmentsCollectOperation = providerkit.OperationSchema[AssessmentsConfig]()
	// vulnerabilitiesCollectSchema is the reflected JSON schema for the vulnerabilities collect operation config
	// VulnerabilitiesCollectOperation is the operation ref for the Security Hub vulnerabilities collection operation
	vulnerabilitiesCollectSchema, VulnerabilitiesCollectOperation = providerkit.OperationSchema[FindingsConfig]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// AssumeRoleCredentialSchema holds the AWS assume-role and collection-scope inputs shared by the service clients
type AssumeRoleCredentialSchema struct {
	// RoleARN is the cross-account IAM role ARN Openlane should assume in the tenant environment
	RoleARN string `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN,description=Cross-account role Openlane should assume in the tenant environment."`
	// ExternalID is the external ID required in the tenant role trust policy
	ExternalID string `json:"externalId"                jsonschema:"required,title=External ID,description=External ID required in the tenant role trust policy."`
	// HomeRegion is the AWS region where Security Hub cross-region aggregation is managed
	HomeRegion string `json:"homeRegion"                jsonschema:"required,title=Home Region,description=AWS region used for Security Hub aggregation and other service API calls (e.g. us-east-1)."`
	// AccountID is the AWS account ID for reference in assessment summaries and run metadata
	AccountID string `json:"accountId,omitempty"       jsonschema:"title=Account ID,description=Optional AWS account ID for reference in results and reporting."`
	// AccountScope controls whether collection covers all delegated accounts or a subset
	AccountScope string `json:"accountScope,omitempty"    jsonschema:"title=Account Scope,description=Collect from all delegated accounts or restrict to specific account IDs.,enum=all,enum=specific"`
	// AccountIDs lists the specific AWS account IDs used when account scope is specific
	AccountIDs []string `json:"accountIds,omitempty"      jsonschema:"title=Account IDs,description=Required when accountScope is specific."`
	// LinkedRegions limits findings collection to the listed source regions
	LinkedRegions []string `json:"linkedRegions,omitempty"   jsonschema:"title=Linked Regions,description=Filter findings to these source regions. Empty means all regions."`
	// SessionName is an optional STS session name override
	SessionName string `json:"sessionName,omitempty"     jsonschema:"title=Session Name,description=Optional STS session name override."`
	// SessionDuration is an optional STS session duration override
	SessionDuration string `json:"sessionDuration,omitempty" jsonschema:"title=Session Duration,description=Optional STS session duration (e.g. 1h)."`
}

// SourceCredentialSchema holds the optional static source credential used to assume the configured AWS role
type SourceCredentialSchema struct {
	// AccessKeyID is an optional static source credential when runtime IAM is unavailable
	AccessKeyID string `json:"accessKeyId"     jsonschema:"required,title=Access Key ID,description=Static source credential used when runtime IAM is unavailable."`
	// SecretAccessKey is the AWS secret access key for static credentials
	SecretAccessKey string `json:"secretAccessKey" jsonschema:"required,title=Secret Access Key"`
	// SessionToken is the AWS session token for static credentials
	SessionToken string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}

// InstallationMetadata holds the non-secret AWS connection attributes persisted for one installation
type InstallationMetadata struct {
	// RoleARN is the cross-account IAM role ARN Openlane assumes for this installation
	RoleARN string `json:"roleArn,omitempty" jsonschema:"title=IAM Role ARN"`
	// HomeRegion is the AWS region used for Security Hub aggregation and API calls
	HomeRegion string `json:"homeRegion,omitempty" jsonschema:"title=Home Region"`
	// AccountID is the primary AWS account identifier when supplied during setup
	AccountID string `json:"accountId,omitempty" jsonschema:"title=Account ID"`
	// AccountScope indicates whether collection targets all delegated accounts or a specific set
	AccountScope string `json:"accountScope,omitempty" jsonschema:"title=Account Scope"`
	// AccountIDs lists the explicitly selected AWS account identifiers when account scope is specific
	AccountIDs []string `json:"accountIds,omitempty" jsonschema:"title=Account IDs"`
	// LinkedRegions limits collection to the listed AWS source regions when configured
	LinkedRegions []string `json:"linkedRegions,omitempty" jsonschema:"title=Linked Regions"`
	// UsesSourceCredential reports whether a static source credential is persisted alongside the assume-role configuration
	UsesSourceCredential bool `json:"usesSourceCredential,omitempty" jsonschema:"title=Uses Source Credential"`
}
