package enums

import "io"

// TrustCenterControlVisibility is a custom type for control visibility on the trust center
type TrustCenterControlVisibility string

var (
	// TrustCenterControlVisibilityPubliclyVisible indicates that the control is publicly visible on the trust center
	TrustCenterControlVisibilityPubliclyVisible TrustCenterControlVisibility = "PUBLICLY_VISIBLE"
	// TrustCenterControlVisibilityNotVisible indicates that the control is not visible on the trust center
	TrustCenterControlVisibilityNotVisible TrustCenterControlVisibility = "NOT_VISIBLE"
	// TrustCenterControlVisibilityInvalid indicates that the control visibility is invalid
	TrustCenterControlVisibilityInvalid TrustCenterControlVisibility = "CONTROL_VISIBILITY_INVALID"
)

var trustCenterControlVisibilityValues = []TrustCenterControlVisibility{
	TrustCenterControlVisibilityPubliclyVisible,
	TrustCenterControlVisibilityNotVisible,
}

// Values returns a slice of strings that represents all the possible values of the TrustCenterControlVisibility enum.
func (TrustCenterControlVisibility) Values() []string {
	return stringValues(trustCenterControlVisibilityValues)
}

// String returns the control visibility as a string
func (r TrustCenterControlVisibility) String() string { return string(r) }

// ToTrustCenterControlVisibility returns the control visibility enum based on string input
func ToTrustCenterControlVisibility(r string) *TrustCenterControlVisibility {
	return parse(r, trustCenterControlVisibilityValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterControlVisibility) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterControlVisibility) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
