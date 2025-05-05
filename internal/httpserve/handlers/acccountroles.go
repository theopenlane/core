package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/models"
)

// AccountAccessHandler list roles a subject has access to in relation an object
func (h *Handler) AccountRolesHandler(ctx echo.Context) error {
	var in models.AccountRolesRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting authenticated user")

		return h.InternalServerError(ctx, err)
	}

	req := fgax.ListAccess{
		SubjectType: in.SubjectType,
		SubjectID:   au.SubjectID,
		ObjectID:    in.ObjectID,
		ObjectType:  fgax.Kind(in.ObjectType),
		Relations:   in.Relations,
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	roles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error checking access")

		return h.BadRequest(ctx, ErrInvalidInput)
	}

	return h.Success(ctx, models.AccountRolesReply{
		Reply: rout.Reply{Success: true},
		Roles: roles,
	})
}

// BindAccountRoles returns the OpenAPI3 operation for accepting an account roles request
func (h *Handler) BindAccountRoles() *openapi3.Operation {
	roles := openapi3.NewOperation()
	roles.Description = "List roles a subject has in relation to an object"
	roles.Tags = []string{"account"}
	roles.OperationID = "AccountRoles"
	roles.Security = AllSecurityRequirements()

	h.AddRequestBody("AccountRolesRequest", models.ExampleAccountRolesRequest, roles)
	h.AddResponse("AccountRolesReply", "success", models.ExampleAccountRolesReply, roles, http.StatusOK)
	roles.AddResponse(http.StatusInternalServerError, internalServerError())
	roles.AddResponse(http.StatusBadRequest, badRequest())
	roles.AddResponse(http.StatusBadRequest, invalidInput())

	return roles
}
