package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/auth"
	"github.com/theopenlane/core/pkg/models"
)

// DefaultAllRelations is the default list of relations to check
// these come from the fga/model/model.fga file relations
// TODO (sfunk): look into a way to get this from the fga model
var DefaultAllRelations = []string{
	"can_view",
	"can_edit",
	"can_delete",
	"audit_log_viewer",
	"can_invite_admins",
	"can_invite_members",
}

// AccountAccessHandler list roles a subject has access to in relation an object
func (h *Handler) AccountRolesHandler(ctx echo.Context) error {
	var in models.AccountRolesRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
	}

	req := fgax.ListAccess{
		SubjectType: in.SubjectType,
		ObjectID:    in.ObjectID,
		ObjectType:  fgax.Kind(in.ObjectType),
		Relations:   in.Relations,
	}

	// if no relations are provided, default to all relations
	if len(req.Relations) == 0 {
		req.Relations = DefaultAllRelations
	}

	subjectID, err := auth.GetUserIDFromContext(ctx.Request().Context())
	if err != nil {
		h.Logger.Errorw("error getting user id from context", "error", err)

		return h.InternalServerError(ctx, err)
	}

	req.SubjectID = subjectID

	roles, err := h.DBClient.Authz.ListRelations(ctx.Request().Context(), req)
	if err != nil {
		h.Logger.Error("error checking access", "error", err)

		return h.InternalServerError(ctx, err)
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
	roles.OperationID = "AccountRoles"
	roles.Security = &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"bearerAuth": []string{},
		},
	}

	h.AddRequestBody("AccountRolesRequest", models.ExampleAccountRolesRequest, roles)
	h.AddResponse("AccountRolesReply", "success", models.ExampleAccountRolesReply, roles, http.StatusOK)
	roles.AddResponse(http.StatusInternalServerError, internalServerError())
	roles.AddResponse(http.StatusBadRequest, badRequest())
	roles.AddResponse(http.StatusUnauthorized, unauthorized())

	return roles
}
