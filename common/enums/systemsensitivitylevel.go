package enums

import (
	"io"
	"strings"
)

// SystemSensitivityLevel is a custom type for system sensitivity categorization.
type SystemSensitivityLevel string

var (
	// SystemSensitivityLevelLow indicates low system sensitivity.
	SystemSensitivityLevelLow SystemSensitivityLevel = "LOW"
	// SystemSensitivityLevelModerate indicates moderate system sensitivity.
	SystemSensitivityLevelModerate SystemSensitivityLevel = "MODERATE"
	// SystemSensitivityLevelHigh indicates high system sensitivity.
	SystemSensitivityLevelHigh SystemSensitivityLevel = "HIGH"
	// SystemSensitivityLevelUnknown indicates unspecified or unknown sensitivity.
	SystemSensitivityLevelUnknown SystemSensitivityLevel = "UNKNOWN"
	// SystemSensitivityLevelInvalid indicates an invalid or unknown enum input.
	SystemSensitivityLevelInvalid SystemSensitivityLevel = "SYSTEM_SENSITIVITY_LEVEL_INVALID"
)

var systemSensitivityLevelValues = []SystemSensitivityLevel{
	SystemSensitivityLevelLow,
	SystemSensitivityLevelModerate,
	SystemSensitivityLevelHigh,
	SystemSensitivityLevelUnknown,
}

// Values returns all valid SystemSensitivityLevel values.
func (SystemSensitivityLevel) Values() []string { return stringValues(systemSensitivityLevelValues) }

// String returns the string representation of the level.
func (r SystemSensitivityLevel) String() string { return string(r) }

// ToSystemSensitivityLevel parses a string into a SystemSensitivityLevel.
// An empty input defaults to unknown.
func ToSystemSensitivityLevel(r string) *SystemSensitivityLevel {
	if strings.TrimSpace(r) == "" {
		return &SystemSensitivityLevelUnknown
	}

	return parse(r, systemSensitivityLevelValues, &SystemSensitivityLevelInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r SystemSensitivityLevel) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *SystemSensitivityLevel) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
