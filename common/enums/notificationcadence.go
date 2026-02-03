package enums

import (
	"fmt"
	"io"
	"strings"
)

// NotificationCadence represents how often a notification should be delivered.
type NotificationCadence string

var (
	// NotificationCadenceImmediate delivers notifications immediately.
	NotificationCadenceImmediate NotificationCadence = "IMMEDIATE"
	// NotificationCadenceDailyDigest delivers notifications in a daily digest.
	NotificationCadenceDailyDigest NotificationCadence = "DAILY_DIGEST"
	// NotificationCadenceWeeklyDigest delivers notifications in a weekly digest.
	NotificationCadenceWeeklyDigest NotificationCadence = "WEEKLY_DIGEST"
	// NotificationCadenceMonthlyDigest delivers notifications in a monthly digest.
	NotificationCadenceMonthlyDigest NotificationCadence = "MONTHLY_DIGEST"
	// NotificationCadenceMute disables notifications for the preference.
	NotificationCadenceMute NotificationCadence = "MUTE"
	// NotificationCadenceInvalid represents an invalid cadence.
	NotificationCadenceInvalid NotificationCadence = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the NotificationCadence enum.
func (NotificationCadence) Values() []string {
	return []string{
		NotificationCadenceImmediate.String(),
		NotificationCadenceDailyDigest.String(),
		NotificationCadenceWeeklyDigest.String(),
		NotificationCadenceMonthlyDigest.String(),
		NotificationCadenceMute.String(),
	}
}

// String returns the cadence as a string.
func (r NotificationCadence) String() string {
	return string(r)
}

// ToNotificationCadence returns the cadence enum based on string input.
func ToNotificationCadence(r string) *NotificationCadence {
	switch strings.ToUpper(r) {
	case NotificationCadenceImmediate.String():
		return &NotificationCadenceImmediate
	case NotificationCadenceDailyDigest.String():
		return &NotificationCadenceDailyDigest
	case NotificationCadenceWeeklyDigest.String():
		return &NotificationCadenceWeeklyDigest
	case NotificationCadenceMonthlyDigest.String():
		return &NotificationCadenceMonthlyDigest
	case NotificationCadenceMute.String():
		return &NotificationCadenceMute
	default:
		return &NotificationCadenceInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationCadence) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationCadence) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationCadence, got: %T", v) //nolint:err113
	}

	*r = NotificationCadence(str)

	return nil
}
