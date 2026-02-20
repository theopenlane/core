package enums

import "io"

// AssessmentType is a custom type representing the various states of AssessmentType.
type AssessmentType string

var (
	// AssessmentTypeInternal indicates the internal.
	AssessmentTypeInternal AssessmentType = "INTERNAL"
	// AssessmentTypeExternal indicates the external.
	AssessmentTypeExternal AssessmentType = "EXTERNAL"
	// AssessmentTypeInvalid is used when an unknown or unsupported value is provided.
	AssessmentTypeInvalid AssessmentType = "ASSESSMENTTYPE_INVALID"
)

var assessmentTypeValues = []AssessmentType{
	AssessmentTypeInternal,
	AssessmentTypeExternal,
}

// Values returns a slice of strings representing all valid AssessmentType values.
func (AssessmentType) Values() []string { return stringValues(assessmentTypeValues) }

// String returns the string representation of the AssessmentType value.
func (r AssessmentType) String() string { return string(r) }

// ToAssessmentType converts a string to its corresponding AssessmentType enum value.
func ToAssessmentType(r string) *AssessmentType {
	return parse(r, assessmentTypeValues, &AssessmentTypeInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r AssessmentType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *AssessmentType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
