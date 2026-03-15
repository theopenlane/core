package handlers

import (
	"fmt"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
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

	if strings.TrimSpace(in.Provider) == "" {
		return h.BadRequest(ctx, rout.MissingField("provider"), openapi)
	}

	def, ok := h.resolveIntegrationDefinition(in.Provider)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapi)
	}

	integrationID := in.IntegrationID
	if integrationID == "" {
		record, err := h.loadOwnedIntegrationByDefinition(userCtx, caller.OrganizationID, def.Spec.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", def.Spec.Slug, ErrIntegrationNotFound)), openapi)
			}
			return h.InternalServerError(ctx, wrapIntegrationError("find", err), openapi)
		}
		integrationID = record.ID
	} else {
		record, err := h.loadOwnedIntegration(userCtx, caller.OrganizationID, integrationID)
		if err != nil {
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapi)
		}
		if err := validateInstallationDefinition(record, def); err != nil {
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
