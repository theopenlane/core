package jobspec

import (
	"time"

	"github.com/riverqueue/river"

	"github.com/theopenlane/core/common/models"
)

// CreateTaskArgs for the worker to process
// This is a generic task creation job that accepts pre-built task details.
// External systems (schedulers, triggers, polling jobs) should handle business logic
// like fetching policies, selecting assignees, building descriptions, etc.
type CreateTaskArgs struct {
	// OrganizationID is the organization that owns the task (required)
	OrganizationID string `json:"organization_id"`

	// Title of the task (required)
	Title string `json:"title"`

	// Description/details of the task (required)
	Description string `json:"description"`

	// Category of the task (optional) - e.g. "Policy Review", "Evidence Upload", "Onboarding"
	Category *string `json:"category,omitempty"`

	// AssigneeID is the user to assign the task to (optional)
	AssigneeID *string `json:"assignee_id,omitempty"`

	// AssignerID is the user who created/assigned the task (optional)
	AssignerID *string `json:"assigner_id,omitempty"`

	// DueDate for the task (optional)
	DueDate *models.DateTime `json:"due_date,omitempty"`

	// InternalPolicyIDs to link the task to internal policies (optional)
	InternalPolicyIDs []string `json:"internal_policy_ids,omitempty"`

	// Tags associated with the task (optional)
	Tags []string `json:"tags,omitempty"`

	// ScheduledAt allows scheduling the job for a future time (optional)
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// Kind satisfies the river.Job interface
func (CreateTaskArgs) Kind() string { return "create_task" }

// InsertOpts provides the default configuration when processing this job.
func (CreateTaskArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueDefault}
}
