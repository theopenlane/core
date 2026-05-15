package enums

import "io"

// TemplateProjectionTransform is a custom type representing the various states of TemplateProjectionTransform.
type TemplateProjectionTransform string

var (
	// TemplateProjectionTransformSlugify indicates the slugify.
	TemplateProjectionTransformSlugify TemplateProjectionTransform = "SLUGIFY"
	// TemplateProjectionTransformDate indicates the date.
	TemplateProjectionTransformDate TemplateProjectionTransform = "DATE"
	// TemplateProjectionTransformString indicates the string.
	TemplateProjectionTransformString TemplateProjectionTransform = "STRING"
	// TemplateProjectionTransformBool indicates the bool.
	TemplateProjectionTransformBool TemplateProjectionTransform = "BOOL"
	// TemplateProjectionTransformFloat indicates the float.
	TemplateProjectionTransformFloat TemplateProjectionTransform = "FLOAT"
	// TemplateProjectionTransformStringArray indicates the string array.
	TemplateProjectionTransformStringArray TemplateProjectionTransform = "STRING_ARRAY"
	// TemplateProjectionTransformInvalid is used when an unknown or unsupported value is provided.
	TemplateProjectionTransformInvalid TemplateProjectionTransform = "TEMPLATEPROJECTIONTRANSFORM_INVALID"
)

var templateProjectionTransformValues = []TemplateProjectionTransform{
	TemplateProjectionTransformSlugify,
	TemplateProjectionTransformDate,
	TemplateProjectionTransformString,
	TemplateProjectionTransformBool,
	TemplateProjectionTransformFloat,
	TemplateProjectionTransformStringArray,
}

// Values returns a slice of strings representing all valid TemplateProjectionTransform values.
func (TemplateProjectionTransform) Values() []string {
	return stringValues(templateProjectionTransformValues)
}

// String returns the string representation of the TemplateProjectionTransform value.
func (r TemplateProjectionTransform) String() string { return string(r) }

// ToTemplateProjectionTransform converts a string to its corresponding TemplateProjectionTransform enum value.
func ToTemplateProjectionTransform(r string) *TemplateProjectionTransform {
	return parse(r, templateProjectionTransformValues, &TemplateProjectionTransformInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateProjectionTransform) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateProjectionTransform) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
