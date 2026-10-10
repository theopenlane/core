package handlers

import (
	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party integration definitions
func (h *Handler) ListIntegrationProviders(ctx echo.Context) error {
	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	defs := lo.Filter(h.IntegrationsRuntime.Registry().Definitions(), func(def types.Definition, _ int) bool {
		return def.Visible
	})

	return h.Success(ctx, IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Providers: lo.Map(defs, func(def types.Definition, _ int) IntegrationProvider { return projectIntegrationProvider(def) }),
	})
}

// projectIntegrationProvider builds the provider listing projection of one definition
func projectIntegrationProvider(def types.Definition) IntegrationProvider {
	connections := def.ConnectionList()

	return IntegrationProvider{
		Spec:           def.DefinitionSpec,
		OperatorConfig: def.OperatorConfig,
		UserInput:      def.UserInput,
		CredentialRegistrations: lo.Map(connections, func(c types.Connection, _ int) IntegrationProviderCredential {
			return IntegrationProviderCredential{
				Ref:         c.Credential.Name,
				Name:        c.Name,
				Description: c.Description,
				Schema:      c.Form,
				Recommended: c.Recommended,
			}
		}),
		Connections: lo.Map(connections, func(c types.Connection, _ int) IntegrationProviderConnection {
			return IntegrationProviderConnection{
				CredentialRef:  c.Credential.Name,
				Name:           c.Name,
				Description:    c.Description,
				Meta:           c.Meta,
				CredentialRefs: []string{c.Credential.Name},
				Auth:           lo.Ternary(c.Auth != nil, &IntegrationProviderAuth{}, nil),
				Disconnect:     c.Disconnect,
			}
		}),
		Operations: lo.Filter(def.Operations, func(op types.OperationRegistration, _ int) bool {
			return op.CustomerSelectable == nil || *op.CustomerSelectable
		}),
		Webhooks: def.Webhooks,
	}
}
