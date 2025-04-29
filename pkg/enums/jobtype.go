package enums

import (
	"fmt"
	"io"
	"strings"
)

// JobType is a custom type representing the various states of JobType.
type JobType string

var (
	// JobTypeSsl indicates the ssl.
	JobTypeSsl JobType = "SSL"
	// JobTypeInvalid is used when an unknown or unsupported value is provided.
	JobTypeInvalid JobType = "JOBTYPE_INVALID"
)

// Values returns a slice of strings representing all valid JobType values.
func (JobType) Values() []string {
	return []string{
		string(JobTypeSsl),
	}
}

// String returns the string representation of the JobType value.
func (r JobType) String() string {
	return string(r)
}

// ToJobType converts a string to its corresponding JobType enum value.
func ToJobType(r string) *JobType {
	switch strings.ToUpper(r) {
	case JobTypeSsl.String():
		return &JobTypeSsl
	default:
		return &JobTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobType, got: %T", v)
	}
	*r = JobType(str)
	return nil
}
