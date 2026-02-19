package enums

import "io"

// JobRunnerStatus is a custom type representing the various states of JobRunnerStatus.
type JobRunnerStatus string

var (
	// JobRunnerStatusOnline indicates the online.
	JobRunnerStatusOnline JobRunnerStatus = "ONLINE"
	// JobRunnerStatusOffline indicates the offline.
	JobRunnerStatusOffline JobRunnerStatus = "OFFLINE"
	// JobRunnerStatusInvalid is used when an unknown or unsupported value is provided.
	JobRunnerStatusInvalid JobRunnerStatus = "JOBRUNNERSTATUS_INVALID"
)

var jobRunnerStatusValues = []JobRunnerStatus{JobRunnerStatusOnline, JobRunnerStatusOffline}

// Values returns a slice of strings representing all valid JobRunnerStatus values.
func (JobRunnerStatus) Values() []string { return stringValues(jobRunnerStatusValues) }

// String returns the string representation of the JobRunnerStatus value.
func (r JobRunnerStatus) String() string { return string(r) }

// ToJobRunnerStatus converts a string to its corresponding JobRunnerStatus enum value.
func ToJobRunnerStatus(r string) *JobRunnerStatus {
	return parse(r, jobRunnerStatusValues, &JobRunnerStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobRunnerStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobRunnerStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
