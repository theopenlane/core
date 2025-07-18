package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
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

	reqCtx := ctx.Request().Context()

	subject, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting user id from context")

		return h.InternalServerError(ctx, err)
	}

	req := fgax.AccessCheck{
		SubjectType: in.SubjectType,
		Relation:    in.Relation,
		ObjectID:    in.ObjectID,
		ObjectType:  fgax.Kind(in.ObjectType),
		SubjectID:   subject.SubjectID,
		Context:     utils.NewOrganizationContextKey(subject.SubjectEmail),
	}

	allow, err := h.DBClient.Authz.CheckAccess(reqCtx, req)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")

		return h.BadRequest(ctx, ErrInvalidInput)
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
	checkAccess.Tags = []string{"account"}
	checkAccess.OperationID = "AccountAccess"
	checkAccess.Security = AllSecurityRequirements()

	h.AddRequestBody("AccountAccessRequest", models.ExampleAccountAccessRequest, checkAccess)
	h.AddResponse("AccountAccessReply", "success", models.ExampleAccountAccessReply, checkAccess, http.StatusOK)
	checkAccess.AddResponse(http.StatusInternalServerError, internalServerError())
	checkAccess.AddResponse(http.StatusBadRequest, badRequest())
	checkAccess.AddResponse(http.StatusUnauthorized, unauthorized())
	checkAccess.AddResponse(http.StatusBadRequest, invalidInput())

	return checkAccess
}
