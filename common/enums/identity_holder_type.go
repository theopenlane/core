package enums

import "io"

// IdentityHolderType is a custom type representing the identity holder classification
type IdentityHolderType string

var (
	// IdentityHolderTypeEmployee indicates an employee
	IdentityHolderTypeEmployee IdentityHolderType = "EMPLOYEE"
	// IdentityHolderTypeContractor indicates a contractor
	IdentityHolderTypeContractor IdentityHolderType = "CONTRACTOR"
	// IdentityHolderTypeInvalid is used when an unsupported value is provided
	IdentityHolderTypeInvalid IdentityHolderType = "INVALID"
	// IdentityHolderTypeUnspecified is used when the identity holder type is unknown or not provided
	IdentityHolderTypeUnspecified IdentityHolderType = "UNSPECIFIED"
	// IdentityHolderTypeIntern indicates an intern / student worker
	IdentityHolderTypeIntern IdentityHolderType = "INTERN"
	// IdentityHolderTypeService indicates a service identity
	IdentityHolderTypeService IdentityHolderType = "SERVICE"
	// IdentityHolderTypePartner indicates a partner identity
	IdentityHolderTypePartner IdentityHolderType = "PARTNER"
)

var identityHolderTypeValues = []IdentityHolderType{IdentityHolderTypeEmployee, IdentityHolderTypeContractor, IdentityHolderTypeUnspecified, IdentityHolderTypeIntern, IdentityHolderTypeService, IdentityHolderTypePartner}

// Values returns a slice of strings that represents all the possible values of the IdentityHolderType enum
// Possible default values are "EMPLOYEE" and "CONTRACTOR"
func (IdentityHolderType) Values() []string { return stringValues(identityHolderTypeValues) }

// String returns the IdentityHolderType as a string
func (r IdentityHolderType) String() string { return string(r) }

// ToIdentityHolderType returns the identity holder type enum based on string input
func ToIdentityHolderType(r string) *IdentityHolderType {
	return parse(r, identityHolderTypeValues, &IdentityHolderTypeInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r IdentityHolderType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *IdentityHolderType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
