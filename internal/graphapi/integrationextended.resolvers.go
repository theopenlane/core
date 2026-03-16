package graphapi

// This file will be automatically regenerated based on the schema, any resolver
// implementations will be copied through when generating and any unknown code
// will be moved to the end.

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
)

const (
	genericIntegrationWebhookPathPrefix = "/v1/integrations/webhook"
	gitHubAppWebhookPath                = "/v1/github/app/webhook"
)

// WebhookURLs is the resolver for the webhookURLs field.
func (r *integrationResolver) WebhookURLs(ctx context.Context, obj *generated.Integration) (map[string]any, error) {
	if r.integrationsRuntime == nil {
		return nil, nil
	}

	def, ok := r.integrationsRuntime.Definition(obj.DefinitionID)
	if !ok {
		return nil, nil
	}

	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return nil, nil
	}

	webhookURLs := make(map[string]any, len(def.Webhooks))
	for _, webhook := range def.Webhooks {
		webhookURLs[webhook.Name] = integrationWebhookURL(ec.Scheme(), ec.Request().Host, obj.DefinitionID, obj.ID, webhook.Name)
	}

	return webhookURLs, nil
}

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
