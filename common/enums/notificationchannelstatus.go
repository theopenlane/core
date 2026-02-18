package enums

import "io"

// NotificationChannelStatus represents the configuration status for a notification channel.
type NotificationChannelStatus string

var (
	// NotificationChannelStatusEnabled indicates the channel is active.
	NotificationChannelStatusEnabled NotificationChannelStatus = "ENABLED"
	// NotificationChannelStatusDisabled indicates the channel is disabled.
	NotificationChannelStatusDisabled NotificationChannelStatus = "DISABLED"
	// NotificationChannelStatusPending indicates the channel is pending verification.
	NotificationChannelStatusPending NotificationChannelStatus = "PENDING"
	// NotificationChannelStatusVerified indicates the channel is verified and ready.
	NotificationChannelStatusVerified NotificationChannelStatus = "VERIFIED"
	// NotificationChannelStatusError indicates the channel is in an error state.
	NotificationChannelStatusError NotificationChannelStatus = "ERROR"
	// NotificationChannelStatusInvalid represents an invalid status.
	NotificationChannelStatusInvalid NotificationChannelStatus = "INVALID"
)

var notificationChannelStatusValues = []NotificationChannelStatus{
	NotificationChannelStatusEnabled, NotificationChannelStatusDisabled, NotificationChannelStatusPending,
	NotificationChannelStatusVerified, NotificationChannelStatusError,
}

// Values returns a slice of strings that represents all the possible values of the NotificationChannelStatus enum.
func (NotificationChannelStatus) Values() []string {
	return stringValues(notificationChannelStatusValues)
}

// String returns the status as a string.
func (r NotificationChannelStatus) String() string { return string(r) }

// ToNotificationChannelStatus returns the status enum based on string input.
func ToNotificationChannelStatus(r string) *NotificationChannelStatus {
	return parse(r, notificationChannelStatusValues, &NotificationChannelStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationChannelStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationChannelStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
