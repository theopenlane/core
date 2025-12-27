package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// SwitchHandler is responsible for handling requests to the `/switch` endpoint, and changing the user's logged in organization context
func (h *Handler) SwitchHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleSwitchSuccessRequest, &models.SwitchOrganizationReply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	ac, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err, openapi)
	}

	// ensure the user is not already in the target organization
	if ac.OrganizationID == in.TargetOrganizationID {
		return h.BadRequest(ctx, ErrAlreadySwitchedIntoOrg, openapi)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, ac.SubjectID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err, openapi)
	}

	// check if SSO is enforced for the target organization. If so, redirect
	// the user through the SSO login flow unless they are an owner.
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	status, err := h.fetchSSOStatus(allowCtx, in.TargetOrganizationID, user.ID)

	if err == nil && status.Enforced {
		member, mErr := transaction.FromContext(allowCtx).OrgMembership.Query().Where(
			orgmembership.UserID(user.ID),
			orgmembership.OrganizationID(in.TargetOrganizationID),
		).Only(allowCtx)
		if mErr == nil && member.Role != enums.RoleOwner {
			authURL, err := h.generateSSOAuthURL(ctx, in.TargetOrganizationID)
			if err != nil {
				logx.FromContext(reqCtx).Error().Err(err).Msg("unable to generate SSO auth URL")
				return h.BadRequest(ctx, err, openapi)
			}

			sessions.SetCookie(ctx.Response().Writer, authenticatedUserSSOCookieValue, authenticatedUserSSOCookieName, *h.SessionConfig.CookieConfig)

			out := &models.SwitchOrganizationReply{
				Reply:       rout.Reply{Success: true},
				NeedsSSO:    true,
				RedirectURI: authURL,
			}

			return h.Success(ctx, out, openapi)
		}
	}

	// check if TFA is enforced for the target organization and user doesn't have TFA enabled
	if err == nil && status.OrgTFAEnforced {
		// Check if user has TFA enabled
		if user.Edges.Setting == nil || !user.Edges.Setting.IsTfaEnabled {
			out := &models.SwitchOrganizationReply{
				Reply:    rout.Reply{Success: true},
				NeedsTFA: true,
			}

			return h.Success(ctx, out, openapi)
		}
	}

	// create new claims for the user
	authData, err := h.AuthManager.GenerateUserAuthSessionWithOrg(reqCtx, ctx.Response().Writer, user, in.TargetOrganizationID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.Unauthorized(ctx, err)
	}

	// set the out attributes we send back to the client only on success
	out := &models.SwitchOrganizationReply{
		Reply:    rout.Reply{Success: true},
		AuthData: *authData,
	}

	return h.Success(ctx, out, openapi)
}
