package handlers

import (
	"fmt"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/utils/rout"
)

// DisconnectIntegration removes the stored integration configuration and secrets for a provider.
func (h *Handler) DisconnectIntegration(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleDisconnectIntegrationRequest, models.DeleteIntegrationResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	userCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, ErrUnauthorized, openapi)
	}

	if h.IntegrationStore == nil {
		return h.InternalServerError(ctx, errIntegrationStoreNotConfigured, openapi)
	}

	provider := strings.TrimSpace(in.Provider)
	providerType, err := parseProviderType(provider)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	integrationID := strings.TrimSpace(in.IntegrationID)
	if integrationID == "" {
		record, err := h.DBClient.Integration.Query().
			Where(
				integration.OwnerID(caller.OrganizationID),
				integration.Kind(string(providerType)),
			).
			Only(userCtx)
		if err != nil {
			if ent.IsNotFound(err) {
				return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", providerType, ErrIntegrationNotFound)), openapi)
			}
			return h.InternalServerError(ctx, wrapIntegrationError("find", err), openapi)
		}
		integrationID = record.ID
	}

	deletedProvider, deletedID, err := h.IntegrationStore.DeleteIntegration(userCtx, caller.OrganizationID, integrationID)
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", providerType, ErrIntegrationNotFound)), openapi)
		default:
			return h.InternalServerError(ctx, wrapIntegrationError("delete", err), openapi)
		}
	}

	displayName := string(deletedProvider)
	out := models.DeleteIntegrationResponse{
		Reply:     rout.Reply{Success: true},
		Message:   fmt.Sprintf("%s integration disconnected", displayName),
		DeletedID: deletedID,
	}

	return h.Success(ctx, out)
}
