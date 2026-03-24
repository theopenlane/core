package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"
)

// ListIntegrationProviders returns declarative metadata about available third-party integration definitions
func (h *Handler) ListIntegrationProviders(ctx echo.Context, _ *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return nil
	}

	return h.Success(ctx, IntegrationProvidersResponse{
		Reply:     rout.Reply{Success: true},
		Providers: h.IntegrationsRuntime.Registry().Definitions(),
	})
}
