package enums

import (
	"fmt"
	"io"
	"strings"
)

// AssessmentResponseStatus is a custom type representing the various states of AssessmentResponseStatus.
type AssessmentResponseStatus string

var (
	// AssessmentResponseStatusNotStarted indicates the not started.
	AssessmentResponseStatusNotStarted AssessmentResponseStatus = "NOT_STARTED"
	// AssessmentResponseStatusSent indicates the sent.
	AssessmentResponseStatusSent AssessmentResponseStatus = "SENT"
	// AssessmentResponseStatusCompleted indicates the completed.
	AssessmentResponseStatusCompleted AssessmentResponseStatus = "COMPLETED"
	// AssessmentResponseStatusOverdue indicates the overdue.
	AssessmentResponseStatusOverdue AssessmentResponseStatus = "OVERDUE"
	// AssessmentResponseStatusInvalid is used when an unknown or unsupported value is provided.
	AssessmentResponseStatusInvalid AssessmentResponseStatus = "ASSESSMENTRESPONSESTATUS_INVALID"
)

// Values returns a slice of strings representing all valid AssessmentResponseStatus values.
func (AssessmentResponseStatus) Values() []string {
	return []string{
		string(AssessmentResponseStatusNotStarted),
		string(AssessmentResponseStatusSent),
		string(AssessmentResponseStatusCompleted),
		string(AssessmentResponseStatusOverdue),
	}
}

// String returns the string representation of the AssessmentResponseStatus value.
func (r AssessmentResponseStatus) String() string {
	return string(r)
}

// ToAssessmentResponseStatus converts a string to its corresponding AssessmentResponseStatus enum value.
func ToAssessmentResponseStatus(r string) *AssessmentResponseStatus {
	switch strings.ToUpper(r) {
	case AssessmentResponseStatusNotStarted.String():
		return &AssessmentResponseStatusNotStarted
	case AssessmentResponseStatusSent.String():
		return &AssessmentResponseStatusSent
	case AssessmentResponseStatusCompleted.String():
		return &AssessmentResponseStatusCompleted
	case AssessmentResponseStatusOverdue.String():
		return &AssessmentResponseStatusOverdue
	default:
		return &AssessmentResponseStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r AssessmentResponseStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *AssessmentResponseStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for AssessmentResponseStatus, got: %T", v) //nolint:err113
	}

	*r = AssessmentResponseStatus(str)

	return nil
}
