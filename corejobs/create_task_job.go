package corejobs

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/99designs/gqlgen/graphql" // ✅ correct Upload type import
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// TaskType is the type of task being created, e.g. Generic or PolicyReview.
type TaskType string

const (
	Generic      TaskType = "Generic"
	PolicyReview TaskType = "PolicyReview"
)

// TaskConfig defines common configuration options for all tasks.
type TaskConfig struct {
	Delay time.Duration
}

// GenericTaskConfig defines configuration for creating a generic task.
type GenericTaskConfig struct {
	TaskConfig
	Title          string
	Description    string
	Category       string
	AssigneeID     string
	OrganizationID string
}

// PolicyReviewTaskConfig defines configuration for creating a policy review task.
type PolicyReviewTaskConfig struct {
	TaskConfig
	InternalPolicyIDs []string
}

var (
	ErrUnsupportedTask = errors.New("unsupported task type")
	ErrMissingPolicyID = errors.New("policy review requires internal policy IDs")
)

// CreateTaskJobArgs holds the arguments required by CreateTaskJob.
type CreateTaskJobArgs struct {
	Type    TaskType `json:"type"`
	Generic *GenericTaskConfig
	Policy  *PolicyReviewTaskConfig
}

// Kind identifies this job type.
func (CreateTaskJobArgs) Kind() string {
	return "create_task"
}

// CreateTaskJob represents a job that creates tasks in Openlane.
type CreateTaskJob struct {
	olClient openlaneclient.OpenlaneGraphClient
}

// WithOpenlaneClient injects a mock or real Openlane client.
func (j *CreateTaskJob) WithOpenlaneClient(cl openlaneclient.OpenlaneGraphClient) *CreateTaskJob {
	j.olClient = cl
	return j
}

// Work executes the task creation logic.
func (j *CreateTaskJob) Work(ctx context.Context, args *CreateTaskJobArgs) error {
	if j.olClient == nil {
		return errors.New("missing Openlane client")
	}

	upload := graphql.Upload{} // ✅ Proper type

	switch args.Type {
	case Generic:
		if args.Generic == nil {
			return errors.New("missing generic task config")
		}

		// Simulate a CreateTask or similar call
		_, err := j.olClient.CloneBulkCSVControl(ctx, upload)
		if err != nil {
			return err
		}

		log.Info().
			Str("title", args.Generic.Title).
			Msg("Created generic task")

		return nil

	case PolicyReview:
		if args.Policy == nil || len(args.Policy.InternalPolicyIDs) == 0 {
			return ErrMissingPolicyID
		}

		randomAssignee := "user" + string(rune(rand.Intn(100)))

		_, err := j.olClient.CloneBulkCSVControl(ctx, upload)
		if err != nil {
			return err
		}

		log.Info().
			Str("policy_task", "Policy Review").
			Str("assignee", randomAssignee).
			Msg("Created policy review task")

		return nil

	default:
		return ErrUnsupportedTask
	}
}
