package enums

import "io"

// TemplateProjectionTarget is a custom type representing the various states of TemplateProjectionTarget.
type TemplateProjectionTarget string

var (
	// TemplateProjectionTargetEntity indicates the entity.
	TemplateProjectionTargetEntity TemplateProjectionTarget = "ENTITY"
	// TemplateProjectionTargetInvalid is used when an unknown or unsupported value is provided.
	TemplateProjectionTargetInvalid TemplateProjectionTarget = "TEMPLATEPROJECTIONTARGET_INVALID"
)

var templateProjectionTargetValues = []TemplateProjectionTarget{
	TemplateProjectionTargetEntity,
}

// Values returns a slice of strings representing all valid TemplateProjectionTarget values.
func (TemplateProjectionTarget) Values() []string {
	return stringValues(templateProjectionTargetValues)
}

// String returns the string representation of the TemplateProjectionTarget value.
func (r TemplateProjectionTarget) String() string { return string(r) }

// ToTemplateProjectionTarget converts a string to its corresponding TemplateProjectionTarget enum value.
func ToTemplateProjectionTarget(r string) *TemplateProjectionTarget {
	return parse(r, templateProjectionTargetValues, &TemplateProjectionTargetInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateProjectionTarget) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateProjectionTarget) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
