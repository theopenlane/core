package awsauditmanager

import (
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the AWS Audit Manager integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AWSAUDITM0000000000001")
	// AuditManagerClient is the client ref for the AWS Audit Manager client used by this definition
	AuditManagerClient = types.NewClientRef[*auditmanager.Client]()
	// HealthDefaultOperation is the operation ref for the AWS Audit Manager health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	// AssessmentsListOperation is the operation ref for the AWS Audit Manager assessments list operation
	AssessmentsListOperation = types.NewOperationRef[AssessmentsList]("assessments.list")
)

// Slug is the unique identifier for the AWS Audit Manager integration
const Slug = "aws_audit_manager"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty"  jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// AssessmentID optionally scopes collection to a single Audit Manager assessment
	AssessmentID string `json:"assessmentId,omitempty" jsonschema:"title=Assessment ID,description=Optional assessment ID to scope collection to a single assessment."`
}

// CredentialSchema holds the AWS STS role and optional static key material for one Audit Manager installation
// Fields are named to match awskit.ProviderData JSON tags so MetadataFromProviderData decodes them correctly
type CredentialSchema struct {
	// RoleARN is the cross-account IAM role ARN Openlane should assume in the tenant environment
	RoleARN string `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN,description=Cross-account role Openlane should assume in the tenant environment."`
	// ExternalID is the external ID required in the tenant role trust policy
	ExternalID string `json:"externalId"                jsonschema:"required,title=External ID,description=External ID required in the tenant role trust policy."`
	// HomeRegion is the AWS region where Audit Manager data is managed
	HomeRegion string `json:"homeRegion"                jsonschema:"required,title=Home Region,description=AWS region where Audit Manager data is managed (e.g. us-east-1)."`
	// AccountID is the AWS account ID for reference
	AccountID string `json:"accountId,omitempty"       jsonschema:"title=Account ID,description=AWS account ID for reference."`
	// SessionName is an optional STS session name override
	SessionName string `json:"sessionName,omitempty"     jsonschema:"title=Session Name,description=Optional STS session name override."`
	// SessionDuration is an optional STS session duration override
	SessionDuration string `json:"sessionDuration,omitempty" jsonschema:"title=Session Duration,description=Optional STS session duration (e.g. 1h)."`
	// AccessKeyID is an optional static source credential when runtime IAM is unavailable
	AccessKeyID string `json:"accessKeyId,omitempty"     jsonschema:"title=Access Key ID,description=Optional static source credential when runtime IAM is unavailable."`
	// SecretAccessKey is the AWS secret access key for static credentials
	SecretAccessKey string `json:"secretAccessKey,omitempty" jsonschema:"title=Secret Access Key"`
	// SessionToken is the AWS session token for static credentials
	SessionToken string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}
