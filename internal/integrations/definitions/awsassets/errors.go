package awsassets

import "errors"

var (
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("awsassets: credential metadata required")
	// ErrCredentialMetadataInvalid indicates credential metadata could not be decoded
	ErrCredentialMetadataInvalid = errors.New("awsassets: credential metadata invalid")
	// ErrRoleARNMissing indicates the IAM role ARN is missing
	ErrRoleARNMissing = errors.New("awsassets: role ARN required")
	// ErrAWSConfigBuildFailed indicates the AWS SDK config could not be constructed
	ErrAWSConfigBuildFailed = errors.New("awsassets: aws config build failed")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("awsassets: unexpected client type")
	// ErrCallerIdentityLookupFailed indicates STS GetCallerIdentity failed
	ErrCallerIdentityLookupFailed = errors.New("awsassets: caller identity lookup failed")
	// ErrIdentityVerificationFailed indicates asset collection identity verification failed
	ErrIdentityVerificationFailed = errors.New("awsassets: identity verification failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("awsassets: result encode failed")
)
