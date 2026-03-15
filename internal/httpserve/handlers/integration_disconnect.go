package handlers

import (
	"fmt"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
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

	if in.Provider == "" {
		return h.BadRequest(ctx, rout.MissingField("provider"), openapi)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(in.Provider))
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapi)
	}

	integrationID := in.IntegrationID
	if integrationID == "" {
		record, err := h.DBClient.Integration.Query().
			Where(
				integration.OwnerIDEQ(caller.OrganizationID),
				integration.DefinitionIDEQ(string(def.Spec.ID)),
			).
			Only(userCtx)
		if err != nil {
			if ent.IsNotSingular(err) {
				return h.BadRequest(ctx, ErrIntegrationIDRequired, openapi)
			}
			if ent.IsNotFound(err) {
				return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", def.Spec.Slug, ErrIntegrationNotFound)), openapi)
			}
			return h.InternalServerError(ctx, wrapIntegrationError("find", err), openapi)
		}
		integrationID = record.ID
	} else {
		record, err := h.DBClient.Integration.Query().
			Where(
				integration.OwnerIDEQ(caller.OrganizationID),
				integration.IDEQ(integrationID),
			).
			Only(userCtx)
		if err != nil {
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapi)
		}
		if record.DefinitionID != string(def.Spec.ID) {
			return h.BadRequest(ctx, ErrInvalidProvider, openapi)
		}
	}

	if err := h.IntegrationsRuntime.CredentialStore().DeleteCredential(userCtx, integrationID); err != nil {
		return h.InternalServerError(ctx, wrapIntegrationError("delete credentials for", err), openapi)
	}

	if err := h.DBClient.Integration.DeleteOneID(integrationID).Exec(userCtx); err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", def.Spec.Slug, ErrIntegrationNotFound)), openapi)
		}
		return h.InternalServerError(ctx, wrapIntegrationError("delete", err), openapi)
	}

	displayName := def.Spec.DisplayName
	if displayName == "" {
		displayName = def.Spec.Slug
	}

	out := models.DeleteIntegrationResponse{
		Reply:     rout.Reply{Success: true},
		Message:   fmt.Sprintf("%s integration disconnected", displayName),
		DeletedID: integrationID,
	}

	return h.Success(ctx, out)
}
