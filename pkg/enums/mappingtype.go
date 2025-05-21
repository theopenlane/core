package enums

import (
	"fmt"
	"io"
	"strings"
)

// MappingType is a custom type representing the various states of MappingType.
type MappingType string

var (
	// MappingTypeEqual indicates the equal.
	MappingTypeEqual MappingType = "EQUAL"
	// MappingTypeSuperset indicates the superset.
	MappingTypeSuperset MappingType = "SUPERSET"
	// MappingTypeSubset indicates the subset.
	MappingTypeSubset MappingType = "SUBSET"
	// MappingTypeIntersect indicates the intersect.
	MappingTypeIntersect MappingType = "INTERSECT"
	// MappingTypeInvalid is used when an unknown or unsupported value is provided.
	MappingTypeInvalid MappingType = "MAPPINGTYPE_INVALID"
)

// Values returns a slice of strings representing all valid MappingType values.
func (MappingType) Values() []string {
	return []string{
		string(MappingTypeEqual),
		string(MappingTypeSuperset),
		string(MappingTypeSubset),
		string(MappingTypeIntersect),
	}
}

// String returns the string representation of the MappingType value.
func (r MappingType) String() string {
	return string(r)
}

// ToMappingType converts a string to its corresponding MappingType enum value.
func ToMappingType(r string) *MappingType {
	switch strings.ToUpper(r) {
	case MappingTypeEqual.String():
		return &MappingTypeEqual
	case MappingTypeSuperset.String():
		return &MappingTypeSuperset
	case MappingTypeSubset.String():
		return &MappingTypeSubset
	case MappingTypeIntersect.String():
		return &MappingTypeIntersect
	default:
		return &MappingTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r MappingType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *MappingType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for MappingType, got: %T", v) //nolint:err113
	}

	*r = MappingType(str)

	return nil
}
