package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/logx"
)

// AccountRolesOrganizationHandler lists roles a subject has in relation to an organization
func (h *Handler) AccountRolesOrganizationHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.AccountRolesOrganizationRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting authenticated user")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	orgID, orgIDErr := h.getOrganizationID(in.ID, caller)
	if orgIDErr != nil {
		return h.BadRequest(ctx, orgIDErr)
	}

	in.ID = orgID

	// validate the input
	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
	}

	req := fgax.ListAccess{
		SubjectType: caller.SubjectType(),
		SubjectID:   caller.SubjectID,
		ObjectID:    in.ID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
	}

	roles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, models.AccountRolesOrganizationResponse{
		Reply:          rout.Reply{Success: true},
		Roles:          roles,
		OrganizationID: req.ObjectID,
	})
}
