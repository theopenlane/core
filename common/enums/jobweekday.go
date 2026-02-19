package enums

import (
	"io"
	"strings"
	"time"
)

// JobWeekday is a custom type representing the various states of JobWeekday.
type JobWeekday string

var (
	// JobWeekdaySunday indicates the sunday.
	JobWeekdaySunday JobWeekday = "SUNDAY"
	// JobWeekdayMonday indicates the monday.
	JobWeekdayMonday JobWeekday = "MONDAY"
	// JobWeekdayTuesday indicates the tuesday.
	JobWeekdayTuesday JobWeekday = "TUESDAY"
	// JobWeekdayWednesday indicates the wednesday.
	JobWeekdayWednesday JobWeekday = "WEDNESDAY"
	// JobWeekdayThursday indicates the thursday.
	JobWeekdayThursday JobWeekday = "THURSDAY"
	// JobWeekdayFriday indicates the friday.
	JobWeekdayFriday JobWeekday = "FRIDAY"
	// JobWeekdaySaturday indicates the saturday.
	JobWeekdaySaturday JobWeekday = "SATURDAY"
	// JobWeekdayInvalid is used when an unknown or unsupported value is provided.
	JobWeekdayInvalid JobWeekday = "JOBWEEKDAY_INVALID"
)

var jobWeekdayValues = []JobWeekday{
	JobWeekdaySunday, JobWeekdayMonday, JobWeekdayTuesday, JobWeekdayWednesday,
	JobWeekdayThursday, JobWeekdayFriday, JobWeekdaySaturday,
}

// Values returns a slice of strings representing all valid JobWeekday values.
func (JobWeekday) Values() []string { return stringValues(jobWeekdayValues) }

// String returns the string representation of the JobWeekday value.
func (r JobWeekday) String() string { return strings.ToUpper(string(r)) }

// ToTimeWeekday maps the human readable enums to Go's weekday type
func ToTimeWeekday(r JobWeekday) time.Weekday {
	switch r {
	case JobWeekdaySunday:
		return time.Sunday
	case JobWeekdayMonday:
		return time.Monday
	case JobWeekdayTuesday:
		return time.Tuesday
	case JobWeekdayWednesday:
		return time.Wednesday
	case JobWeekdayThursday:
		return time.Thursday
	case JobWeekdayFriday:
		return time.Friday
	case JobWeekdaySaturday:
		return time.Saturday
	default:
		const defaultWeekDay = 10
		return time.Weekday(defaultWeekDay)
	}
}

// ToJobWeekday converts a string to its corresponding JobWeekday enum value.
func ToJobWeekday(r string) *JobWeekday { return parse(r, jobWeekdayValues, &JobWeekdayInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobWeekday) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobWeekday) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
