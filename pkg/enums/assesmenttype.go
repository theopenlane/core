package enums

import (
	"fmt"
	"io"
	"strings"
)

// AssesmentType is a custom type representing the various states of AssesmentType.
type AssesmentType string

var (
	// AssesmentTypeInternal indicates the internal.
	AssesmentTypeInternal AssesmentType = "INTERNAL"
	// AssesmentTypeExternal indicates the external.
	AssesmentTypeExternal AssesmentType = "EXTERNAL"
	// AssesmentTypeInvalid is used when an unknown or unsupported value is provided.
	AssesmentTypeInvalid AssesmentType = "ASSESMENTTYPE_INVALID"
)

// Values returns a slice of strings representing all valid AssesmentType values.
func (AssesmentType) Values() []string {
	return []string{
		string(AssesmentTypeInternal),
		string(AssesmentTypeExternal),
	}
}

// String returns the string representation of the AssesmentType value.
func (r AssesmentType) String() string {
	return string(r)
}

// ToAssesmentType converts a string to its corresponding AssesmentType enum value.
func ToAssesmentType(r string) *AssesmentType {
	switch strings.ToUpper(r) {
	case AssesmentTypeInternal.String():
		return &AssesmentTypeInternal
	case AssesmentTypeExternal.String():
		return &AssesmentTypeExternal
	default:
		return &AssesmentTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r AssesmentType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *AssesmentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for AssesmentType, got: %T", v)  //nolint:err113
	}

	*r = AssesmentType(str)

	return nil
}
