package enums

import "io"

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

var platformStatusValues = []PlatformStatus{PlatformStatusActive, PlatformStatusInactive, PlatformStatusRetired}

// Values returns a slice of strings that represents all the possible values of the PlatformStatus enum
// Possible default values are "ACTIVE", "INACTIVE", and "RETIRED"
func (PlatformStatus) Values() []string { return stringValues(platformStatusValues) }

// String returns the PlatformStatus as a string
func (r PlatformStatus) String() string { return string(r) }

// ToPlatformStatus returns the platform status enum based on string input
func ToPlatformStatus(r string) *PlatformStatus {
	return parse(r, platformStatusValues, &PlatformStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r PlatformStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *PlatformStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
