package enums

import (
	"fmt"
	"io"
	"strings"
)

// JobCadenceFrequency is a custom type representing the various states of JobCadenceFrequency.
type JobCadenceFrequency string

var (
	// JobCadenceFrequencyDaily indicates the daily.
	JobCadenceFrequencyDaily JobCadenceFrequency = "DAILY"
	// JobCadenceFrequencyWeekly indicates the weekly.
	JobCadenceFrequencyWeekly JobCadenceFrequency = "WEEKLY"
	// JobCadenceFrequencyMonthly indicates the monthly.
	JobCadenceFrequencyMonthly JobCadenceFrequency = "MONTHLY"
	// JobCadenceFrequencyInvalid is used when an unknown or unsupported value is provided.
	JobCadenceFrequencyInvalid JobCadenceFrequency = "JOBCADENCEFREQUENCY_INVALID"
)

// Values returns a slice of strings representing all valid JobCadenceFrequency values.
func (JobCadenceFrequency) Values() []string {
	return []string{
		string(JobCadenceFrequencyDaily),
		string(JobCadenceFrequencyWeekly),
		string(JobCadenceFrequencyMonthly),
	}
}

// String returns the string representation of the JobCadenceFrequency value.
func (r JobCadenceFrequency) String() string {
	return strings.ToUpper(string(r))
}

// ToJobCadenceFrequency converts a string to its corresponding JobCadenceFrequency enum value.
func ToJobCadenceFrequency(r string) *JobCadenceFrequency {
	switch strings.ToUpper(r) {
	case JobCadenceFrequencyDaily.String():
		return &JobCadenceFrequencyDaily
	case JobCadenceFrequencyWeekly.String():
		return &JobCadenceFrequencyWeekly
	case JobCadenceFrequencyMonthly.String():
		return &JobCadenceFrequencyMonthly
	default:
		return &JobCadenceFrequencyInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobCadenceFrequency) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobCadenceFrequency) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobCadenceFrequency, got: %T", v) //nolint:err113
	}

	*r = JobCadenceFrequency(str)

	return nil
}
