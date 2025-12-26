package enums

import (
	"fmt"
	"io"
	"strings"
)

type Region string

var (
	Amer          Region = "AMER"
	Emea          Region = "EMEA"
	Apac          Region = "APAC"
	InvalidRegion Region = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the Region enum.
// Possible default values are "AMER", "EMEA", and "APAC"
func (Region) Values() (kinds []string) {
	for _, s := range []Region{Amer, Emea, Apac} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the Region as a string
func (r Region) String() string {
	return string(r)
}

// ToRegion returns the database provider enum based on string input
func ToRegion(p string) *Region {
	switch p := strings.ToUpper(p); p {
	case Amer.String():
		return &Amer
	case Emea.String():
		return &Emea
	case Apac.String():
		return &Apac
	default:
		return &InvalidRegion
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Region) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Region) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Region, got: %T", v) //nolint:err113
	}

	*r = Region(str)

	return nil
}
