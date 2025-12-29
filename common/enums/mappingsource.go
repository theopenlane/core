package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings representing all valid MappingSource values.
func (MappingSource) Values() []string {
	return []string{
		string(MappingSourceManual),
		string(MappingSourceSuggested),
		string(MappingSourceImported),
	}
}

// String returns the string representation of the MappingSource value.
func (r MappingSource) String() string {
	return string(r)
}

// ToMappingSource converts a string to its corresponding MappingSource enum value.
func ToMappingSource(r string) *MappingSource {
	switch strings.ToUpper(r) {
	case MappingSourceManual.String():
		return &MappingSourceManual
	case MappingSourceSuggested.String():
		return &MappingSourceSuggested
	case MappingSourceImported.String():
		return &MappingSourceImported
	default:
		return &MappingSourceInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r MappingSource) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *MappingSource) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for MappingSource, got: %T", v) //nolint:err113
	}

	*r = MappingSource(str)

	return nil
}
