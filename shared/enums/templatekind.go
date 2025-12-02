package enums

import (
	"fmt"
	"io"
	"strings"
)

// TemplateKind is a custom type representing the various states of TemplateKind.
type TemplateKind string

var (
	// TemplateKindQuestionnaire indicates the questionnaire.
	TemplateKindQuestionnaire TemplateKind = "QUESTIONNAIRE"
	// TemplateKindTrustCenterNda indicates the trust center NDA.
	TemplateKindTrustCenterNda TemplateKind = "TRUSTCENTER_NDA"
	// TemplateKindInvalid is used when an unknown or unsupported value is provided.
	TemplateKindInvalid TemplateKind = "TEMPLATEKIND_INVALID"
)

// Values returns a slice of strings representing all valid TemplateKind values.
func (TemplateKind) Values() []string {
	return []string{
		string(TemplateKindQuestionnaire),
		string(TemplateKindTrustCenterNda),
	}
}

// String returns the string representation of the TemplateKind value.
func (r TemplateKind) String() string {
	return string(r)
}

// ToTemplateKind converts a string to its corresponding TemplateKind enum value.
func ToTemplateKind(r string) *TemplateKind {
	switch strings.ToUpper(r) {
	case TemplateKindQuestionnaire.String():
		return &TemplateKindQuestionnaire
	case TemplateKindTrustCenterNda.String():
		return &TemplateKindTrustCenterNda
	default:
		return &TemplateKindInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateKind) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateKind) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TemplateKind, got: %T", v) //nolint:err113
	}

	*r = TemplateKind(str)

	return nil
}
