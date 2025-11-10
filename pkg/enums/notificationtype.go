package enums

import (
	"fmt"
	"io"
	"strings"
)

// NotificationType is a custom type representing the various types of notifications.
type NotificationType string

var (
	// NotificationTypeOrganization indicates an organization-level notification.
	NotificationTypeOrganization NotificationType = "ORGANIZATION"
	// NotificationTypeUser indicates a user-level notification.
	NotificationTypeUser NotificationType = "USER"
	// NotificationTypeInvalid is used when an unknown or unsupported value is provided.
	NotificationTypeInvalid NotificationType = "NOTIFICATIONTYPE_INVALID"
)

// Values returns a slice of strings representing all valid NotificationType values.
func (NotificationType) Values() []string {
	return []string{
		string(NotificationTypeOrganization),
		string(NotificationTypeUser),
	}
}

// String returns the string representation of the NotificationType value.
func (r NotificationType) String() string {
	return string(r)
}

// ToNotificationType converts a string to its corresponding NotificationType enum value.
func ToNotificationType(r string) *NotificationType {
	switch strings.ToUpper(r) {
	case NotificationTypeOrganization.String():
		return &NotificationTypeOrganization
	case NotificationTypeUser.String():
		return &NotificationTypeUser
	default:
		return &NotificationTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationType, got: %T", v) //nolint:err113
	}

	*r = NotificationType(str)

	return nil
}
