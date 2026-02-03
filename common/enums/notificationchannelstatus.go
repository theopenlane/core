package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the NotificationChannelStatus enum.
func (NotificationChannelStatus) Values() []string {
	return []string{
		NotificationChannelStatusEnabled.String(),
		NotificationChannelStatusDisabled.String(),
		NotificationChannelStatusPending.String(),
		NotificationChannelStatusVerified.String(),
		NotificationChannelStatusError.String(),
	}
}

// String returns the status as a string.
func (r NotificationChannelStatus) String() string {
	return string(r)
}

// ToNotificationChannelStatus returns the status enum based on string input.
func ToNotificationChannelStatus(r string) *NotificationChannelStatus {
	switch strings.ToUpper(r) {
	case NotificationChannelStatusEnabled.String():
		return &NotificationChannelStatusEnabled
	case NotificationChannelStatusDisabled.String():
		return &NotificationChannelStatusDisabled
	case NotificationChannelStatusPending.String():
		return &NotificationChannelStatusPending
	case NotificationChannelStatusVerified.String():
		return &NotificationChannelStatusVerified
	case NotificationChannelStatusError.String():
		return &NotificationChannelStatusError
	default:
		return &NotificationChannelStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationChannelStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationChannelStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationChannelStatus, got: %T", v) //nolint:err113
	}

	*r = NotificationChannelStatus(str)

	return nil
}
