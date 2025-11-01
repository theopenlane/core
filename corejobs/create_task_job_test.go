package corejobs

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/mocks"
)

func TestCreateTaskJob_Generic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOpenlaneGraphClient(ctrl)
	mockClient.EXPECT().
		CloneBulkCSVControl(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil)

	job := &CreateTaskJob{}
	job.WithOpenlaneClient(mockClient)

	args := &CreateTaskJobArgs{
		Type: Generic,
		Generic: &GenericTaskConfig{
			Title:       "Generic Task",
			Description: "This is a test",
		},
	}

	err := job.Work(context.Background(), args)
	require.NoError(t, err)
}

func TestCreateTaskJob_PolicyReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOpenlaneGraphClient(ctrl)
	mockClient.EXPECT().
		CloneBulkCSVControl(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil)

	job := &CreateTaskJob{}
	job.WithOpenlaneClient(mockClient)

	args := &CreateTaskJobArgs{
		Type: PolicyReview,
		Policy: &PolicyReviewTaskConfig{
			InternalPolicyIDs: []string{"policy123"},
		},
	}

	err := job.Work(context.Background(), args)
	require.NoError(t, err)
}
