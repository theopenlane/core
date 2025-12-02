//go:build examples

package common

import "errors"

var (
	// ErrTimeoutWaitingForMinIO is returned when MinIO fails to become ready within the timeout period
	ErrTimeoutWaitingForMinIO = errors.New("timeout waiting for MinIO")

	// ErrTimeoutWaitingForGCS is returned when fake GCS fails to become ready within the timeout period
	ErrTimeoutWaitingForGCS = errors.New("timeout waiting for fake GCS")

	// ErrEmptyBucketName is returned when an empty bucket name is provided
	ErrEmptyBucketName = errors.New("bucket name cannot be empty")
)
