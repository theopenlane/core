package enums

import (
	"fmt"
	"io"
	"strings"
)

// JobPlatformType is a custom type representing the various states of JobPlatformType.
type JobPlatformType string

var (
	// JobPlatformTypeGo indicates the go.
	JobPlatformTypeGo JobPlatformType = "GO"
	// JobPlatformTypeTs indicates the ts.
	JobPlatformTypeTs JobPlatformType = "TS"
	// JobPlatformTypeInvalid is used when an unknown or unsupported value is provided.
	JobPlatformTypeInvalid JobPlatformType = "JOBPLATFORMTYPE_INVALID"
)

// Values returns a slice of strings representing all valid JobPlatformType values.
func (JobPlatformType) Values() []string {
	return []string{
		string(JobPlatformTypeGo),
		string(JobPlatformTypeTs),
	}
}

// String returns the string representation of the JobPlatformType value.
func (r JobPlatformType) String() string {
	return string(r)
}

// ToJobPlatformType converts a string to its corresponding JobPlatformType enum value.
func ToJobPlatformType(r string) *JobPlatformType {
	switch strings.ToUpper(r) {
	case JobPlatformTypeGo.String():
		return &JobPlatformTypeGo
	case JobPlatformTypeTs.String():
		return &JobPlatformTypeTs
	default:
		return &JobPlatformTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobPlatformType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobPlatformType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobPlatformType, got: %T", v) //nolint:err113
	}

	*r = JobPlatformType(str)

	return nil
}
