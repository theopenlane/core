package graphapi

import (
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
)

const (
	genericIntegrationWebhookPathPrefix = "/v1/integrations/webhook"
	gitHubAppWebhookPath                = "/v1/github/app/webhook"
)

func integrationWebhookURL(scheme, host, definitionID, integrationID, webhookName string) string {
	return (&url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   integrationWebhookPath(definitionID, integrationID, webhookName),
	}).String()
}

func integrationWebhookPath(definitionID, integrationID, webhookName string) string {
	if definitionID == githubapp.DefinitionID.ID() {
		return gitHubAppWebhookPath
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		genericIntegrationWebhookPathPrefix,
		url.PathEscape(integrationID),
		url.PathEscape(webhookName),
	)
}
