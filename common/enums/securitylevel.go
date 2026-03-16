package enums

import "io"

// SecurityLevel is a custom type representing the various states of SecurityLevel.
type SecurityLevel string

var (
	// SecurityLevelNone indicates the none.
	SecurityLevelNone SecurityLevel = "NONE"
	// SecurityLevelLow indicates the low.
	SecurityLevelLow SecurityLevel = "LOW"
	// SecurityLevelMedium indicates the medium.
	SecurityLevelMedium SecurityLevel = "MEDIUM"
	// SecurityLevelHigh indicates the high.
	SecurityLevelHigh SecurityLevel = "HIGH"
	// SecurityLevelCritical indicates the critical.
	SecurityLevelCritical SecurityLevel = "CRITICAL"
	// SecurityLevelInvalid is used when an unknown or unsupported value is provided.
	SecurityLevelInvalid SecurityLevel = "SECURITYLEVEL_INVALID"
)

var securityLevelValues = []SecurityLevel{
	SecurityLevelNone,
	SecurityLevelLow,
	SecurityLevelMedium,
	SecurityLevelHigh,
	SecurityLevelCritical,
}

// Values returns a slice of strings representing all valid SecurityLevel values.
func (SecurityLevel) Values() []string { return stringValues(securityLevelValues) }

// String returns the string representation of the SecurityLevel value.
func (r SecurityLevel) String() string { return string(r) }

// ToSecurityLevel converts a string to its corresponding SecurityLevel enum value.
func ToSecurityLevel(r string) *SecurityLevel { return parse(r, securityLevelValues, &SecurityLevelInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r SecurityLevel) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *SecurityLevel) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
