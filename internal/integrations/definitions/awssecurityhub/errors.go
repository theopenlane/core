package awssecurityhub

import "errors"

var (
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("awssecurityhub: unexpected client type")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("awssecurityhub: credential metadata required")
	// ErrRoleARNMissing indicates the IAM role ARN is missing from the credential
	ErrRoleARNMissing = errors.New("awssecurityhub: roleArn required")
	// ErrRegionMissing indicates the home region is missing from the credential
	ErrRegionMissing = errors.New("awssecurityhub: homeRegion required")
)
