package enums

import (
	"fmt"
	"io"
	"strings"
)

type TrustCenterThemeMode string

var (
	// TrustCenterThemeModeEasy is the easy theme mode
	TrustCenterThemeModeEasy TrustCenterThemeMode = "EASY"
	// TrustCenterThemeModeAdvanced is the advanced theme mode
	TrustCenterThemeModeAdvanced TrustCenterThemeMode = "ADVANCED"
	// TrustCenterThemeModeInvalid is the invalid theme mode
	TrustCenterThemeModeInvalid TrustCenterThemeMode = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the TrustCenterThemeMode enum.
// Possible default values are "EASY" and "ADVANCED"
func (TrustCenterThemeMode) Values() (kinds []string) {
	for _, s := range []TrustCenterThemeMode{TrustCenterThemeModeEasy, TrustCenterThemeModeAdvanced} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the TrustCenterThemeMode as a string
func (r TrustCenterThemeMode) String() string {
	return string(r)
}

// ToTrustCenterThemeMode returns the trust center theme mode enum based on string input
func ToTrustCenterThemeMode(r string) *TrustCenterThemeMode {
	switch r := strings.ToUpper(r); r {
	case TrustCenterThemeModeEasy.String():
		return &TrustCenterThemeModeEasy
	case TrustCenterThemeModeAdvanced.String():
		return &TrustCenterThemeModeAdvanced
	default:
		return &TrustCenterThemeModeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterThemeMode) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterThemeMode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TrustCenterThemeMode, got: %T", v) //nolint:err113
	}

	*r = TrustCenterThemeMode(str)

	return nil
}
