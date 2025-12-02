package handlers

import (
	"fmt"
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/utils/passwd"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	models "github.com/theopenlane/core/pkg/openapi"
	sso "github.com/theopenlane/core/pkg/ssoutils"
	"github.com/theopenlane/ent/generated/privacy"
)

// LoginHandler validates the user credentials and returns a valid cookie
// this handler only supports username password login
func (h *Handler) LoginHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleLoginSuccessRequest, models.ExampleLoginSuccessResponse, openapi.Registry)
	if err != nil {
		metrics.RecordLogin(false)
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// check user in the database, username == email and ensure only one record is returned
	user, err := h.getUserByEmail(reqCtx, req.Username)
	if err != nil {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Err(err).Msg("unable to find user by email")

		return h.BadRequest(ctx, ErrLoginFailed, openapi)
	}

	if user.Edges.Setting.Status != enums.UserStatusActive {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("user not active")

		return h.BadRequest(ctx, ErrLoginFailed, openapi)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	orgStatus := h.orgEnforcementsForUser(allowCtx, req.Username)
	if orgStatus != nil && orgStatus.Enforced {
		return ctx.Redirect(http.StatusFound, sso.SSOLogin(ctx.Echo(), orgStatus.OrganizationID))
	}

	if user.Password == nil {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("no password set for user")

		return h.BadRequest(ctx, ErrLoginFailed, openapi)
	}

	// verify the password is correct
	valid, err := passwd.VerifyDerivedKey(*user.Password, req.Password)
	if err != nil || !valid {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("invalid password provided during login")

		return h.BadRequest(ctx, ErrLoginFailed, openapi)
	}

	if !user.Edges.Setting.EmailConfirmed {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("user email not verified, unable to login")

		return h.BadRequest(ctx, fmt.Errorf("%w: please check your email and verify your account", auth.ErrUnverifiedUser), openapi)
	}

	// set context for remaining request based on logged in user
	userCtx := setAuthenticatedContext(reqCtx, user)

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(userCtx, ctx.Response().Writer, user)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if err := h.updateUserLastSeen(userCtx, user.ID, enums.AuthProviderCredentials); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to update last seen")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// check if orgStatus is enforced, but user has not yet configured
	// if not yet configured we want to direct to the setup first
	tfaSetupRequired := false

	if orgStatus != nil && orgStatus.OrgTFAEnforced {
		// Check if user has TFA enabled
		if user.Edges.Setting == nil || !user.Edges.Setting.IsTfaEnabled {
			tfaSetupRequired = true
		}
	}

	out := models.LoginReply{
		Reply:            rout.Reply{Success: true},
		TFAEnabled:       user.Edges.Setting.IsTfaEnabled,
		TFASetupRequired: tfaSetupRequired,
		Message:          "success",
		AuthData:         *auth,
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, out, openapi)
}
