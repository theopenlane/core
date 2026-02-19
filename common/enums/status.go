package enums

import "io"

type UserStatus string

var (
	UserStatusActive      UserStatus = "ACTIVE"
	UserStatusInactive    UserStatus = "INACTIVE"
	UserStatusDeactivated UserStatus = "DEACTIVATED"
	UserStatusSuspended   UserStatus = "SUSPENDED"
	UserStatusOnboarding  UserStatus = "ONBOARDING"
	UserStatusInvalid     UserStatus = "INVALID"
)

var userStatusValues = []UserStatus{UserStatusActive, UserStatusInactive, UserStatusDeactivated, UserStatusSuspended, UserStatusOnboarding}

// Values returns a slice of strings that represents all the possible values of the UserStatus enum.
// Possible default values are "ACTIVE", "INACTIVE", "DEACTIVATED", and "SUSPENDED".
func (UserStatus) Values() []string { return stringValues(userStatusValues) }

// String returns the UserStatus as a string
func (r UserStatus) String() string { return string(r) }

// ToUserStatus returns the user status enum based on string input
func ToUserStatus(r string) *UserStatus { return parse(r, userStatusValues, &UserStatusInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r UserStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *UserStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
