package enums

import "io"

// MappingSource is a custom type representing the various states of MappingSource.
type MappingSource string

var (
	// MappingSourceManual indicates the mapping is manual
	MappingSourceManual MappingSource = "MANUAL"
	// MappingSourceSuggested indicates the mapping was suggested
	MappingSourceSuggested MappingSource = "SUGGESTED"
	// MappingSourceImported indicates the  mapping was imported
	MappingSourceImported MappingSource = "IMPORTED"
	// MappingSourceInvalid is used when an unknown or unsupported value is provided.
	MappingSourceInvalid MappingSource = "MAPPINGSOURCE_INVALID"
)

var mappingSourceValues = []MappingSource{MappingSourceManual, MappingSourceSuggested, MappingSourceImported}

// Values returns a slice of strings representing all valid MappingSource values.
func (MappingSource) Values() []string { return stringValues(mappingSourceValues) }

// String returns the string representation of the MappingSource value.
func (r MappingSource) String() string { return string(r) }

// ToMappingSource converts a string to its corresponding MappingSource enum value.
func ToMappingSource(r string) *MappingSource { return parse(r, mappingSourceValues, &MappingSourceInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r MappingSource) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *MappingSource) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
