package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

// SwitchHandler is responsible for handling requests to the `/switch` endpoint, and changing the user's logged in organization context
func (h *Handler) SwitchHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleSwitchSuccessRequest, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	ac, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		log.Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err, openapi)
	}

	// ensure the user is not already in the target organization
	if ac.OrganizationID == in.TargetOrganizationID {
		return h.BadRequest(ctx, ErrAlreadySwitchedIntoOrg, openapi)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, ac.SubjectID)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err, openapi)
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

	return h.Success(ctx, out, openapi)
}
