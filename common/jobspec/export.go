package jobspec

import (
	"github.com/riverqueue/river"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
)

// ExportContentArgs for the worker to process and update the record for the updated content
type ExportContentArgs struct {
	// ExportID is the ID of the export job
	ExportID string `json:"export_id,omitempty"`
	// UserID of the user who requested the export (for system admin context)
	UserID string `json:"user_id,omitempty"`
	// OrganizationID of the organization context for the export
	OrganizationID string `json:"organization_id,omitempty"`
	// Mode is the export mode (e.g., flat or folder)
	Mode enums.ExportMode `json:"mode,omitempty"`
	// ExportMetadata contains additional metadata for the export
	ExportMetadata *models.ExportMetadata `json:"export_metadata,omitempty"`
}

// Kind satisfies the river.Job interface
func (ExportContentArgs) Kind() string { return "export_content" }

// InsertOpts provides the default configuration when processing this job.
func (ExportContentArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueDefault}
}
