package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/sso"
)

// SwitchHandler is responsible for handling requests to the `/switch` endpoint, and changing the user's logged in organization context
func (h *Handler) SwitchHandler(ctx echo.Context) error {
	var in models.SwitchOrganizationRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	ac, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		log.Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err)
	}

	// ensure the user is not already in the target organization
	if ac.OrganizationID == in.TargetOrganizationID {
		return h.BadRequest(ctx, ErrAlreadySwitchedIntoOrg)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, ac.SubjectID)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err)
	}

	// Check if SSO is enforced for the target organization. If so, redirect
	// the user through the SSO login flow unless they are an owner.
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	status, err := h.fetchSSOStatus(allowCtx, in.TargetOrganizationID)

	if err == nil && status.Enforced {
		member, mErr := transaction.FromContext(allowCtx).OrgMembership.Query().Where(
			orgmembership.UserID(user.ID),
			orgmembership.OrganizationID(in.TargetOrganizationID),
		).Only(allowCtx)
		if mErr == nil && member.Role != enums.RoleOwner {
			loginURL := sso.SSOLogin(ctx.Echo(), in.TargetOrganizationID)
			return ctx.Redirect(http.StatusFound, loginURL)
		}
	}

	// create new claims for the user
	authData, err := h.AuthManager.GenerateUserAuthSessionWithOrg(reqCtx, ctx.Response().Writer, user, in.TargetOrganizationID)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

		return h.Unauthorized(ctx, err)
	}

	// set the out attributes we send back to the client only on success
	out := &models.SwitchOrganizationReply{
		Reply:    rout.Reply{Success: true},
		AuthData: *authData,
	}

	return h.Success(ctx, out)
}

// BindSwitchHandler binds the reset password handler to the OpenAPI schema
func (h *Handler) BindSwitchHandler() *openapi3.Operation {
	switchHandler := openapi3.NewOperation()
	switchHandler.Description = "Switch the user's organization context"
	switchHandler.Tags = []string{"switchorganizations"}
	switchHandler.OperationID = "OrganizationSwitch"
	switchHandler.Security = AllSecurityRequirements()

	h.AddRequestBody("SwitchOrganizationRequest", models.ExampleSwitchSuccessRequest, switchHandler)
	h.AddResponse("SwitchOrganizationReply", "success", models.ExampleSwitchSuccessReply, switchHandler, http.StatusOK)
	switchHandler.AddResponse(http.StatusInternalServerError, internalServerError())
	switchHandler.AddResponse(http.StatusBadRequest, badRequest())
	switchHandler.AddResponse(http.StatusUnauthorized, unauthorized())
	switchHandler.AddResponse(http.StatusBadRequest, invalidInput())

	return switchHandler
}
