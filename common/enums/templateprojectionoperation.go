package enums

import "io"

// TemplateProjectionOperation is a custom type representing the various states of TemplateProjectionOperation.
type TemplateProjectionOperation string

var (
	// TemplateProjectionOperationCreate indicates the create.
	TemplateProjectionOperationCreate TemplateProjectionOperation = "CREATE"
	// TemplateProjectionOperationUpdate indicates the update.
	TemplateProjectionOperationUpdate TemplateProjectionOperation = "UPDATE"
	// TemplateProjectionOperationInvalid is used when an unknown or unsupported value is provided.
	TemplateProjectionOperationInvalid TemplateProjectionOperation = "TEMPLATEPROJECTIONOPERATION_INVALID"
)

var templateProjectionOperationValues = []TemplateProjectionOperation{
	TemplateProjectionOperationCreate,
	TemplateProjectionOperationUpdate,
}

// Values returns a slice of strings representing all valid TemplateProjectionOperation values.
func (TemplateProjectionOperation) Values() []string { return stringValues(templateProjectionOperationValues) }

// String returns the string representation of the TemplateProjectionOperation value.
func (r TemplateProjectionOperation) String() string { return string(r) }

// ToTemplateProjectionOperation converts a string to its corresponding TemplateProjectionOperation enum value.
func ToTemplateProjectionOperation(r string) *TemplateProjectionOperation { return parse(r, templateProjectionOperationValues, &TemplateProjectionOperationInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateProjectionOperation) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateProjectionOperation) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
