package awsauditmanager

import "errors"

var (
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("awsauditmanager: unexpected client type")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("awsauditmanager: credential metadata required")
	// ErrCredentialMetadataInvalid indicates credential metadata could not be decoded
	ErrCredentialMetadataInvalid = errors.New("awsauditmanager: credential metadata invalid")
	// ErrRoleARNMissing indicates the IAM role ARN is missing from the credential
	ErrRoleARNMissing = errors.New("awsauditmanager: roleArn required")
	// ErrRegionMissing indicates the home region is missing from the credential
	ErrRegionMissing = errors.New("awsauditmanager: homeRegion required")
	// ErrAWSConfigBuildFailed indicates the AWS SDK config could not be constructed
	ErrAWSConfigBuildFailed = errors.New("awsauditmanager: aws config build failed")
	// ErrGetAccountStatusFailed indicates GetAccountStatus failed
	ErrGetAccountStatusFailed = errors.New("awsauditmanager: get account status failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("awsauditmanager: operation config invalid")
	// ErrListAssessmentsFailed indicates ListAssessments failed
	ErrListAssessmentsFailed = errors.New("awsauditmanager: list assessments failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("awsauditmanager: result encode failed")
)
