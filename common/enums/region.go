package enums

import "io"

type Region string

var (
	Amer          Region = "AMER"
	Emea          Region = "EMEA"
	Apac          Region = "APAC"
	InvalidRegion Region = "INVALID"
)

var regionValues = []Region{Amer, Emea, Apac}

// Values returns a slice of strings that represents all the possible values of the Region enum.
// Possible default values are "AMER", "EMEA", and "APAC"
func (Region) Values() []string { return stringValues(regionValues) }

// String returns the Region as a string
func (r Region) String() string { return string(r) }

// ToRegion returns the database provider enum based on string input
func ToRegion(p string) *Region { return parse(p, regionValues, &InvalidRegion) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Region) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Region) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
