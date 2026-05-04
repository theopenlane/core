package enums

import "io"

// TemplateProjectionTrigger is a custom type representing the various states of TemplateProjectionTrigger.
type TemplateProjectionTrigger string

var (
	// TemplateProjectionTriggerCompleted indicates the completed.
	TemplateProjectionTriggerCompleted TemplateProjectionTrigger = "COMPLETED"
	// TemplateProjectionTriggerInvalid is used when an unknown or unsupported value is provided.
	TemplateProjectionTriggerInvalid TemplateProjectionTrigger = "TEMPLATEPROJECTIONTRIGGER_INVALID"
)

var templateProjectionTriggerValues = []TemplateProjectionTrigger{
	TemplateProjectionTriggerCompleted,
}

// Values returns a slice of strings representing all valid TemplateProjectionTrigger values.
func (TemplateProjectionTrigger) Values() []string { return stringValues(templateProjectionTriggerValues) }

// String returns the string representation of the TemplateProjectionTrigger value.
func (r TemplateProjectionTrigger) String() string { return string(r) }

// ToTemplateProjectionTrigger converts a string to its corresponding TemplateProjectionTrigger enum value.
func ToTemplateProjectionTrigger(r string) *TemplateProjectionTrigger { return parse(r, templateProjectionTriggerValues, &TemplateProjectionTriggerInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateProjectionTrigger) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateProjectionTrigger) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
