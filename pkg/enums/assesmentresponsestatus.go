package enums

import (
	"fmt"
	"io"
	"strings"
)

// AssesmentResponseStatus is a custom type representing the various states of AssesmentResponseStatus.
type AssesmentResponseStatus string

var (
	// AssesmentResponseStatusNotStarted indicates the not started.
	AssesmentResponseStatusNotStarted AssesmentResponseStatus = "NOT_STARTED"
	// AssesmentResponseStatusCompleted indicates the completed.
	AssesmentResponseStatusCompleted AssesmentResponseStatus = "COMPLETED"
	// AssesmentResponseStatusOverdue indicates the overdue.
	AssesmentResponseStatusOverdue AssesmentResponseStatus = "OVERDUE"
	// AssesmentResponseStatusInvalid is used when an unknown or unsupported value is provided.
	AssesmentResponseStatusInvalid AssesmentResponseStatus = "ASSESMENTRESPONSESTATUS_INVALID"
)

// Values returns a slice of strings representing all valid AssesmentResponseStatus values.
func (AssesmentResponseStatus) Values() []string {
	return []string{
		string(AssesmentResponseStatusNotStarted),
		string(AssesmentResponseStatusCompleted),
		string(AssesmentResponseStatusOverdue),
	}
}

// String returns the string representation of the AssesmentResponseStatus value.
func (r AssesmentResponseStatus) String() string {
	return string(r)
}

// ToAssesmentResponseStatus converts a string to its corresponding AssesmentResponseStatus enum value.
func ToAssesmentResponseStatus(r string) *AssesmentResponseStatus {
	switch strings.ToUpper(r) {
	case AssesmentResponseStatusNotStarted.String():
		return &AssesmentResponseStatusNotStarted
	case AssesmentResponseStatusCompleted.String():
		return &AssesmentResponseStatusCompleted
	case AssesmentResponseStatusOverdue.String():
		return &AssesmentResponseStatusOverdue
	default:
		return &AssesmentResponseStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r AssesmentResponseStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *AssesmentResponseStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for AssesmentResponseStatus, got: %T", v)  //nolint:err113
	}

	*r = AssesmentResponseStatus(str)

	return nil
}
