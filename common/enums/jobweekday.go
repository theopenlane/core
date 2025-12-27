package enums

import (
	"fmt"
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

// Values returns a slice of strings representing all valid JobWeekday values.
func (JobWeekday) Values() []string {
	return []string{
		string(JobWeekdaySunday),
		string(JobWeekdayMonday),
		string(JobWeekdayTuesday),
		string(JobWeekdayWednesday),
		string(JobWeekdayThursday),
		string(JobWeekdayFriday),
		string(JobWeekdaySaturday),
	}
}

// String returns the string representation of the JobWeekday value.
func (r JobWeekday) String() string {
	return strings.ToUpper(string(r))
}

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
func ToJobWeekday(r string) *JobWeekday {
	switch strings.ToUpper(r) {
	case JobWeekdaySunday.String():
		return &JobWeekdaySunday
	case JobWeekdayMonday.String():
		return &JobWeekdayMonday
	case JobWeekdayTuesday.String():
		return &JobWeekdayTuesday
	case JobWeekdayWednesday.String():
		return &JobWeekdayWednesday
	case JobWeekdayThursday.String():
		return &JobWeekdayThursday
	case JobWeekdayFriday.String():
		return &JobWeekdayFriday
	case JobWeekdaySaturday.String():
		return &JobWeekdaySaturday
	default:
		return &JobWeekdayInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobWeekday) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobWeekday) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobWeekday, got: %T", v) //nolint:err113
	}

	*r = JobWeekday(str)

	return nil
}
