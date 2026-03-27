package enums

import "io"

// NotificationTopic is a custom type representing the various states of NotificationTopic.
type NotificationTopic string

var (
	// NotificationTopicTaskAssignment indicates the task assignment.
	NotificationTopicTaskAssignment NotificationTopic = "TASK_ASSIGNMENT"
	// NotificationTopicApproval indicates the approval.
	NotificationTopicApproval NotificationTopic = "APPROVAL"
	// NotificationTopicMention indicates the mention.
	NotificationTopicMention NotificationTopic = "MENTION"
	// NotificationTopicExport indicates the export.
	NotificationTopicExport NotificationTopic = "EXPORT"
	// NotificationTopicStandardUpdate indicates the standard update.
	NotificationTopicStandardUpdate NotificationTopic = "STANDARD_UPDATE"
	// NotificationTopicDomainScan indicates the domain scan.
	NotificationTopicDomainScan NotificationTopic = "DOMAIN_SCAN"
	// NotificationTopicInvalid is used when an unknown or unsupported value is provided.
	NotificationTopicInvalid NotificationTopic = "NOTIFICATIONTOPIC_INVALID"
)

var notificationTopicValues = []NotificationTopic{
	NotificationTopicTaskAssignment,
	NotificationTopicApproval,
	NotificationTopicMention,
	NotificationTopicExport,
	NotificationTopicStandardUpdate,
	NotificationTopicDomainScan,
}

// Values returns a slice of strings representing all valid NotificationTopic values.
func (NotificationTopic) Values() []string { return stringValues(notificationTopicValues) }

// String returns the string representation of the NotificationTopic value.
func (r NotificationTopic) String() string { return string(r) }

// ToNotificationTopic converts a string to its corresponding NotificationTopic enum value.
func ToNotificationTopic(r string) *NotificationTopic { return parse(r, notificationTopicValues, &NotificationTopicInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationTopic) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationTopic) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
