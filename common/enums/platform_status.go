package enums

import (
	"fmt"
	"io"
	"strings"
)

// PlatformStatus is a custom type representing the lifecycle state of a platform
type PlatformStatus string

var (
	// PlatformStatusActive indicates a platform is active
	PlatformStatusActive PlatformStatus = "ACTIVE"
	// PlatformStatusInactive indicates a platform is inactive
	PlatformStatusInactive PlatformStatus = "INACTIVE"
	// PlatformStatusRetired indicates a platform is retired
	PlatformStatusRetired PlatformStatus = "RETIRED"
	// PlatformStatusInvalid is used when an unknown or unsupported value is provided
	PlatformStatusInvalid PlatformStatus = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the PlatformStatus enum
// Possible default values are "ACTIVE", "INACTIVE", and "RETIRED"
func (PlatformStatus) Values() (kinds []string) {
	for _, s := range []PlatformStatus{PlatformStatusActive, PlatformStatusInactive, PlatformStatusRetired} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the PlatformStatus as a string
func (r PlatformStatus) String() string {
	return string(r)
}

// ToPlatformStatus returns the platform status enum based on string input
func ToPlatformStatus(r string) *PlatformStatus {
	switch r := strings.ToUpper(r); r {
	case PlatformStatusActive.String():
		return &PlatformStatusActive
	case PlatformStatusInactive.String():
		return &PlatformStatusInactive
	case PlatformStatusRetired.String():
		return &PlatformStatusRetired
	default:
		return &PlatformStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r PlatformStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *PlatformStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for PlatformStatus, got: %T", v) //nolint:err113
	}

	*r = PlatformStatus(str)

	return nil
}
