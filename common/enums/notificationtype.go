package enums

import "io"

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

var notificationTypeValues = []NotificationType{NotificationTypeOrganization, NotificationTypeUser}

// Values returns a slice of strings that represents all the possible values of the NotificationType enum.
// Possible default values are "ORGANIZATION" and "USER".
func (NotificationType) Values() []string { return stringValues(notificationTypeValues) }

// String returns the NotificationType as a string
func (r NotificationType) String() string { return string(r) }

// ToNotificationType returns the notification type enum based on string input
func ToNotificationType(r string) *NotificationType {
	return parse(r, notificationTypeValues, &NotificationTypeInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r NotificationType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *NotificationType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
