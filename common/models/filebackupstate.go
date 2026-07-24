package models

import (
	"time"

	"github.com/theopenlane/core/common/enums"
)

// FileBackupState captures the internal backup replication state for a file, e.g. an R2
// object replicated to S3. It is stored as a dedicated JSONB column, kept out of the shared
// file metadata map, and is never exposed via the API.
type FileBackupState struct {
	// Status is the current backup replication status
	Status enums.FileBackupStatus `json:"status,omitempty"`
	// Provider is the destination provider the file was backed up to (e.g. s3)
	Provider string `json:"provider,omitempty"`
	// URI is the full URI of the backed up object at the destination
	URI string `json:"uri,omitempty"`
	// Bucket is the destination bucket the backup was written to
	Bucket string `json:"bucket,omitempty"`
	// CompletedAt is when the backup replication succeeded
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Attempts is the number of replication attempts made
	Attempts int `json:"attempts,omitempty"`
	// Error is the last replication error, if any
	Error string `json:"error,omitempty"`
}
