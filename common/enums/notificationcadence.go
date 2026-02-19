package enums

import "io"

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

var notificationCadenceValues = []NotificationCadence{
	NotificationCadenceImmediate, NotificationCadenceDailyDigest, NotificationCadenceWeeklyDigest,
	NotificationCadenceMonthlyDigest, NotificationCadenceMute,
}

// Values returns a slice of strings that represents all the possible values of the NotificationCadence enum.
func (NotificationCadence) Values() []string { return stringValues(notificationCadenceValues) }

// String returns the cadence as a string.
func (r NotificationCadence) String() string { return string(r) }

// ToNotificationCadence returns the cadence enum based on string input.
func ToNotificationCadence(r string) *NotificationCadence {
	return parse(r, notificationCadenceValues, &NotificationCadenceInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationCadence) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationCadence) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
