package awsauditmanager

import "errors"

var (
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("awsauditmanager: unexpected client type")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("awsauditmanager: credential metadata required")
	// ErrRoleARNMissing indicates the IAM role ARN is missing from the credential
	ErrRoleARNMissing = errors.New("awsauditmanager: roleArn required")
	// ErrRegionMissing indicates the home region is missing from the credential
	ErrRegionMissing = errors.New("awsauditmanager: homeRegion required")
)
