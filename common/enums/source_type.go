package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the SourceType enum
// Possible default values are "MANUAL", "DISCOVERED", "IMPORTED", and "API"
func (SourceType) Values() (kinds []string) {
	for _, s := range []SourceType{
		SourceTypeManual,
		SourceTypeDiscovered,
		SourceTypeImported,
		SourceTypeAPI,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the SourceType as a string
func (r SourceType) String() string {
	return string(r)
}

// ToSourceType returns the source type enum based on string input
func ToSourceType(r string) *SourceType {
	switch r := strings.ToUpper(r); r {
	case SourceTypeManual.String():
		return &SourceTypeManual
	case SourceTypeDiscovered.String():
		return &SourceTypeDiscovered
	case SourceTypeImported.String():
		return &SourceTypeImported
	case SourceTypeAPI.String():
		return &SourceTypeAPI
	default:
		return &SourceTypeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r SourceType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *SourceType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for SourceType, got: %T", v) //nolint:err113
	}

	*r = SourceType(str)

	return nil
}
