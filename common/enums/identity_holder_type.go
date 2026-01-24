package enums

import (
	"fmt"
	"io"
	"strings"
)

// IdentityHolderType is a custom type representing the identity holder classification
type IdentityHolderType string

var (
	// IdentityHolderTypeEmployee indicates an employee
	IdentityHolderTypeEmployee IdentityHolderType = "EMPLOYEE"
	// IdentityHolderTypeContractor indicates a contractor
	IdentityHolderTypeContractor IdentityHolderType = "CONTRACTOR"
	// IdentityHolderTypeInvalid is used when an unknown or unsupported value is provided
	IdentityHolderTypeInvalid IdentityHolderType = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the IdentityHolderType enum
// Possible default values are "EMPLOYEE" and "CONTRACTOR"
func (IdentityHolderType) Values() (kinds []string) {
	for _, s := range []IdentityHolderType{IdentityHolderTypeEmployee, IdentityHolderTypeContractor} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the IdentityHolderType as a string
func (r IdentityHolderType) String() string {
	return string(r)
}

// ToIdentityHolderType returns the identity holder type enum based on string input
func ToIdentityHolderType(r string) *IdentityHolderType {
	switch r := strings.ToUpper(r); r {
	case IdentityHolderTypeEmployee.String():
		return &IdentityHolderTypeEmployee
	case IdentityHolderTypeContractor.String():
		return &IdentityHolderTypeContractor
	default:
		return &IdentityHolderTypeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r IdentityHolderType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *IdentityHolderType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for IdentityHolderType, got: %T", v) //nolint:err113
	}

	*r = IdentityHolderType(str)

	return nil
}
