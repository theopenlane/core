package enums

import "io"

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
	// AssessmentResponseStatusDraft indicates the draft.
	AssessmentResponseStatusDraft AssessmentResponseStatus = "DRAFT"
	// AssessmentResponseStatusInvalid is used when an unknown or unsupported value is provided.
	AssessmentResponseStatusInvalid AssessmentResponseStatus = "ASSESSMENTRESPONSESTATUS_INVALID"
)

var assessmentResponseStatusValues = []AssessmentResponseStatus{
	AssessmentResponseStatusNotStarted,
	AssessmentResponseStatusSent,
	AssessmentResponseStatusCompleted,
	AssessmentResponseStatusOverdue,
	AssessmentResponseStatusDraft,
}

// Values returns a slice of strings representing all valid AssessmentResponseStatus values.
func (AssessmentResponseStatus) Values() []string {
	return stringValues(assessmentResponseStatusValues)
}

// String returns the string representation of the AssessmentResponseStatus value.
func (r AssessmentResponseStatus) String() string {
	return string(r)
}

// ToAssessmentResponseStatus converts a string to its corresponding AssessmentResponseStatus enum value.
func ToAssessmentResponseStatus(r string) *AssessmentResponseStatus {
	return parse(r, assessmentResponseStatusValues, &AssessmentResponseStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r AssessmentResponseStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *AssessmentResponseStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
