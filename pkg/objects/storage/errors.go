package storage

import "errors"

var (
	// ErrInvalidS3Bucket is returned when an invalid s3 bucket is provided
	ErrInvalidS3Bucket = errors.New("invalid s3 bucket provided")
	// ErrInvalidFolderPath is returned when an invalid folder path is provided
	ErrInvalidFolderPath = errors.New("invalid folder path provided")
	// ErrMissingRequiredAWSParams is returned when required AWS parameters are missing
	ErrMissingRequiredAWSParams = errors.New("missing required AWS parameters")
)
