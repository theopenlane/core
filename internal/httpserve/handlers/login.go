package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/utils/passwd"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/metrics"
	models "github.com/theopenlane/core/pkg/openapi"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

// LoginHandler validates the user credentials and returns a valid cookie
// this handler only supports username password login
func (h *Handler) LoginHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleLoginSuccessRequest, models.ExampleLoginSuccessResponse, openapi.Registry)
	if err != nil {
		metrics.RecordLogin(false)
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	// check user in the database, username == email and ensure only one record is returned
	user, err := h.getUserByEmail(reqCtx, req.Username)
	if err != nil {
		metrics.RecordLogin(false)
		return h.BadRequest(ctx, auth.ErrNoAuthUser, openapi)
	}

	if user.Edges.Setting.Status != enums.UserStatusActive {
		metrics.RecordLogin(false)
		return h.BadRequest(ctx, auth.ErrNoAuthUser, openapi)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	if orgID, ok := h.ssoOrgForUser(allowCtx, req.Username); ok {
		metrics.RecordLogin(false)
		return ctx.Redirect(http.StatusFound, sso.SSOLogin(ctx.Echo(), orgID))
	}

	if user.Password == nil {
		metrics.RecordLogin(false)
		return h.BadRequest(ctx, rout.ErrInvalidCredentials, openapi)
	}

	// verify the password is correct
	valid, err := passwd.VerifyDerivedKey(*user.Password, req.Password)
	if err != nil || !valid {
		metrics.RecordLogin(false)
		return h.BadRequest(ctx, rout.ErrInvalidCredentials, openapi)
	}

	if !user.Edges.Setting.EmailConfirmed {
		metrics.RecordLogin(false)
		return h.BadRequest(ctx, auth.ErrUnverifiedUser, openapi)
	}

	// set context for remaining request based on logged in user
	userCtx := setAuthenticatedContext(reqCtx, user)

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(userCtx, ctx.Response().Writer, user)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err, openapi)
	}

	if err := h.updateUserLastSeen(userCtx, user.ID, enums.AuthProviderCredentials); err != nil {
		log.Error().Err(err).Msg("unable to update last seen")

		return h.InternalServerError(ctx, err, openapi)
	}

	out := models.LoginReply{
		Reply:      rout.Reply{Success: true},
		TFAEnabled: user.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *auth,
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, out, openapi)
}
