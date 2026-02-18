package enums

import "io"

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

var jobPlatformTypeValues = []JobPlatformType{JobPlatformTypeGo, JobPlatformTypeTs}

// Values returns a slice of strings representing all valid JobPlatformType values.
func (JobPlatformType) Values() []string { return stringValues(jobPlatformTypeValues) }

// String returns the string representation of the JobPlatformType value.
func (r JobPlatformType) String() string { return string(r) }

// ToJobPlatformType converts a string to its corresponding JobPlatformType enum value.
func ToJobPlatformType(r string) *JobPlatformType {
	return parse(r, jobPlatformTypeValues, &JobPlatformTypeInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobPlatformType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobPlatformType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
