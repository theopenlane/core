package handlers

import (
	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party integration definitions
func (h *Handler) ListIntegrationProviders(ctx echo.Context) error {
	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	defs := h.IntegrationsRuntime.Registry().Definitions()
	for i := range defs {
		defs[i].Operations = lo.Filter(defs[i].Operations, func(op types.OperationRegistration, _ int) bool {
			return op.CustomerSelectable == nil || *op.CustomerSelectable
		})
	}

	return h.Success(ctx, IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Providers: defs,
	})
}
