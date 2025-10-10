package corejobs

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// --- Mock client for Openlane API ---
type OpenlaneClient interface {
	CreateTask(ctx context.Context, input CreateTaskInput) error
}

type CreateTaskInput struct {
	Title          string
	Description    string
	Category       string
	Assignee       string
	OrganizationID string
	PolicyIDs      []string
}

// --- Mock implementation ---
type MockOpenlaneClient struct{}

func (m *MockOpenlaneClient) CreateTask(ctx context.Context, input CreateTaskInput) error {
	fmt.Println("Mock task created:", input)
	return nil
}

// --- Job args ---
type CreateTaskJobArgs struct {
	Type              string
	Title             string
	Description       string
	Category          string
	Assignee          string
	OrganizationID    string
	InternalPolicyIDs []string
	Delay             time.Duration
}

// --- The job struct ---
type CreateTaskJob struct {
	Client OpenlaneClient
}

// --- Job execution ---
func (j *CreateTaskJob) Run(ctx context.Context, args CreateTaskJobArgs) error {
	// Optional delay
	if args.Delay > 0 {
		fmt.Println("Delaying job for:", args.Delay)
		time.Sleep(args.Delay)
	}

	var task CreateTaskInput

	switch args.Type {
	case "Generic":
		task = CreateTaskInput{
			Title:          args.Title,
			Description:    args.Description,
			Category:       args.Category,
			Assignee:       args.Assignee,
			OrganizationID: args.OrganizationID,
		}
	case "Policy Review":
		if len(args.InternalPolicyIDs) == 0 {
			return fmt.Errorf("Policy Review requires InternalPolicyIDs")
		}
		// Random assignee mock
		approvers := []string{"user1", "user2", "user3"}
		randomAssignee := approvers[rand.Intn(len(approvers))]

		task = CreateTaskInput{
			Title:          fmt.Sprintf("Policy Review %s", args.InternalPolicyIDs[0]),
			Description:    "Conduct the annual review of this internal policy to ensure it remains accurate, effective, and aligned with current business practices, legal requirements, and compliance frameworks. Review and update the content as needed, obtain necessary approvals, and document any changes or confirmations.",
			Category:       "Policy Review",
			Assignee:       randomAssignee,
			OrganizationID: args.OrganizationID,
			PolicyIDs:      args.InternalPolicyIDs,
		}
	default:
		return fmt.Errorf("unsupported task type: %s", args.Type)
	}

	return j.Client.CreateTask(ctx, task)
}
