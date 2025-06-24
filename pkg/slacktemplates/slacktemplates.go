package slacktemplates

import "embed"

const (
	SubscriberTemplateName = "new_subscriber.tmpl"
	UserTemplateName       = "new_user.tmpl"
)

// Templates holds embedded Slack notification templates.
//
//go:embed templates/*.tmpl
var Templates embed.FS
