package enums

import (
	"fmt"
	"io"
	"strings"
)

type UserStatus string

var (
	UserStatusActive      UserStatus = "ACTIVE"
	UserStatusInactive    UserStatus = "INACTIVE"
	UserStatusDeactivated UserStatus = "DEACTIVATED"
	UserStatusSuspended   UserStatus = "SUSPENDED"
	UserStatusOnboarding  UserStatus = "ONBOARDING"
	UserStatusInvalid     UserStatus = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the UserStatus enum.
// Possible default values are "ACTIVE", "INACTIVE", "DEACTIVATED", and "SUSPENDED".
func (UserStatus) Values() (kinds []string) {
	for _, s := range []UserStatus{UserStatusActive, UserStatusInactive, UserStatusDeactivated, UserStatusSuspended, UserStatusOnboarding} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the UserStatus as a string
func (r UserStatus) String() string {
	return string(r)
}

// ToUserStatus returns the user status enum based on string input
func ToUserStatus(r string) *UserStatus {
	switch r := strings.ToUpper(r); r {
	case UserStatusActive.String():
		return &UserStatusActive
	case UserStatusInactive.String():
		return &UserStatusInactive
	case UserStatusDeactivated.String():
		return &UserStatusDeactivated
	case UserStatusSuspended.String():
		return &UserStatusSuspended
	case UserStatusOnboarding.String():
		return &UserStatusOnboarding
	default:
		return &UserStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r UserStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *UserStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for UserStatus, got: %T", v) //nolint:err113
	}

	*r = UserStatus(str)

	return nil
}
