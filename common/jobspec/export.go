package jobspec

import "github.com/riverqueue/river"

// ExportContentArgs for the worker to process and update the record for the updated content
type ExportContentArgs struct {
	// ExportID is the ID of the export job
	ExportID string `json:"export_id,omitempty"`
	// UserID of the user who requested the export (for system admin context)
	UserID string `json:"user_id,omitempty"`
	// OrganizationID of the organization context for the export
	OrganizationID string `json:"organization_id,omitempty"`
}

// Kind satisfies the river.Job interface
func (ExportContentArgs) Kind() string { return "export_content" }

// InsertOpts provides the default configuration when processing this job.
func (ExportContentArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueDefault}
}
