package enums

import "io"

type TaskStatus string

var (
	TaskStatusOpen       TaskStatus = "OPEN"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusInReview   TaskStatus = "IN_REVIEW"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusWontDo     TaskStatus = "WONT_DO"
	TaskStatusInvalid    TaskStatus = "INVALID"
)

var taskStatusValues = []TaskStatus{TaskStatusOpen, TaskStatusInProgress, TaskStatusInReview, TaskStatusCompleted, TaskStatusWontDo}

// Values returns a slice of strings that represents all the possible values of the TaskStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS", "IN_REVIEW", "COMPLETED", and "WONT_DO".
func (TaskStatus) Values() []string { return stringValues(taskStatusValues) }

// String returns the TaskStatus as a string
func (r TaskStatus) String() string { return string(r) }

// ToTaskStatus returns the task status enum based on string input
func ToTaskStatus(r string) *TaskStatus { return parse(r, taskStatusValues, &TaskStatusInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TaskStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TaskStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
