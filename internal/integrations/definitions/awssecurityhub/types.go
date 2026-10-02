package awssecurityhub

import (
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the AWS Security Hub integration definition
	definitionID = types.NewDefinitionRef("def_01K0AWSSECHUB0000000000001")
	// installation is the typed installation metadata handle for the AWS Security Hub definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// awsAssumeRoleCredential is the typed credential slot for AWS STS assume-role auth
	awsAssumeRoleCredential = types.CredentialRefOf[AssumeRoleCredentialSchema]()
	// awsServiceAccountCredential is the credential slot for static service account credentials
	awsServiceAccountCredential = types.CredentialRefOf[ServiceAccountCredentialSchema]()
	// securityHubClient is the client ref for the AWS Security Hub client used by this definition
	securityHubClient = types.ClientRefOf[*securityhub.Client]()
	// configServiceClient is the client ref for the AWS Config client
	configServiceClient = types.ClientRefOf[*configservice.Client]()
	// iamClient is the client ref for the AWS IAM client used by directory sync operations
	iamClient = types.ClientRefOf[*iam.Client]()
)

// DirectorySync are the configuration settings for the directory sync from AWS IAM
type DirectorySync struct {
	// DisableGroupSync will just sync users and no groups or group memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from AWS IAM, disable groups sync operations"`
}

// FindingSync are configuration settings for the findings sync
type FindingSync struct{}

// CheckSync are the configuration settings for the check sync from AWS Config
type CheckSync struct{}

// AssetSync are the configuration settings for the asset sync
type AssetSync struct{}

// AssumeRoleCredentialSchema holds the AWS assume-role and collection-scope inputs
type AssumeRoleCredentialSchema struct {
	// RoleARN is the cross-account IAM role ARN Openlane should assume in the tenant environment
	RoleARN string `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN,description=Cross-account role Openlane should assume in the tenant environment.,secret=true"`
	// ExternalID is the external ID required in the tenant role trust policy
	ExternalID string `json:"externalId"                jsonschema:"required,title=External ID,description=External ID required in the tenant role trust policy." jsonschema_extras:"generate=true"`
	// HomeRegion is the AWS region where Security Hub cross-region aggregation is managed
	HomeRegion string `json:"homeRegion"                jsonschema:"required,title=Home Region,description=AWS region used for Security Hub aggregation and other service API calls (e.g. us-east-1)."`
	// AccountID is the AWS account ID, derived from RoleARN when not provided
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

// ServiceAccountCredentialSchema is the service account based credential schema
type ServiceAccountCredentialSchema struct {
	// AccessKeyID is an service account credential when runtime IAM is unavailable
	AccessKeyID string `json:"accessKeyId"     jsonschema:"required,title=Access Key ID,description=Static source credential used when runtime IAM is unavailable."`
	// SecretAccessKey is the AWS secret access key for static credentials
	SecretAccessKey string `json:"secretAccessKey" jsonschema:"required,title=Secret Access Key"`
	// SessionToken is the AWS session token for static credentials
	SessionToken string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
	// Region is the AWS region for API calls (e.g. us-east-1)
	Region string `json:"region" jsonschema:"required,title=Region,description=AWS region used for Security Hub and other service API calls (e.g. us-east-1)."`
}

// InstallationMetadata holds the non-secret AWS connection attributes for an installation
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
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.AccountID,
	}
}
