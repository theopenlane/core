package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the NotificationTemplateFormat enum.
func (NotificationTemplateFormat) Values() []string {
	return []string{
		NotificationTemplateFormatText.String(),
		NotificationTemplateFormatMarkdown.String(),
		NotificationTemplateFormatHTML.String(),
		NotificationTemplateFormatJSON.String(),
	}
}

// String returns the template format as a string.
func (r NotificationTemplateFormat) String() string {
	return string(r)
}

// ToNotificationTemplateFormat returns the template format enum based on string input.
func ToNotificationTemplateFormat(r string) *NotificationTemplateFormat {
	switch strings.ToUpper(r) {
	case NotificationTemplateFormatText.String():
		return &NotificationTemplateFormatText
	case NotificationTemplateFormatMarkdown.String():
		return &NotificationTemplateFormatMarkdown
	case NotificationTemplateFormatHTML.String():
		return &NotificationTemplateFormatHTML
	case NotificationTemplateFormatJSON.String():
		return &NotificationTemplateFormatJSON
	default:
		return &NotificationTemplateFormatInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r NotificationTemplateFormat) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *NotificationTemplateFormat) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for NotificationTemplateFormat, got: %T", v) //nolint:err113
	}

	*r = NotificationTemplateFormat(str)

	return nil
}
