package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// AccountRolesOrganizationHandler lists roles a subject has in relation to an organization
func (h *Handler) AccountRolesOrganizationHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, models.ExampleAccountRolesOrganizationRequest, models.ExampleAccountRolesOrganizationReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting authenticated user")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	orgID, orgIDErr := h.getOrganizationID(in.ID, caller)
	if orgIDErr != nil {
		return h.BadRequest(ctx, orgIDErr, openapi)
	}

	in.ID = orgID

	// validate the input
	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	req := fgax.ListAccess{
		SubjectType: caller.SubjectType(),
		SubjectID:   caller.SubjectID,
		ObjectID:    in.ID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
		Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
	}

	roles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, models.AccountRolesOrganizationReply{
		Reply:          rout.Reply{Success: true},
		Roles:          roles,
		OrganizationID: req.ObjectID,
	}, openapi)
}
