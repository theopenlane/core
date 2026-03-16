package handlers

import (
	"errors"
	"fmt"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/utils/rout"
)

// DisconnectIntegration removes the stored integration configuration and secrets for a provider.
func (h *Handler) DisconnectIntegration(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleDisconnectIntegrationRequest, models.DeleteIntegrationResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapi); err != nil {
		return err
	}

	userCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, ErrUnauthorized, openapi)
	}

	if in.Provider == "" {
		return h.BadRequest(ctx, rout.MissingField("provider"), openapi)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(in.Provider)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapi)
	}

	record, err := h.IntegrationsRuntime.ResolveInstallation(userCtx, caller.OrganizationID, in.IntegrationID, def.ID)
	if err != nil {
		switch {
		case errors.Is(err, integrationsruntime.ErrInstallationIDRequired):
			return h.BadRequest(ctx, ErrIntegrationIDRequired, openapi)
		case errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch):
			return h.BadRequest(ctx, ErrInvalidProvider, openapi)
		case errors.Is(err, integrationsruntime.ErrInstallationNotFound):
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", def.Slug, ErrIntegrationNotFound)), openapi)
		default:
			return h.InternalServerError(ctx, wrapIntegrationError("find", err), openapi)
		}
	}

	integrationID := record.ID

	if err := h.IntegrationsRuntime.DeleteCredential(userCtx, integrationID); err != nil {
		return h.InternalServerError(ctx, wrapIntegrationError("delete credentials for", err), openapi)
	}

	if err := h.DBClient.Integration.DeleteOneID(integrationID).Exec(userCtx); err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", def.Slug, ErrIntegrationNotFound)), openapi)
		}
		return h.InternalServerError(ctx, wrapIntegrationError("delete", err), openapi)
	}

	displayName := def.DisplayName
	if displayName == "" {
		displayName = def.Slug
	}

	out := models.DeleteIntegrationResponse{
		Reply:     rout.Reply{Success: true},
		Message:   fmt.Sprintf("%s integration disconnected", displayName),
		DeletedID: integrationID,
	}

	return h.Success(ctx, out)
}
