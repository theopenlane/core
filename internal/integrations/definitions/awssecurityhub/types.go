package awssecurityhub

import (
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the AWS Security Hub integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AWSSECHUB0000000000001")
	// awsAssumeRoleCredential is the assume-role credential slot shared by the AWS service clients in this definition
	awsAssumeRoleCredential = types.NewCredentialRef(Slug + ".assume_role")
	// awsSourceCredential is the optional static source credential slot used to assume the configured AWS role
	awsSourceCredential = types.NewCredentialRef(Slug + ".source")
	// SecurityHubClient is the client ref for the AWS Security Hub client used by this definition
	SecurityHubClient = types.NewClientRef[*securityhub.Client]()
	// AuditManagerClient is the client ref for the AWS Audit Manager client used by this definition
	AuditManagerClient = types.NewClientRef[*auditmanager.Client]()
	// HealthDefaultOperation is the operation ref for the AWS Security Hub health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// AssessmentsListOperation is the operation ref for the AWS Audit Manager assessments list operation
	AssessmentsListOperation = types.NewOperationRef[AssessmentsConfig]("assessments.list")
	// VulnerabilitiesCollectOperation is the operation ref for the Security Hub vulnerabilities collection operation
	VulnerabilitiesCollectOperation = types.NewOperationRef[FindingsConfig]("vulnerabilities.collect")
)

// Slug is the unique identifier for the AWS Security Hub integration
const Slug = "aws_security_hub"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// AssumeRoleCredentialSchema holds the AWS assume-role and collection-scope inputs shared by the service clients.
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

// SourceCredentialSchema holds the optional static source credential used to assume the configured AWS role.
type SourceCredentialSchema struct {
	// AccessKeyID is an optional static source credential when runtime IAM is unavailable
	AccessKeyID string `json:"accessKeyId"     jsonschema:"required,title=Access Key ID,description=Static source credential used when runtime IAM is unavailable."`
	// SecretAccessKey is the AWS secret access key for static credentials
	SecretAccessKey string `json:"secretAccessKey" jsonschema:"required,title=Secret Access Key"`
	// SessionToken is the AWS session token for static credentials
	SessionToken string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}
