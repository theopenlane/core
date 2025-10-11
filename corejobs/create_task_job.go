package corejobs

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/riverqueue/river"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
)

// --- Constants and Types ---

type TaskType string

const (
	TaskTypeGeneric      TaskType = "Generic"
	TaskTypePolicyReview TaskType = "PolicyReview"
)

var (
	ErrUnsupportedTask = errors.New("unsupported task type")
)

type TaskConfig struct {
	Delay          time.Duration
	OrganizationID string
}

type GenericTaskConfig struct {
	TaskConfig
	Title       string
	Description string
	Category    string
}

type PolicyReviewConfig struct {
	TaskConfig
	InternalPolicyIDs []string
}

type CreateTaskArgs struct {
	Type TaskType `json:"type"`

	GenericConfig      *GenericTaskConfig  `json:"generic_config,omitempty"`
	PolicyReviewConfig *PolicyReviewConfig `json:"policy_review_config,omitempty"`
}

// --- Worker ---

type CreateTaskWorker struct {
	river.WorkerDefaults[CreateTaskArgs]

	olClient    olclient.OpenlaneClient
	riverClient riverqueue.JobClient
}

// WithOpenlaneClient sets the Openlane client
func (w *CreateTaskWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *CreateTaskWorker {
	w.olClient = cl
	return w
}

// WithRiverClient sets the River client
func (w *CreateTaskWorker) WithRiverClient(cl riverqueue.JobClient) *CreateTaskWorker {
	w.riverClient = cl
	return w
}

// Work implements river.Worker interface
func (w *CreateTaskWorker) Work(ctx context.Context, job *river.Job[CreateTaskArgs]) error {
	log.Debug().Str("task_type", string(job.Args.Type)).Msg("starting task creation")

	if w.olClient == nil {
		cl, err := getOpenlaneClient(TaskConfig{}) // implement this to return your olClient
		if err != nil {
			return err
		}
		w.olClient = cl
	}

	if w.riverClient == nil {
		client, err := riverqueue.New(ctx)
		if err != nil {
			return err
		}
		w.riverClient = client
	}

	// Schedule job if Delay is set
	var delay time.Duration
	switch job.Args.Type {
	case TaskTypeGeneric:
		delay = job.Args.GenericConfig.Delay
	case TaskTypePolicyReview:
		delay = job.Args.PolicyReviewConfig.Delay
	}

	if delay > 0 {
		log.Debug().Dur("delay", delay).Msg("scheduling delayed task")
		_, err := w.riverClient.Insert(ctx, job.Args, &river.InsertOpts{
			RunAt: time.Now().Add(delay),
		})
		if err != nil {
			return err
		}
		return nil
	}

	// Task creation logic
	switch job.Args.Type {
	case TaskTypeGeneric:
		cfg := job.Args.GenericConfig
		if cfg == nil {
			return fmt.Errorf("%w: missing GenericTaskConfig", ErrUnsupportedTask)
		}
		task := openlaneclient.CreateTaskInput{
			Title:          cfg.Title,
			Description:    cfg.Description,
			Category:       cfg.Category,
			OrganizationID: cfg.OrganizationID,
		}
		_, err := w.olClient.CreateTask(ctx, task)
		if err != nil {
			return err
		}
		log.Info().Str("title", cfg.Title).Msg("Generic task created")

	case TaskTypePolicyReview:
		cfg := job.Args.PolicyReviewConfig
		if cfg == nil || len(cfg.InternalPolicyIDs) == 0 {
			return fmt.Errorf("%w: PolicyReview requires InternalPolicyIDs", ErrUnsupportedTask)
		}

		// Pick random approver from internal policy
		policy, err := w.olClient.GetInternalPolicy(ctx, cfg.InternalPolicyIDs[0])
		if err != nil {
			return err
		}
		if len(policy.ApproversGroup.Users) == 0 {
			return errors.New("policy has no approvers")
		}
		assignee := policy.ApproversGroup.Users[rand.Intn(len(policy.ApproversGroup.Users))]

		task := openlaneclient.CreateTaskInput{
			Title:          fmt.Sprintf("Policy Review %s", cfg.InternalPolicyIDs[0]),
			Description:    policy.Description,
			OrganizationID: cfg.OrganizationID,
			AssigneeID:     assignee.ID,
			PolicyIDs:      cfg.InternalPolicyIDs,
		}
		_, err = w.olClient.CreateTask(ctx, task)
		if err != nil {
			return err
		}
		log.Info().Str("policy_id", cfg.InternalPolicyIDs[0]).Str("assignee", assignee.ID).Msg("Policy review task created")

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedTask, job.Args.Type)
	}

	return nil
}
