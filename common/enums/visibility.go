package enums

import "io"

type Visibility string

var (
	VisibilityPublic  Visibility = "PUBLIC"
	VisibilityPrivate Visibility = "PRIVATE"
	VisibilityInvalid Visibility = "INVALID"
)

var visibilityValues = []Visibility{VisibilityPublic, VisibilityPrivate}

// Values returns a slice of strings that represents all the possible values of the Visibility enum.
// Possible default values are "PUBLIC", and "PRIVATE".
func (Visibility) Values() []string { return stringValues(visibilityValues) }

// String returns the visibility as a string
func (r Visibility) String() string { return string(r) }

// ToGroupVisibility returns the user status enum based on string input
func ToGroupVisibility(r string) *Visibility {
	return parse(r, visibilityValues, &VisibilityInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Visibility) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Visibility) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
