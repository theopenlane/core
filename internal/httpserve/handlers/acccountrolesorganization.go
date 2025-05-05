package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
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
func (h *Handler) AccountRolesOrganizationHandler(ctx echo.Context) error {
	var in models.AccountRolesOrganizationRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting authenticated user")

		return h.InternalServerError(ctx, err)
	}

	in.ID, err = h.getOrganizationID(in.ID, au)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// validate the input
	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
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
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error checking access")

		return h.InternalServerError(ctx, err)
	}

	return h.Success(ctx, models.AccountRolesOrganizationReply{
		Reply:          rout.Reply{Success: true},
		Roles:          roles,
		OrganizationID: req.ObjectID,
	})
}

// BindAccountRolesOrganization returns the OpenAPI3 operation for accepting an account roles organization request
func (h *Handler) BindAccountRolesOrganization() *openapi3.Operation {
	orgRoles := openapi3.NewOperation()
	orgRoles.Description = "List roles a subject has in relation to the authenticated organization"
	orgRoles.Tags = []string{"account"}
	orgRoles.OperationID = "AccountRolesOrganization"
	orgRoles.Security = AllSecurityRequirements()

	orgRoles.AddResponse(http.StatusInternalServerError, internalServerError())
	orgRoles.AddResponse(http.StatusBadRequest, badRequest())
	orgRoles.AddResponse(http.StatusBadRequest, invalidInput())

	return orgRoles
}

// BindAccountRolesOrganization returns the OpenAPI3 operation for accepting an account roles organization request
func (h *Handler) BindAccountRolesOrganizationByID() *openapi3.Operation {
	orgRoles := openapi3.NewOperation()
	orgRoles.Description = "List roles a subject has in relation to the organization ID provided"
	orgRoles.Tags = []string{"account"}
	orgRoles.OperationID = "AccountRolesOrganizationByID"
	orgRoles.Security = AllSecurityRequirements()

	h.AddResponse("AccountRolesOrganizationReply", "success", models.ExampleAccountRolesOrganizationReply, orgRoles, http.StatusOK)
	orgRoles.AddResponse(http.StatusInternalServerError, internalServerError())
	orgRoles.AddResponse(http.StatusBadRequest, badRequest())
	orgRoles.AddResponse(http.StatusUnauthorized, unauthorized())

	return orgRoles
}
