package enums

import (
	"fmt"
	"io"
	"strings"
)

type CustomDomainStatus string

var (
	CustomDomainStatusVerified     CustomDomainStatus = "VERIFIED"
	CustomDomainStatusFailedVerify CustomDomainStatus = "FAILED_VERIFY"
	CustomDomainStatusPending      CustomDomainStatus = "PENDING"
	CustomDomainStatusInvalid      CustomDomainStatus = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the CustomDomainStatus enum.
// Possible default values are "ACTIVE", "INACTIVE", "DEACTIVATED", and "SUSPENDED".
func (CustomDomainStatus) Values() (kinds []string) {
	for _, s := range []CustomDomainStatus{CustomDomainStatusInvalid, CustomDomainStatusVerified, CustomDomainStatusFailedVerify, CustomDomainStatusPending} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the CustomDomainStatus as a string
func (r CustomDomainStatus) String() string {
	return string(r)
}

// ToCustomDomainStatus returns the user status enum based on string input
func ToCustomDomainStatus(r string) *CustomDomainStatus {
	switch r := strings.ToUpper(r); r {
	case CustomDomainStatusVerified.String():
		return &CustomDomainStatusVerified
	case CustomDomainStatusFailedVerify.String():
		return &CustomDomainStatusFailedVerify
	case CustomDomainStatusPending.String():
		return &CustomDomainStatusPending
	default:
		return &CustomDomainStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r CustomDomainStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *CustomDomainStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for CustomDomainStatus, got: %T", v) //nolint:err113
	}

	*r = CustomDomainStatus(str)

	return nil
}
