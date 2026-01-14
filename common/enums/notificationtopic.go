package enums

import (
	"fmt"
	"io"
	"strings"
)

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
	// NotificationTopicInvalid represents an invalid notification topic
	NotificationTopicInvalid NotificationTopic = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the NotificationTopic enum.
func (NotificationTopic) Values() (kinds []string) {
	for _, s := range []NotificationTopic{
		NotificationTopicTaskAssignment,
		NotificationTopicApproval,
		NotificationTopicMention,
		NotificationTopicExport,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the NotificationTopic as a string
func (r NotificationTopic) String() string {
	return string(r)
}

// ToNotificationTopic returns the notification topic enum based on string input
func ToNotificationTopic(r string) *NotificationTopic {
	switch r := strings.ToUpper(r); r {
	case NotificationTopicTaskAssignment.String():
		return &NotificationTopicTaskAssignment
	case NotificationTopicApproval.String():
		return &NotificationTopicApproval
	case NotificationTopicMention.String():
		return &NotificationTopicMention
	case NotificationTopicExport.String():
		return &NotificationTopicExport
	default:
		return &NotificationTopicInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r NotificationTopic) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *NotificationTopic) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationTopic, got: %T", v) //nolint:err113
	}

	*r = NotificationTopic(str)

	return nil
}
