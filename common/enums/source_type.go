package enums

import "io"

// SourceType is a custom type representing the origin of a record
type SourceType string

var (
	// SourceTypeManual indicates the record was manually created
	SourceTypeManual SourceType = "MANUAL"
	// SourceTypeDiscovered indicates the record was discovered automatically
	SourceTypeDiscovered SourceType = "DISCOVERED"
	// SourceTypeImported indicates the record was imported from another system
	SourceTypeImported SourceType = "IMPORTED"
	// SourceTypeAPI indicates the record was created via API
	SourceTypeAPI SourceType = "API"
	// SourceTypeInvalid is used when an unknown or unsupported value is provided
	SourceTypeInvalid SourceType = "INVALID"
)

var sourceTypeValues = []SourceType{SourceTypeManual, SourceTypeDiscovered, SourceTypeImported, SourceTypeAPI}

// Values returns a slice of strings that represents all the possible values of the SourceType enum
// Possible default values are "MANUAL", "DISCOVERED", "IMPORTED", and "API"
func (SourceType) Values() []string { return stringValues(sourceTypeValues) }

// String returns the SourceType as a string
func (r SourceType) String() string { return string(r) }

// ToSourceType returns the source type enum based on string input
func ToSourceType(r string) *SourceType { return parse(r, sourceTypeValues, &SourceTypeInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r SourceType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *SourceType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
