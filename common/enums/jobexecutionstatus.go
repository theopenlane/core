package enums

import "io"

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

var jobExecutionStatusValues = []JobExecutionStatus{
	JobExecutionStatusCanceled,
	JobExecutionStatusSuccess,
	JobExecutionStatusPending,
	JobExecutionStatusFailed,
}

// Values returns a slice of strings representing all valid JobExecutionStatus values.
func (JobExecutionStatus) Values() []string { return stringValues(jobExecutionStatusValues) }

// String returns the string representation of the JobExecutionStatus value.
func (r JobExecutionStatus) String() string { return string(r) }

// ToJobExecutionStatus converts a string to its corresponding JobExecutionStatus enum value.
func ToJobExecutionStatus(r string) *JobExecutionStatus {
	return parse(r, jobExecutionStatusValues, &JobExecutionStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobExecutionStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobExecutionStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
