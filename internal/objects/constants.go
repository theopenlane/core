package objects

import "time"

const (
	// DefaultClientPoolTTL is the default TTL for client pool entries.
	DefaultClientPoolTTL = 15 * time.Minute
	// DefaultDevStorageBucket is used when a development storage bucket is required but not provided.
	DefaultDevStorageBucket = "/tmp/dev-storage"
	// DefaultS3Region is used when no region is specified for S3-compatible providers.
	DefaultS3Region = "us-east-1"
	// DefaultLocalDiskURL is the default local URL for disk storage providers.
	DefaultLocalDiskURL = "http://localhost:17608/files"
)
