package slacktemplates

import "embed"

const (
	// Embed paths are relative to the directory containing this file. When
	// parsing templates using the embedded filesystem the full path must be
	// provided, so include the templates directory prefix here.
	SubscriberTemplateName     = "templates/new_subscriber.tmpl"
	UserTemplateName           = "templates/new_user.tmpl"
	GalaSubscriberTemplateName = "templates/new_subscriber_gala.tmpl"
	GalaUserTemplateName       = "templates/new_user_gala.tmpl"
)

// Templates holds embedded Slack notification templates.
//
//go:embed templates/*.tmpl
var Templates embed.FS
