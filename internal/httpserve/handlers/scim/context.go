package scim

import (
	"context"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
)

// integrationContextKey is the context key for IntegrationContext
var integrationContextKey = contextx.NewKey[*IntegrationContext]()

// IntegrationContext carries SCIM source attribution and provision mode for a request
type IntegrationContext struct {
	// IntegrationID is the integration record ID from the URL path parameter
	IntegrationID string
	// OrgID is the organization that owns the integration
	OrgID string
	// ProvisionMode controls how SCIM pushes are persisted
	ProvisionMode enums.SCIMProvisionMode
}

// WithIntegrationContext stores the IntegrationContext in the given context
func WithIntegrationContext(ctx context.Context, ic *IntegrationContext) context.Context {
	return integrationContextKey.Set(ctx, ic)
}

// IntegrationContextFromContext retrieves the IntegrationContext from the given context
func IntegrationContextFromContext(ctx context.Context) (*IntegrationContext, bool) {
	return integrationContextKey.Get(ctx)
}

// ProvisionModeFromContext returns the SCIMProvisionMode from the context, defaulting to
// SCIMProvisionModeUsers when no IntegrationContext is present (e.g. on the legacy route)
func ProvisionModeFromContext(ctx context.Context) enums.SCIMProvisionMode {
	ic, ok := integrationContextKey.Get(ctx)
	if !ok || ic == nil {
		return enums.SCIMProvisionModeUsers
	}

	return ic.ProvisionMode
}
