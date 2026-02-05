package awssts

import "errors"

var (
	// ErrProviderMetadataRequired indicates provider metadata is required but not supplied
	ErrProviderMetadataRequired = errors.New("awssts: provider metadata required")
	// ErrRoleARNRequired indicates the roleArn field is missing from metadata
	ErrRoleARNRequired = errors.New("awssts: roleArn required")
	// ErrRegionRequired indicates the region field is missing from metadata
	ErrRegionRequired = errors.New("awssts: region required")
	// ErrAuthTypeMismatch indicates the provider spec specifies an incompatible auth type
	ErrAuthTypeMismatch = errors.New("awssts: auth type mismatch")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for AWS STS providers
	ErrBeginAuthNotSupported = errors.New("awssts: BeginAuth is not supported; configure credentials via metadata")
	// ErrProviderNotInitialized indicates the provider instance is nil
	ErrProviderNotInitialized = errors.New("awssts: provider not initialized")
)
