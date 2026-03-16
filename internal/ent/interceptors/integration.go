package interceptors

import (
	"context"
	"net/url"
	"reflect"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/gqlgen-plugins/graphutils"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	integrationregistry "github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// Keep these path shapes aligned with the registered inbound integration routes.
	genericIntegrationWebhookPathPrefix = "/v1/integrations/webhook"
	gitHubAppWebhookPath                = "/github/app/webhook"
)

// InterceptorIntegrationWebhookURLs lazily enriches Integration responses with
// fully-qualified inbound webhook URLs when the GraphQL field is selected.
func InterceptorIntegrationWebhookURLs() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.IntegrationFunc(func(ctx context.Context, q *generated.IntegrationQuery) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("InterceptorIntegrationWebhookURLs")

			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if !graphutils.CheckForRequestedField(ctx, "webhookURLs") {
				return v, nil
			}

			reg := integrationRegistryFromQuery(q)
			if reg == nil {
				logx.FromContext(ctx).Warn().Msg("integration registry is nil, skipping integration webhook URL enrichment")

				return v, nil
			}

			baseURL, ok := requestBaseURL(ctx)
			if !ok {
				logx.FromContext(ctx).Warn().Msg("request base URL unavailable, skipping integration webhook URL enrichment")

				return v, nil
			}

			switch res := v.(type) {
			case []*generated.Integration:
				for _, integration := range res {
					setIntegrationWebhookURLs(integration, buildWebhookURLs(integration, reg, baseURL))
				}
			case *generated.Integration:
				setIntegrationWebhookURLs(res, buildWebhookURLs(res, reg, baseURL))
			}

			return v, nil
		})
	})
}

func integrationRegistryFromQuery(q *generated.IntegrationQuery) *integrationregistry.Registry {
	if q == nil {
		return nil
	}

	field := reflect.ValueOf(q).Elem().FieldByName("IntegrationRegistry")
	if !field.IsValid() || field.IsNil() || !field.CanInterface() {
		return nil
	}

	reg, ok := field.Interface().(*integrationregistry.Registry)
	if !ok {
		return nil
	}

	return reg
}

func requestBaseURL(ctx context.Context) (*url.URL, bool) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return nil, false
	}

	req := ec.Request()
	if req == nil {
		return nil, false
	}

	scheme := forwardedHeaderValue(req.Header.Get("X-Forwarded-Proto"))
	if scheme == "" && req.URL != nil {
		scheme = req.URL.Scheme
	}
	if scheme == "" {
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := forwardedHeaderValue(req.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = req.Host
	}
	if host == "" && req.URL != nil {
		host = req.URL.Host
	}
	if host == "" {
		return nil, false
	}

	prefix := normalizeForwardedPrefix(forwardedHeaderValue(req.Header.Get("X-Forwarded-Prefix")))

	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   prefix,
	}, true
}

func buildWebhookURLs(integration *generated.Integration, reg *integrationregistry.Registry, baseURL *url.URL) map[string]any {
	if integration == nil || reg == nil || baseURL == nil || integration.DefinitionID == "" {
		return nil
	}

	def, ok := reg.Definition(integration.DefinitionID)
	if !ok {
		return nil
	}

	urls := make(map[string]any, len(def.Webhooks))
	for _, webhook := range def.Webhooks {
		urls[webhook.Name] = buildWebhookURL(baseURL, integration.DefinitionID, integration.ID, webhook.Name)
	}

	return urls
}

func buildWebhookURL(baseURL *url.URL, definitionID, integrationID, webhookName string) string {
	if baseURL == nil {
		return ""
	}

	clone := *baseURL
	clone.RawQuery = ""
	clone.Fragment = ""
	clone.Path = joinURLPath(baseURL.Path, webhookPath(definitionID, integrationID, webhookName))

	return clone.String()
}

func webhookPath(definitionID, integrationID, webhookName string) string {
	// GitHub App uses a provider-global inbound route today rather than the generic
	// installation-scoped webhook route.
	if definitionID == githubapp.DefinitionID.ID() {
		return gitHubAppWebhookPath
	}

	return joinURLPath(
		genericIntegrationWebhookPathPrefix,
		url.PathEscape(integrationID),
		url.PathEscape(webhookName),
	)
}

func setIntegrationWebhookURLs(integration *generated.Integration, webhookURLs map[string]any) {
	if integration == nil {
		return
	}

	field := reflect.ValueOf(integration).Elem().FieldByName("WebhookURLs")
	if !field.IsValid() || !field.CanSet() {
		return
	}

	if webhookURLs == nil {
		field.Set(reflect.Zero(field.Type()))
		return
	}

	value := reflect.ValueOf(webhookURLs)
	if value.Type().AssignableTo(field.Type()) {
		field.Set(value)
	}
}

func forwardedHeaderValue(value string) string {
	if value == "" {
		return ""
	}

	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}

func normalizeForwardedPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	return strings.TrimRight(prefix, "/")
}

func joinURLPath(base string, parts ...string) string {
	segments := make([]string, 0, len(parts)+1)
	if trimmedBase := strings.Trim(base, "/"); trimmedBase != "" {
		segments = append(segments, trimmedBase)
	}

	for _, part := range parts {
		if trimmed := strings.Trim(part, "/"); trimmed != "" {
			segments = append(segments, trimmed)
		}
	}

	if len(segments) == 0 {
		return "/"
	}

	return "/" + strings.Join(segments, "/")
}
