package enums

import "io"

// CustomDomainType is a custom type representing the various states of CustomDomainType.
type CustomDomainType string

var (
	// CustomDomainTypePreview indicates the preview.
	CustomDomainTypePreview CustomDomainType = "PREVIEW"
	// CustomDomainTypeExternal indicates the external.
	CustomDomainTypeExternal CustomDomainType = "EXTERNAL"
	// CustomDomainTypeUnknown indicates the unknown.
	CustomDomainTypeUnknown CustomDomainType = "UNKNOWN"
	// CustomDomainTypeInvalid is used when an unknown or unsupported value is provided.
	CustomDomainTypeInvalid CustomDomainType = "CUSTOMDOMAINTYPE_INVALID"
)

var customDomainTypeValues = []CustomDomainType{
	CustomDomainTypePreview,
	CustomDomainTypeExternal,
	CustomDomainTypeUnknown,
}

// Values returns a slice of strings representing all valid CustomDomainType values.
func (CustomDomainType) Values() []string { return stringValues(customDomainTypeValues) }

// String returns the string representation of the CustomDomainType value.
func (r CustomDomainType) String() string { return string(r) }

// ToCustomDomainType converts a string to its corresponding CustomDomainType enum value.
func ToCustomDomainType(r string) *CustomDomainType { return parse(r, customDomainTypeValues, &CustomDomainTypeInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r CustomDomainType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *CustomDomainType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
