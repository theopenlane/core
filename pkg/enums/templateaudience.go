package enums

import (
	"fmt"
	"io"
	"strings"
)

// TemplateAudience is a custom type representing the various states of TemplateAudience.
type TemplateAudience string

var (
	// TemplateAudienceInternal indicates the internal.
	TemplateAudienceInternal TemplateAudience = "INTERNAL"
	// TemplateAudienceExternal indicates the external.
	TemplateAudienceExternal TemplateAudience = "EXTERNAL"
	// TemplateAudienceInvalid is used when an unknown or unsupported value is provided.
	TemplateAudienceInvalid TemplateAudience = "TEMPLATEAUDIENCE_INVALID"
)

// Values returns a slice of strings representing all valid TemplateAudience values.
func (TemplateAudience) Values() []string {
	return []string{
		string(TemplateAudienceInternal),
		string(TemplateAudienceExternal),
	}
}

// String returns the string representation of the TemplateAudience value.
func (r TemplateAudience) String() string {
	return string(r)
}

// ToTemplateAudience converts a string to its corresponding TemplateAudience enum value.
func ToTemplateAudience(r string) *TemplateAudience {
	switch strings.ToUpper(r) {
	case TemplateAudienceInternal.String():
		return &TemplateAudienceInternal
	case TemplateAudienceExternal.String():
		return &TemplateAudienceExternal
	default:
		return &TemplateAudienceInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateAudience) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateAudience) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TemplateAudience, got: %T", v)  //nolint:err113
	}

	*r = TemplateAudience(str)

	return nil
}
