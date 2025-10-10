package r2

import "errors"

var (
	// ErrR2CredentialsRequired is returned when required R2 credentials are missing
	ErrR2CredentialsRequired = errors.New("missing required R2 credentials: bucket, account_id, access_key_id, secret_access_key")
	// ErrR2BucketRequired is returned when R2 bucket is not specified
	ErrR2BucketRequired = errors.New("R2 bucket is required")
	// ErrR2AccountIDRequired is returned when R2 account ID is not specified
	ErrR2AccountIDRequired = errors.New("R2 account ID is required")
	// ErrR2CredentialsMissing is returned when R2 access keys are missing
	ErrR2CredentialsMissing = errors.New("R2 access key ID and secret access key are required")
)
