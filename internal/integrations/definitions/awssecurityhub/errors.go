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
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("awssecurityhub: operation config invalid")
	// ErrListAssessmentsFailed indicates ListAssessments failed
	ErrListAssessmentsFailed = errors.New("awssecurityhub: list assessments failed")
	// ErrFindingsFetchFailed indicates GetFindings failed
	ErrFindingsFetchFailed = errors.New("awssecurityhub: findings fetch failed")
	// ErrFindingEncode indicates a finding payload could not be serialized
	ErrFindingEncode = errors.New("awssecurityhub: finding encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("awssecurityhub: result encode failed")
)
