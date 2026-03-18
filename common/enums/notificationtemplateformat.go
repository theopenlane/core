package enums

import "io"

// NotificationTemplateFormat represents the format used to render a notification template.
type NotificationTemplateFormat string

var (
	// NotificationTemplateFormatText indicates plain text templates.
	NotificationTemplateFormatText NotificationTemplateFormat = "TEXT"
	// NotificationTemplateFormatMarkdown indicates markdown templates.
	NotificationTemplateFormatMarkdown NotificationTemplateFormat = "MARKDOWN"
	// NotificationTemplateFormatHTML indicates HTML templates.
	NotificationTemplateFormatHTML NotificationTemplateFormat = "HTML"
	// NotificationTemplateFormatJSON indicates structured JSON templates (e.g., Slack blocks).
	NotificationTemplateFormatJSON NotificationTemplateFormat = "JSON"
	// NotificationTemplateFormatInvalid represents an invalid template format.
	NotificationTemplateFormatInvalid NotificationTemplateFormat = "INVALID"
)

var notificationTemplateFormatValues = []NotificationTemplateFormat{
	NotificationTemplateFormatText, NotificationTemplateFormatMarkdown,
	NotificationTemplateFormatHTML, NotificationTemplateFormatJSON,
}

// Values returns a slice of strings that represents all the possible values of the NotificationTemplateFormat enum.
func (NotificationTemplateFormat) Values() []string {
	return stringValues(notificationTemplateFormatValues)
}

// String returns the template format as a string.
func (r NotificationTemplateFormat) String() string { return string(r) }

// ToNotificationTemplateFormat returns the template format enum based on string input.
func ToNotificationTemplateFormat(r string) *NotificationTemplateFormat {
	return parse(r, notificationTemplateFormatValues, &NotificationTemplateFormatInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationTemplateFormat) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationTemplateFormat) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
