package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// ConfigureIntegrationProvider stores non-OAuth credentials for a provider definition.
// When installation_id is provided the credentials on that installation are updated.
// When omitted a new installation is created and its ID is returned in the response
func (h *Handler) ConfigureIntegrationProvider(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	payload, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationConfigPayload, IntegrationConfigResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(payload.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	installationRec, _, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, payload.InstallationID, def)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("payload", payload).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	var credential *types.CredentialSet

	if !jsonx.IsEmptyRawMessage(payload.Body) {
		credential = &types.CredentialSet{Data: jsonx.CloneRawMessage(payload.Body)}
	}

	if err := h.IntegrationsRuntime.Reconcile(requestCtx, installationRec, payload.UserInput, payload.CredentialRef, credential, nil); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("payload", payload).Msg("reconcile failed")
		return h.BadRequest(ctx, ErrProcessingRequest, openapiCtx)
	}

	return h.Success(ctx, IntegrationConfigResponse{
		Reply:          rout.Reply{Success: true},
		Provider:       def.Slug,
		InstallationID: installationRec.ID,
	})
}
