package enums

import "io"

// JobEnvironment is a custom type representing the various states of JobEnvironment.
type JobEnvironment string

var (
	// JobEnvironmentOpenlane indicates the openlane.
	JobEnvironmentOpenlane JobEnvironment = "OPENLANE"
	// JobEnvironmentExternal indicates the external.
	JobEnvironmentExternal JobEnvironment = "EXTERNAL"
	// JobEnvironmentInvalid is used when an unknown or unsupported value is provided.
	JobEnvironmentInvalid JobEnvironment = "JOBENVIRONMENT_INVALID"
)

var jobEnvironmentValues = []JobEnvironment{JobEnvironmentOpenlane, JobEnvironmentExternal}

// Values returns a slice of strings representing all valid JobEnvironment values.
func (JobEnvironment) Values() []string { return stringValues(jobEnvironmentValues) }

// String returns the string representation of the JobEnvironment value.
func (r JobEnvironment) String() string { return string(r) }

// ToJobEnvironment converts a string to its corresponding JobEnvironment enum value.
func ToJobEnvironment(r string) *JobEnvironment {
	return parse(r, jobEnvironmentValues, &JobEnvironmentInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobEnvironment) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobEnvironment) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
