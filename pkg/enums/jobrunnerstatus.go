package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings representing all valid JobRunnerStatus values.
func (JobRunnerStatus) Values() []string {
	return []string{
		string(JobRunnerStatusOnline),
		string(JobRunnerStatusOffline),
	}
}

// String returns the string representation of the JobRunnerStatus value.
func (r JobRunnerStatus) String() string {
	return string(r)
}

// ToJobRunnerStatus converts a string to its corresponding JobRunnerStatus enum value.
func ToJobRunnerStatus(r string) *JobRunnerStatus {
	switch strings.ToUpper(r) {
	case JobRunnerStatusOnline.String():
		return &JobRunnerStatusOnline
	case JobRunnerStatusOffline.String():
		return &JobRunnerStatusOffline
	default:
		return &JobRunnerStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobRunnerStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobRunnerStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobRunnerStatus, got: %T", v) //nolint:err113
	}

	*r = JobRunnerStatus(str)

	return nil
}
