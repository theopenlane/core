package enums

import "io"

// NotificationTopic represents the topic/category of a notification
type NotificationTopic string

var (
	// NotificationTopicTaskAssignment represents a task assignment notification
	NotificationTopicTaskAssignment NotificationTopic = "TASK_ASSIGNMENT"
	// NotificationTopicApproval represents an approval notification
	NotificationTopicApproval NotificationTopic = "APPROVAL"
	// NotificationTopicMention represents a mention notification (comments)
	NotificationTopicMention NotificationTopic = "MENTION"
	// NotificationTopicExport represents an export notification
	NotificationTopicExport NotificationTopic = "EXPORT"
	// NotificationTopicStandardUpdate represents a standard update notification
	NotificationTopicStandardUpdate NotificationTopic = "STANDARD_UPDATE"
	// NotificationTopicInvalid represents an invalid notification topic
	NotificationTopicInvalid NotificationTopic = "INVALID"
)

var notificationTopicValues = []NotificationTopic{
	NotificationTopicTaskAssignment, NotificationTopicApproval,
	NotificationTopicMention, NotificationTopicExport,
	NotificationTopicStandardUpdate,
}

// Values returns a slice of strings that represents all the possible values of the NotificationTopic enum.
func (NotificationTopic) Values() []string { return stringValues(notificationTopicValues) }

// String returns the NotificationTopic as a string
func (r NotificationTopic) String() string { return string(r) }

// ToNotificationTopic returns the notification topic enum based on string input
func ToNotificationTopic(r string) *NotificationTopic {
	return parse(r, notificationTopicValues, &NotificationTopicInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r NotificationTopic) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *NotificationTopic) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
