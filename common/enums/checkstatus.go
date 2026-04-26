package enums

import "io"

// CheckStatus is a custom type representing the various states of CheckStatus.
type CheckStatus string

var (
	// CheckStatusPass indicates the compliant.
	CheckStatusPass CheckStatus = "PASS"
	// CheckStatusFail indicates the not compliant.
	CheckStatusFail CheckStatus = "FAIL"
	// CheckStatusUnknown indicates the unknown.
	CheckStatusUnknown CheckStatus = "UNKNOWN"
	// CheckStatusInvalid is used when an unknown or unsupported value is provided.
	CheckStatusInvalid CheckStatus = "CheckSTATUS_INVALID"
)

var CheckStatusValues = []CheckStatus{
	CheckStatusPass,
	CheckStatusFail,
	CheckStatusUnknown,
}

// Values returns a slice of strings representing all valid CheckStatus values.
func (CheckStatus) Values() []string { return stringValues(CheckStatusValues) }

// String returns the string representation of the CheckStatus value.
func (r CheckStatus) String() string { return string(r) }

// ToCheckStatus converts a string to its corresponding CheckStatus enum value.
func ToCheckStatus(r string) *CheckStatus {
	return parse(r, CheckStatusValues, &CheckStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r CheckStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *CheckStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
