package graphapi

import (
	"fmt"
	"net/url"
)

const (
	integrationWebhookPathPrefix = "/v1/integrations/webhook"
)

func integrationWebhookURL(scheme, host, integrationID, webhookName string) string {
	return (&url.URL{
		Scheme: scheme,
		Host:   host,
		Path: fmt.Sprintf(
			"%s/%s/%s",
			integrationWebhookPathPrefix,
			url.PathEscape(integrationID),
			url.PathEscape(webhookName),
		),
	}).String()
}
