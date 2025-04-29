package enums

import (
	"fmt"
	"io"
	"strings"
)

// JobEnvironment is a custom type representing the various states of JobEnvironment.
type JobEnvironment string

var (
	// JobEnvironmentOpenlane indicates the openlane.
	JobEnvironmentOpenlane JobEnvironment = "OPENLANE"
	// JobEnvironmentCustomer indicates the customer.
	JobEnvironmentCustomer JobEnvironment = "CUSTOMER"
	// JobEnvironmentInvalid is used when an unknown or unsupported value is provided.
	JobEnvironmentInvalid JobEnvironment = "JOBENVIRONMENT_INVALID"
)

// Values returns a slice of strings representing all valid JobEnvironment values.
func (JobEnvironment) Values() []string {
	return []string{
		string(JobEnvironmentOpenlane),
		string(JobEnvironmentCustomer),
	}
}

// String returns the string representation of the JobEnvironment value.
func (r JobEnvironment) String() string {
	return string(r)
}

// ToJobEnvironment converts a string to its corresponding JobEnvironment enum value.
func ToJobEnvironment(r string) *JobEnvironment {
	switch strings.ToUpper(r) {
	case JobEnvironmentOpenlane.String():
		return &JobEnvironmentOpenlane
	case JobEnvironmentCustomer.String():
		return &JobEnvironmentCustomer
	default:
		return &JobEnvironmentInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r JobEnvironment) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *JobEnvironment) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JobEnvironment, got: %T", v) //nolint:err113
	}

	*r = JobEnvironment(str)

	return nil
}
