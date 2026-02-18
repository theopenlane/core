package enums

import "io"

// Role is the role of a user in an organization.
type Role string

var (
	RoleOwner   Role = "OWNER"
	RoleAdmin   Role = "ADMIN"
	RoleMember  Role = "MEMBER"
	RoleUser    Role = "USER"
	RoleInvalid Role = "INVALID"
)

// roleSchemaValues are the values exposed to ent schemas.
var roleSchemaValues = []Role{RoleAdmin, RoleMember}

// roleParseValues are all values accepted by ToRole.
var roleParseValues = []Role{RoleOwner, RoleAdmin, RoleMember, RoleUser}

// Values returns a slice of strings that represents all the possible values of the Role enum.
// Possible default values are "ADMIN", "MEMBER"
func (Role) Values() []string { return stringValues(roleSchemaValues) }

// String returns the role as a string
func (r Role) String() string { return string(r) }

// ToRole returns the Role based on string input
func ToRole(r string) *Role { return parse(r, roleParseValues, &RoleInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Role) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Role) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
