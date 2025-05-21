package enums

import (
	"fmt"
	"io"
	"strings"
)

// JobExecutionStatus is a custom type representing the various states of JobExecutionStatus.
type JobExecutionStatus string

var (
	// JobExecutionStatusCanceled indicates the canceled.
	JobExecutionStatusCanceled JobExecutionStatus = "CANCELED"
	// JobExecutionStatusSuccess indicates the success.
	JobExecutionStatusSuccess JobExecutionStatus = "SUCCESS"
	// JobExecutionStatusPending indicates the pending.
	JobExecutionStatusPending JobExecutionStatus = "PENDING"
	// JobExecutionStatusFailed indicates the failed.
	JobExecutionStatusFailed JobExecutionStatus = "FAILED"
	// JobExecutionStatusInvalid is used when an unknown or unsupported value is provided.
	JobExecutionStatusInvalid JobExecutionStatus = "JOBEXECUTIONSTATUS_INVALID"
)

// Values returns a slice of strings representing all valid JobExecutionStatus values.
func (JobExecutionStatus) Values() []string {
	return []string{
		string(JobExecutionStatusCanceled),
		string(JobExecutionStatusSuccess),
		string(JobExecutionStatusPending),
		string(JobExecutionStatusFailed),
	}
}

// String returns the string representation of the JobExecutionStatus value.
func (r JobExecutionStatus) String() string {
	return string(r)
}

// ToJobExecutionStatus converts a string to its corresponding JobExecutionStatus enum value.
func ToJobExecutionStatus(r string) *JobExecutionStatus {
	switch strings.ToUpper(r) {
	case JobExecutionStatusCanceled.String():
		return &JobExecutionStatusCanceled
	case JobExecutionStatusSuccess.String():
		return &JobExecutionStatusSuccess
	case JobExecutionStatusPending.String():
		return &JobExecutionStatusPending
	case JobExecutionStatusFailed.String():
		return &JobExecutionStatusFailed
	default:
		return &JobExecutionStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobExecutionStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobExecutionStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobExecutionStatus, got: %T", v) //nolint:err113
	}

	*r = JobExecutionStatus(str)

	return nil
}
