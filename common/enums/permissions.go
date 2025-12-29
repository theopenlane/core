package enums

import (
	"fmt"
	"io"
	"strings"
)

type Permission string

var (
	Editor  Permission = "EDITOR"
	Viewer  Permission = "VIEWER"
	Blocked Permission = "BLOCKED"
	Creator Permission = "CREATOR"
)

// Values returns a slice of strings that represents all the possible values of the Permission enum.
// Possible default values are "EDITOR", "VIEWER", "BLOCKED", "CREATOR"
func (Permission) Values() (kinds []string) {
	for _, s := range []Permission{Editor, Viewer, Blocked, Creator} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the permission as a string
func (r Permission) String() string {
	return string(r)
}

// ToPermission returns the Permission based on string input
func ToPermission(r string) *Permission {
	switch r := strings.ToUpper(r); r {
	case Editor.String():
		return &Editor
	case Viewer.String():
		return &Viewer
	case Blocked.String():
		return &Blocked
	case Creator.String():
		return &Creator
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Permission) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Permission) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Permission, got: %T", v) //nolint:err113
	}

	*r = Permission(str)

	return nil
}
