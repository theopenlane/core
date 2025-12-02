package enums

import (
	"fmt"
	"io"
	"strings"
)

type Visibility string

var (
	VisibilityPublic  Visibility = "PUBLIC"
	VisibilityPrivate Visibility = "PRIVATE"
	VisibilityInvalid Visibility = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the Visibility enum.
// Possible default values are "PUBLIC", and "PRIVATE".
func (Visibility) Values() (kinds []string) {
	for _, s := range []Visibility{VisibilityPublic, VisibilityPrivate} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the visibility as a string
func (r Visibility) String() string {
	return string(r)
}

// ToGroupVisibility returns the user status enum based on string input
func ToGroupVisibility(r string) *Visibility {
	switch r := strings.ToUpper(r); r {
	case VisibilityPublic.String():
		return &VisibilityPublic
	case VisibilityPrivate.String():
		return &VisibilityPrivate
	default:
		return &VisibilityInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Visibility) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Visibility) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Visibility, got: %T", v) //nolint:err113
	}

	*r = Visibility(str)

	return nil
}
