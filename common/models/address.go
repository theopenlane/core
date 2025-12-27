package models

import (
	"io"
	"strings"
)

// Address is a custom type for Address
type Address struct {
	// Line1 is the first line of the address
	Line1 string `json:"line1"`
	// Line2 is the second line of the address
	Line2 string `json:"line2"`
	// City is the city of the address
	City string `json:"city"`
	// State is the state of the address
	State string `json:"state"`
	// PostalCode is the postal code of the address
	PostalCode string `json:"postalCode"`
	// Country is the country of the address
	Country string `json:"country"`
}

// String returns a string representation of the address
func (a Address) String() string {
	if a == (Address{}) {
		return ""
	}

	line1 := a.Line1 + " " + a.Line2 + " " + a.City
	line2 := a.State + " " + a.PostalCode + " " + a.Country

	if strings.TrimSpace(line1) == "" {
		return strings.TrimSpace(line2)
	}

	return strings.TrimSpace(line1 + ", " + line2)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (a Address) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, a)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (a *Address) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, a)
}
