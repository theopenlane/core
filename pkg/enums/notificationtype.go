package enums

import (
	"fmt"
	"io"
	"strings"
)

// NotificationType represents the type of notification
type NotificationType string

var (
	// NotificationTypeOrganization represents an organization notification
	NotificationTypeOrganization NotificationType = "ORGANIZATION"
	// NotificationTypeUser represents a user notification
	NotificationTypeUser NotificationType = "USER"
	// NotificationTypeInvalid represents an invalid notification type
	NotificationTypeInvalid NotificationType = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the NotificationType enum.
// Possible default values are "ORGANIZATION" and "USER".
func (NotificationType) Values() (kinds []string) {
	for _, s := range []NotificationType{NotificationTypeOrganization, NotificationTypeUser} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the NotificationType as a string
func (r NotificationType) String() string {
	return string(r)
}

// ToNotificationType returns the notification type enum based on string input
func ToNotificationType(r string) *NotificationType {
	switch r := strings.ToUpper(r); r {
	case NotificationTypeOrganization.String():
		return &NotificationTypeOrganization
	case NotificationTypeUser.String():
		return &NotificationTypeUser
	default:
		return &NotificationTypeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r NotificationType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *NotificationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationType, got: %T", v) //nolint:err113
	}

	*r = NotificationType(str)

	return nil
}
