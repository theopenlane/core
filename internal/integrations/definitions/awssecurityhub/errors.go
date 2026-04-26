package awssecurityhub

import "errors"

var (
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("awssecurityhub: credential metadata required")
	// ErrCredentialMetadataInvalid indicates credential metadata could not be decoded
	ErrCredentialMetadataInvalid = errors.New("awssecurityhub: credential metadata invalid")
	// ErrRoleARNMissing indicates the IAM role ARN is missing from the credential
	ErrRoleARNMissing = errors.New("awssecurityhub: roleArn required")
	// ErrRegionMissing indicates the home region is missing from the credential
	ErrRegionMissing = errors.New("awssecurityhub: homeRegion required")
	// ErrAWSConfigBuildFailed indicates the AWS SDK config could not be constructed
	ErrAWSConfigBuildFailed = errors.New("awssecurityhub: aws config build failed")
	// ErrDescribeHubFailed indicates DescribeHub failed
	ErrDescribeHubFailed = errors.New("awssecurityhub: describe hub failed")
	// ErrDescribeHubFailed indicates DescribeHub failed
	ErrSecurityHubNotEnabled = errors.New("awssecurityhub: security hub not enabled for account")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("awssecurityhub: operation config invalid")
	// ErrListAssessmentsFailed indicates ListAssessments failed
	ErrListAssessmentsFailed = errors.New("awssecurityhub: list assessments failed")
	// ErrAssessmentEncode indicates an assessment payload could not be serialized for ingest
	ErrAssessmentEncode = errors.New("awssecurityhub: assessment encode failed")
	// ErrFindingsFetchFailed indicates GetFindings failed
	ErrFindingsFetchFailed = errors.New("awssecurityhub: findings fetch failed")
	// ErrFindingEncode indicates a finding payload could not be serialized
	ErrFindingEncode = errors.New("awssecurityhub: finding encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("awssecurityhub: result encode failed")
	// ErrConfigRulesFetchFailed indicates DescribeConfigRules failed
	ErrConfigRulesFetchFailed = errors.New("awsconfig: config rules fetch failed")
	// ErrConfigControlEncode indicates a config rule could not be serialized for ingest
	ErrConfigControlEncode = errors.New("awsconfig: config control encode failed")
	// ErrControlCatalogFetchFailed indicates ListControls failed
	ErrControlCatalogFetchFailed = errors.New("awsconfig: control catalog fetch failed")
	// ErrCatalogControlEncode indicates a control catalog entry could not be serialized for ingest
	ErrCatalogControlEncode = errors.New("awsconfig: catalog control encode failed")
	// ErrIAMUsersFetchFailed indicates ListUsers failed
	ErrIAMUsersFetchFailed = errors.New("awsiam: IAM users fetch failed")
	// ErrIAMGroupsFetchFailed indicates ListGroups failed
	ErrIAMGroupsFetchFailed = errors.New("awsiam: IAM groups fetch failed")
	// ErrIAMGroupsForUserFetchFailed indicates ListGroupsForUser failed
	ErrIAMGroupsForUserFetchFailed = errors.New("awsiam: IAM groups for user fetch failed")
	// ErrDirectorySyncPayloadEncode indicates a directory sync payload could not be serialized for ingest
	ErrDirectorySyncPayloadEncode = errors.New("awsiam: directory sync payload encode failed")
)
