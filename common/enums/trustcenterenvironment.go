package enums

import "io"

// TrustCenterEnvironment is a custom type representing the various states of TrustCenterEnvironment.
type TrustCenterEnvironment string

var (
	// TrustCenterEnvironmentLive indicates the live.
	TrustCenterEnvironmentLive TrustCenterEnvironment = "LIVE"
	// TrustCenterEnvironmentPreview indicates the preview.
	TrustCenterEnvironmentPreview TrustCenterEnvironment = "PREVIEW"
	// TrustCenterEnvironmentInvalid is used when an unknown or unsupported value is provided.
	TrustCenterEnvironmentInvalid TrustCenterEnvironment = "TRUSTCENTERENVIRONMENT_INVALID"
)

var trustCenterEnvironmentValues = []TrustCenterEnvironment{TrustCenterEnvironmentLive, TrustCenterEnvironmentPreview}

// Values returns a slice of strings representing all valid TrustCenterEnvironment values.
func (TrustCenterEnvironment) Values() []string { return stringValues(trustCenterEnvironmentValues) }

// String returns the string representation of the TrustCenterEnvironment value.
func (r TrustCenterEnvironment) String() string { return string(r) }

// ToTrustCenterEnvironment converts a string to its corresponding TrustCenterEnvironment enum value.
func ToTrustCenterEnvironment(r string) *TrustCenterEnvironment {
	return parse(r, trustCenterEnvironmentValues, &TrustCenterEnvironmentInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TrustCenterEnvironment) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TrustCenterEnvironment) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
