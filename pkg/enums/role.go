package enums

import (
	"fmt"
	"io"
	"strings"
)

type Role string

var (
	RoleOwner   Role = "OWNER"
	RoleAdmin   Role = "ADMIN"
	RoleMember  Role = "MEMBER"
	RoleUser    Role = "USER"
	RoleInvalid Role = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the Role enum.
// Possible default values are "ADMIN", "MEMBER"
func (Role) Values() (kinds []string) {
	for _, s := range []Role{RoleAdmin, RoleMember} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the role as a string
func (r Role) String() string {
	return string(r)
}

// ToRole returns the Role based on string input
func ToRole(r string) *Role {
	switch r := strings.ToUpper(r); r {
	case RoleOwner.String():
		return &RoleOwner
	case RoleAdmin.String():
		return &RoleAdmin
	case RoleMember.String():
		return &RoleMember
	case RoleUser.String():
		return &RoleUser
	default:
		return &RoleInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Role) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Role, got: %T", v) //nolint:err113
	}

	*r = Role(str)

	return nil
}
