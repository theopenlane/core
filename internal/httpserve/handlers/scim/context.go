package scim

import (
	"context"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

// integrationContextKey is the context key for IntegrationContext
var integrationContextKey = contextx.NewKey[*IntegrationContext]()

// IntegrationContext carries SCIM source attribution for a request
type IntegrationContext struct {
	// IntegrationID is the integration record ID from the URL path parameter
	IntegrationID string
	// Installation is the resolved integration installation for this request
	Installation *generated.Integration
	// OrgID is the organization that owns the integration
	OrgID string
	// Runtime executes shared integration ingest logic for this request
	Runtime *integrationsruntime.Runtime
}

// WithIntegrationContext stores the IntegrationContext in the given context
func WithIntegrationContext(ctx context.Context, ic *IntegrationContext) context.Context {
	return integrationContextKey.Set(ctx, ic)
}

// IntegrationContextFromContext retrieves the IntegrationContext from the given context
func IntegrationContextFromContext(ctx context.Context) (*IntegrationContext, bool) {
	return integrationContextKey.Get(ctx)
}
