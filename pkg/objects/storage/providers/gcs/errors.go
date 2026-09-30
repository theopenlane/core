package gcs

import "errors"

var (
	// ErrBucketRequired is returned when no GCS bucket is specified
	ErrBucketRequired = errors.New("gcs: bucket is required")
	// ErrProjectIDRequired is returned when listing buckets without a project id
	ErrProjectIDRequired = errors.New("gcs: project id is required")
	// ErrClientCreate is returned when the GCS client could not be created
	ErrClientCreate = errors.New("gcs: client creation failed")
)
