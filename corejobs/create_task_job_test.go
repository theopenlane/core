package corejobs

import (
	"context"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// --- Mock Openlane Client ---

type MockOpenlaneClient struct{}

func (m *MockOpenlaneClient) CreateTask(ctx context.Context, input openlaneclient.CreateTaskInput) (*openlaneclient.Task, error) {
	return &openlaneclient.Task{
		ID:    "mock_task_id",
		Title: input.Title,
	}, nil
}

func (m *MockOpenlaneClient) GetInternalPolicy(ctx context.Context, policyID string) (*openlaneclient.InternalPolicy, error) {
	return &openlaneclient.InternalPolicy{
		ID:          policyID,
		Description: "Mock policy description",
		ApproversGroup: openlaneclient.UserGroup{
			Users: []openlaneclient.User{
				{ID: "user1"},
				{ID: "user2"},
			},
		},
	}, nil
}

// --- Tests ---

func TestCreateTaskWorker_Generic(t *testing.T) {
	worker := &CreateTaskWorker{olClient: &MockOpenlaneClient{}}

	args := CreateTaskArgs{
		Type: TaskTypeGeneric,
		GenericConfig: &GenericTaskConfig{
			TaskConfig:  TaskConfig{OrganizationID: "org123"},
			Title:       "Test Generic Task",
			Description: "This is a generic test task",
			Category:    "Testing",
		},
	}

	job := &river.Job[CreateTaskArgs]{Args: args}
	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTaskWorker_PolicyReview(t *testing.T) {
	worker := &CreateTaskWorker{olClient: &MockOpenlaneClient{}}

	args := CreateTaskArgs{
		Type: TaskTypePolicyReview,
		PolicyReviewConfig: &PolicyReviewConfig{
			TaskConfig:        TaskConfig{OrganizationID: "org123"},
			InternalPolicyIDs: []string{"policy123"},
		},
	}

	job := &river.Job[CreateTaskArgs]{Args: args}
	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTaskWorker_WithDelay(t *testing.T) {
	worker := &CreateTaskWorker{olClient: &MockOpenlaneClient{}, riverClient: &MockRiverClient{}}

	args := CreateTaskArgs{
		Type: TaskTypeGeneric,
		GenericConfig: &GenericTaskConfig{
			TaskConfig: TaskConfig{
				OrganizationID: "org123",
				Delay:          1 * time.Second,
			},
			Title:       "Delayed Task",
			Description: "This task is delayed",
		},
	}

	job := &river.Job[CreateTaskArgs]{Args: args}
	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatal(err)
	}
}

// --- Mock River client for scheduled jobs ---

type MockRiverClient struct{}

func (m *MockRiverClient) Insert(ctx context.Context, args interface{}, opts *river.InsertOpts) (*river.Job[interface{}], error) {
	return &river.Job[interface{}]{}, nil
}
