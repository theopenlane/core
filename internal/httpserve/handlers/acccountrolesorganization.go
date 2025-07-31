package handlers

import (
	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/models"
)

// AccountRolesOrganizationHandler lists roles a subject has in relation to an organization
func (h *Handler) AccountRolesOrganizationHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParams(ctx, openapi.Operation, models.ExampleAccountRolesOrganizationRequest, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting authenticated user")

		return h.InternalServerError(ctx, err, openapi)
	}

	in.ID, err = h.getOrganizationID(in.ID, au)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// validate the input
	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	req := fgax.ListAccess{
		SubjectType: auth.GetAuthzSubjectType(reqCtx),
		SubjectID:   au.SubjectID,
		ObjectID:    in.ID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	roles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")

		return h.InternalServerError(ctx, err, openapi)
	}

	return h.Success(ctx, models.AccountRolesOrganizationReply{
		Reply:          rout.Reply{Success: true},
		Roles:          roles,
		OrganizationID: req.ObjectID,
	}, openapi)
}
