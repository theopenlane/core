package enums

import "io"

// TemplateKind is a custom type representing the various states of TemplateKind.
type TemplateKind string

var (
	// TemplateKindQuestionnaire indicates the questionnaire.
	TemplateKindQuestionnaire TemplateKind = "QUESTIONNAIRE"
	// TemplateKindTrustCenterNda indicates the trustcenter nda.
	TemplateKindTrustCenterNda TemplateKind = "TRUSTCENTER_NDA"
	// TemplateKindExternalIntake indicates the external intake.
	TemplateKindExternalIntake TemplateKind = "EXTERNAL_INTAKE"
	// TemplateKindInvalid is used when an unknown or unsupported value is provided.
	TemplateKindInvalid TemplateKind = "TEMPLATEKIND_INVALID"
)

var templateKindValues = []TemplateKind{
	TemplateKindQuestionnaire,
	TemplateKindTrustCenterNda,
	TemplateKindExternalIntake,
}

// Values returns a slice of strings representing all valid TemplateKind values.
func (TemplateKind) Values() []string { return stringValues(templateKindValues) }

// String returns the string representation of the TemplateKind value.
func (r TemplateKind) String() string { return string(r) }

// ToTemplateKind converts a string to its corresponding TemplateKind enum value.
func ToTemplateKind(r string) *TemplateKind {
	return parse(r, templateKindValues, &TemplateKindInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateKind) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateKind) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
