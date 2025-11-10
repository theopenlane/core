package enums

import (
	"fmt"
	"io"
	"strings"
)

// NotificationChannel is a custom type representing the various channels for notifications.
type NotificationChannel string

var (
	// NotificationChannelInApp indicates an in-app notification.
	NotificationChannelInApp NotificationChannel = "IN_APP"
	// NotificationChannelSlack indicates a Slack notification.
	NotificationChannelSlack NotificationChannel = "SLACK"
	// NotificationChannelEmail indicates an email notification.
	NotificationChannelEmail NotificationChannel = "EMAIL"
	// NotificationChannelInvalid is used when an unknown or unsupported value is provided.
	NotificationChannelInvalid NotificationChannel = "NOTIFICATIONCHANNEL_INVALID"
)

// Values returns a slice of strings representing all valid NotificationChannel values.
func (NotificationChannel) Values() []string {
	return []string{
		string(NotificationChannelInApp),
		string(NotificationChannelSlack),
		string(NotificationChannelEmail),
	}
}

// String returns the string representation of the NotificationChannel value.
func (r NotificationChannel) String() string {
	return string(r)
}

// ToNotificationChannel converts a string to its corresponding NotificationChannel enum value.
func ToNotificationChannel(r string) *NotificationChannel {
	switch strings.ToUpper(r) {
	case NotificationChannelInApp.String():
		return &NotificationChannelInApp
	case NotificationChannelSlack.String():
		return &NotificationChannelSlack
	case NotificationChannelEmail.String():
		return &NotificationChannelEmail
	default:
		return &NotificationChannelInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationChannel) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationChannel) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationChannel, got: %T", v) //nolint:err113
	}

	*r = NotificationChannel(str)

	return nil
}
