package corejobs

import (
	"context"
	"testing"
	"time"
)

func TestCreateTaskJob_Generic(t *testing.T) {
	job := &CreateTaskJob{Client: &MockOpenlaneClient{}}

	err := job.Run(context.Background(), CreateTaskJobArgs{
		Type:           "Generic",
		Title:          "Test Task",
		Description:    "This is a generic test task",
		OrganizationID: "org123",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateTaskJob_PolicyReview(t *testing.T) {
	job := &CreateTaskJob{Client: &MockOpenlaneClient{}}

	err := job.Run(context.Background(), CreateTaskJobArgs{
		Type:              "Policy Review",
		InternalPolicyIDs: []string{"policy123"},
		OrganizationID:    "org123",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateTaskJob_WithDelay(t *testing.T) {
	job := &CreateTaskJob{Client: &MockOpenlaneClient{}}

	err := job.Run(context.Background(), CreateTaskJobArgs{
		Type:           "Generic",
		Title:          "Delayed Task",
		OrganizationID: "org123",
		Delay:          1 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
}
