package enums

import "io"

// MappingType is a custom type representing the various states of MappingType.
type MappingType string

var (
	// MappingTypeEqual indicates the two sets are equivalent
	MappingTypeEqual MappingType = "EQUAL"
	// MappingTypeSuperset indicates the mapping is superset
	MappingTypeSuperset MappingType = "SUPERSET"
	// MappingTypeSubset indicates the mapping is subset
	MappingTypeSubset MappingType = "SUBSET"
	// MappingTypeIntersect indicates the overlap is an intersection
	MappingTypeIntersect MappingType = "INTERSECT"
	// MappingTypePartial indicates a partial overlap, but the exact type is unspecified
	MappingTypePartial MappingType = "PARTIAL"
	// MappingTypeInvalid is used when an unknown or unsupported value is provided.
	MappingTypeInvalid MappingType = "MAPPINGTYPE_INVALID"
)

var mappingTypeValues = []MappingType{MappingTypeEqual, MappingTypeSuperset, MappingTypeSubset, MappingTypeIntersect, MappingTypePartial}

// Values returns a slice of strings representing all valid MappingType values.
func (MappingType) Values() []string { return stringValues(mappingTypeValues) }

// String returns the string representation of the MappingType value.
func (r MappingType) String() string { return string(r) }

// ToMappingType converts a string to its corresponding MappingType enum value.
func ToMappingType(r string) *MappingType { return parse(r, mappingTypeValues, &MappingTypeInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r MappingType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *MappingType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
