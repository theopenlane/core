package enums

import "io"

// AudienceType represents how an audience resolves recipients.
type AudienceType string

var (
	AudienceTypeManual  AudienceType = "MANUAL"
	AudienceTypeDynamic AudienceType = "DYNAMIC"
	AudienceTypeInvalid AudienceType = "INVALID"
)

var audienceTypeValues = []AudienceType{
	AudienceTypeManual,
	AudienceTypeDynamic,
}

// Values returns all AudienceType values.
func (AudienceType) Values() []string { return stringValues(audienceTypeValues) }

// String returns the AudienceType as a string.
func (r AudienceType) String() string { return string(r) }

// ToAudienceType parses an AudienceType from a string.
func ToAudienceType(r string) *AudienceType {
	return parse(r, audienceTypeValues, &AudienceTypeInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (r AudienceType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (r *AudienceType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
