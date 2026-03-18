package enums

import "io"

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

var jobCadenceFrequencyValues = []JobCadenceFrequency{JobCadenceFrequencyDaily, JobCadenceFrequencyWeekly, JobCadenceFrequencyMonthly}

// Values returns a slice of strings representing all valid JobCadenceFrequency values.
func (JobCadenceFrequency) Values() []string { return stringValues(jobCadenceFrequencyValues) }

// String returns the string representation of the JobCadenceFrequency value.
func (r JobCadenceFrequency) String() string { return string(r) }

// ToJobCadenceFrequency converts a string to its corresponding JobCadenceFrequency enum value.
func ToJobCadenceFrequency(r string) *JobCadenceFrequency {
	return parse(r, jobCadenceFrequencyValues, &JobCadenceFrequencyInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobCadenceFrequency) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobCadenceFrequency) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
