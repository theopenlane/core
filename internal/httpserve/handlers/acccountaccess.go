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

// AccountAccessHandler checks if a subject has access to an object
func (h *Handler) AccountAccessHandler(ctx echo.Context) error {
	var in models.AccountAccessRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
	}

	req := fgax.AccessCheck{
		SubjectType: in.SubjectType,
		Relation:    in.Relation,
		ObjectID:    in.ObjectID,
		ObjectType:  fgax.Kind(in.ObjectType),
	}

	subjectID, err := auth.GetUserIDFromContext(ctx.Request().Context())
	if err != nil {
		h.Logger.Errorw("error getting user id from context", "error", err)

		return h.InternalServerError(ctx, err)
	}

	req.SubjectID = subjectID

	allow, err := h.DBClient.Authz.CheckAccess(ctx.Request().Context(), req)
	if err != nil {
		h.Logger.Errorw("error checking access", "error", err)

		return h.InternalServerError(ctx, err)
	}

	return h.Success(ctx, models.AccountAccessReply{
		Reply:   rout.Reply{Success: true},
		Allowed: allow,
	})
}

// BindAccountAccess returns the OpenAPI3 operation for accepting an account access request
func (h *Handler) BindAccountAccess() *openapi3.Operation {
	checkAccess := openapi3.NewOperation()
	checkAccess.Description = "Check Subject Access to Object"
	checkAccess.OperationID = "AccountAccess"
	checkAccess.Security = &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"bearerAuth": []string{},
		},
	}

	h.AddRequestBody("AccountAccessRequest", models.ExampleAccountAccessRequest, checkAccess)
	h.AddResponse("AccountAccessReply", "success", models.ExampleAccountAccessReply, checkAccess, http.StatusOK)
	checkAccess.AddResponse(http.StatusInternalServerError, internalServerError())
	checkAccess.AddResponse(http.StatusBadRequest, badRequest())
	checkAccess.AddResponse(http.StatusUnauthorized, unauthorized())

	return checkAccess
}
