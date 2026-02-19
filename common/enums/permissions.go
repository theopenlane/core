package enums

import "io"

type Permission string

var (
	Editor  Permission = "EDITOR"
	Viewer  Permission = "VIEWER"
	Blocked Permission = "BLOCKED"
	Creator Permission = "CREATOR"
)

var permissionValues = []Permission{Editor, Viewer, Blocked, Creator}

// Values returns a slice of strings that represents all the possible values of the Permission enum.
// Possible default values are "EDITOR", "VIEWER", "BLOCKED", "CREATOR"
func (Permission) Values() []string { return stringValues(permissionValues) }

// String returns the permission as a string
func (r Permission) String() string { return string(r) }

// ToPermission returns the Permission based on string input
func ToPermission(r string) *Permission { return parse(r, permissionValues, nil) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Permission) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Permission) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
