package corejobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/models"
	openlaneclient "github.com/theopenlane/core/pkg/openlaneclient"
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

// InsertOpts configures job insertion options including retry behavior and scheduling
func (a CreateTaskArgs) InsertOpts() river.InsertOpts {
	opts := river.InsertOpts{
		MaxAttempts: 3,
	}

	// If scheduled time is specified, set it in the options
	if a.ScheduledAt != nil {
		opts.ScheduledAt = *a.ScheduledAt
	}

	return opts
}

// TaskWorkerConfig contains configuration for the create task worker
type TaskWorkerConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required,description=whether the task worker is enabled"`
}

// CreateTaskWorker processes create task jobs
type CreateTaskWorker struct {
	river.WorkerDefaults[CreateTaskArgs]

	Config TaskWorkerConfig `koanf:"config" json:"config"`

	olClient olclient.OpenlaneClient
}

// WithOpenlaneClient sets the Openlane client (used for testing)
func (w *CreateTaskWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *CreateTaskWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface
func (w *CreateTaskWorker) Work(ctx context.Context, job *river.Job[CreateTaskArgs]) error {
	logger := log.With().
		Str("organization_id", job.Args.OrganizationID).
		Str("task_title", job.Args.Title).
		Logger()

	logger.Info().Msg("starting create task job")

	// Validate required fields
	if err := w.validateArgs(job.Args); err != nil {
		logger.Error().Err(err).Msg("invalid job arguments")
		return err
	}

	// Initialize Openlane client if not already set
	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			logger.Error().Err(err).Msg("failed to create openlane client")
			return err
		}
		w.olClient = cl
	}

	// Build task input from provided arguments
	taskInput := openlaneclient.CreateTaskInput{
		Title:   job.Args.Title,
		Details: &job.Args.Description,
		OwnerID: &job.Args.OrganizationID,
	}

	// Add optional fields if provided
	if job.Args.Category != nil {
		taskInput.Category = job.Args.Category
	}

	if job.Args.AssigneeID != nil {
		taskInput.AssigneeID = job.Args.AssigneeID
	}

	if job.Args.AssignerID != nil {
		taskInput.AssignerID = job.Args.AssignerID
	}

	if job.Args.DueDate != nil {
		taskInput.Due = job.Args.DueDate
	}

	if len(job.Args.InternalPolicyIDs) > 0 {
		taskInput.InternalPolicyIDs = job.Args.InternalPolicyIDs
	}

	if len(job.Args.Tags) > 0 {
		taskInput.Tags = job.Args.Tags
	}

	// Create the task using Openlane client
	createdTask, err := w.olClient.CreateTask(ctx, taskInput)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create task")
		return err
	}

	logger.Info().
		Str("task_id", createdTask.CreateTask.Task.ID).
		Str("task_title", createdTask.CreateTask.Task.Title).
		Msg("task created successfully")

	return nil
}

// validateArgs validates the job arguments
func (w *CreateTaskWorker) validateArgs(args CreateTaskArgs) error {
	// Validate required fields
	if args.OrganizationID == "" {
		return newMissingRequiredArg("organization_id", CreateTaskArgs{}.Kind())
	}

	if args.Title == "" {
		return newMissingRequiredArg("title", CreateTaskArgs{}.Kind())
	}

	if args.Description == "" {
		return newMissingRequiredArg("description", CreateTaskArgs{}.Kind())
	}

	return nil
}
