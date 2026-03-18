package enums

import "io"

type TrustCenterThemeMode string

var (
	// TrustCenterThemeModeEasy is the easy theme mode
	TrustCenterThemeModeEasy TrustCenterThemeMode = "EASY"
	// TrustCenterThemeModeAdvanced is the advanced theme mode
	TrustCenterThemeModeAdvanced TrustCenterThemeMode = "ADVANCED"
	// TrustCenterThemeModeInvalid is the invalid theme mode
	TrustCenterThemeModeInvalid TrustCenterThemeMode = "INVALID"
)

var trustCenterThemeModeValues = []TrustCenterThemeMode{TrustCenterThemeModeEasy, TrustCenterThemeModeAdvanced}

// Values returns a slice of strings that represents all the possible values of the TrustCenterThemeMode enum.
// Possible default values are "EASY" and "ADVANCED"
func (TrustCenterThemeMode) Values() []string { return stringValues(trustCenterThemeModeValues) }

// String returns the TrustCenterThemeMode as a string
func (r TrustCenterThemeMode) String() string { return string(r) }

// ToTrustCenterThemeMode returns the trust center theme mode enum based on string input
func ToTrustCenterThemeMode(r string) *TrustCenterThemeMode {
	return parse(r, trustCenterThemeModeValues, &TrustCenterThemeModeInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterThemeMode) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterThemeMode) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
