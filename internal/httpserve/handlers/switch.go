package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
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

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("unable to get user id from context")

		return h.BadRequest(ctx, auth.ErrNoAuthUser, openapi)
	}

	// ensure the user is not already in the target organization
	if caller.OrganizationID == in.TargetOrganizationID {
		return h.BadRequest(ctx, ErrAlreadySwitchedIntoOrg, openapi)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, caller.SubjectID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err, openapi)
	}

	// check if SSO is enforced for the target organization, then apply owner, per-user, and per-domain
	// exemptions to decide whether this user must be redirected through the SSO login flow.
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	status, err := h.fetchSSOStatus(allowCtx, in.TargetOrganizationID, user.ID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to resolve sso enforcement for organization switch")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// fetchSSOStatus already applied this user's exemption, so status.Enforced reflects whether they
	// must be redirected through SSO for the target organization
	if status.Enforced {
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

	// check if TFA is enforced for the target organization and user doesn't have TFA enabled
	if status.OrgTFAEnforced {
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
