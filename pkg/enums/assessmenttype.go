package enums

import (
	"fmt"
	"io"
	"strings"
)

// AssessmentType is a custom type representing the various states of AssessmentType.
type AssessmentType string

var (
	// AssessmentTypeInternal indicates the internal.
	AssessmentTypeInternal AssessmentType = "INTERNAL"
	// AssessmentTypeInvalid is used when an unknown or unsupported value is provided.
	AssessmentTypeInvalid AssessmentType = "ASSESSMENTTYPE_INVALID"
)

// Values returns a slice of strings representing all valid AssessmentType values.
func (AssessmentType) Values() []string {
	return []string{
		string(AssessmentTypeInternal),
	}
}

// String returns the string representation of the AssessmentType value.
func (r AssessmentType) String() string {
	return string(r)
}

// ToAssessmentType converts a string to its corresponding AssessmentType enum value.
func ToAssessmentType(r string) *AssessmentType {
	switch strings.ToUpper(r) {
	case AssessmentTypeInternal.String():
		return &AssessmentTypeInternal
	default:
		return &AssessmentTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r AssessmentType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *AssessmentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for AssessmentType, got: %T", v)  //nolint:err113
	}

	*r = AssessmentType(str)

	return nil
}
