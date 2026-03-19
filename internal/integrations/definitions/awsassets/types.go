package awsassets

import (
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the AWS Assets integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0AWSAST00000000000000001")
	// AWSAssetsClient is the client ref for the AWS STS client used by this definition
	AWSAssetsClient = types.NewClientRef[*sts.Client]()
	// HealthDefaultOperation is the operation ref for the AWS Assets health check
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	// AssetCollectOperation is the operation ref for the AWS asset collection operation
	AssetCollectOperation = types.NewOperationRef[struct{}]("asset.collect")
)

// Slug is the unique identifier for the AWS Assets integration
const Slug = "aws_assets"

// OperatorConfig holds operator-owned defaults that apply across all AWS Assets installations
type OperatorConfig struct {
	// DefaultSessionDuration is the fallback STS session duration for new installations
	DefaultSessionDuration string `json:"defaultSessionDuration,omitempty" jsonschema:"title=Default Session Duration"`
	// DefaultExternalID is the fallback external ID for new installations
	DefaultExternalID string `json:"defaultExternalId,omitempty"      jsonschema:"title=Default External ID"`
	// DefaultRegions is the fallback list of AWS regions for new installations
	DefaultRegions []string `json:"defaultRegions,omitempty"         jsonschema:"title=Default Regions"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// AccountSelectors limits collection to accounts matching the provided selectors
	AccountSelectors []string `json:"accountSelectors,omitempty" jsonschema:"title=Account Selectors"`
	// RegionSelectors limits collection to regions matching the provided selectors
	RegionSelectors []string `json:"regionSelectors,omitempty"  jsonschema:"title=Region Selectors"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty"       jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// CredentialSchema holds the AWS role and optional static key material for one installation
type CredentialSchema struct {
	// RoleARN is the IAM role ARN to assume for this installation
	RoleARN string `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN"`
	// ExternalID is the external ID required for role assumption
	ExternalID string `json:"externalId"                jsonschema:"required,title=External ID"`
	// HomeRegion is the primary AWS region for this installation
	HomeRegion string `json:"homeRegion"                jsonschema:"required,title=Home Region"`
	// AccountID is the AWS account identifier for reference
	AccountID string `json:"accountId,omitempty"       jsonschema:"title=Account ID"`
	// AccessKeyID is an optional static AWS access key ID
	AccessKeyID string `json:"accessKeyId,omitempty"     jsonschema:"title=Access Key ID"`
	// SecretAccessKey is an optional static AWS secret access key
	SecretAccessKey string `json:"secretAccessKey,omitempty" jsonschema:"title=Secret Access Key"`
	// SessionToken is an optional static AWS session token
	SessionToken string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}
