package awsassets

import "errors"

var (
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("awsassets: credential metadata required")
	// ErrRoleARNMissing indicates the IAM role ARN is missing
	ErrRoleARNMissing = errors.New("awsassets: role ARN required")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("awsassets: unexpected client type")
)
