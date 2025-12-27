package enums

import (
	"fmt"
	"io"
	"strings"
)

type TaskStatus string

var (
	TaskStatusOpen       TaskStatus = "OPEN"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusInReview   TaskStatus = "IN_REVIEW"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusWontDo     TaskStatus = "WONT_DO"
	TaskStatusInvalid    TaskStatus = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the TaskStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS", "IN_REVIEW", "COMPLETED", and "WONT_DO".
func (TaskStatus) Values() (kinds []string) {
	for _, s := range []TaskStatus{TaskStatusOpen, TaskStatusInProgress, TaskStatusInReview, TaskStatusCompleted, TaskStatusWontDo} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the TaskStatus as a string
func (r TaskStatus) String() string {
	return string(r)
}

// ToTaskStatus returns the task status enum based on string input
func ToTaskStatus(r string) *TaskStatus {
	switch r := strings.ToUpper(r); r {
	case TaskStatusOpen.String():
		return &TaskStatusOpen
	case TaskStatusInProgress.String():
		return &TaskStatusInProgress
	case TaskStatusInReview.String():
		return &TaskStatusInReview
	case TaskStatusCompleted.String():
		return &TaskStatusCompleted
	case TaskStatusWontDo.String():
		return &TaskStatusWontDo
	default:
		return &TaskStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TaskStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TaskStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TaskStatus, got: %T", v) //nolint:err113
	}

	*r = TaskStatus(str)

	return nil
}
