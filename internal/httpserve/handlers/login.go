package handlers

import (
	"fmt"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/utils/passwd"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/logx"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/metrics"
	sso "github.com/theopenlane/core/v2/pkg/ssoutils"
)

// LoginHandler validates the user credentials and returns a valid cookie
// this handler only supports username password login
func (h *Handler) LoginHandler(ctx echo.Context) error {
	req, err := BindAndValidate[models.LoginRequest](ctx)
	if err != nil {
		metrics.RecordLogin(false)
		logx.FromContext(ctx.Request().Context()).Warn().Err(err).Msg("invalid login request")

		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	// the Openlane support identity is virtual and authenticated entirely from configuration; detect it
	// before any database lookup so this email can never authenticate against a database user, and route
	// it to the configuration backed support login (password factor plus second factor identity provider)
	if h.SupportAccessConfig.Enabled && strings.EqualFold(req.Username, h.SupportAccessConfig.Email) {
		return h.supportFirstFactor(ctx, req)
	}

	// check user in the database, username == email and ensure only one record is returned
	user, err := h.getUserByEmail(reqCtx, req.Username)
	if err != nil {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Err(err).Msg("unable to find user by email")

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	if user.Edges.Setting.Status != enums.UserStatusActive {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("user not active")

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	internalCtx := auth.WithInternalReadContext(reqCtx)

	orgStatus := h.orgEnforcementsForUser(internalCtx, req.Username)
	if orgStatus != nil && orgStatus.Enforced {
		return ctx.Redirect(http.StatusFound, sso.SSOLogin(ctx.Echo(), orgStatus.OrganizationID))
	}

	orgTFAEnforced := orgStatus != nil && orgStatus.OrgTFAEnforced

	if user.Password == nil {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("no password set for user")

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	// verify the password is correct
	valid, err := passwd.VerifyDerivedKey(*user.Password, req.Password)
	if err != nil || !valid {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("invalid password provided during login")

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	if !user.Edges.Setting.EmailConfirmed {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("email", req.Username).Msg("user email not verified, unable to login")

		return h.BadRequest(ctx, fmt.Errorf("%w: please check your email and verify your account", auth.ErrUnverifiedUser))
	}

	switch {
	case user.Edges.Setting.IsTfaEnabled:
		return h.startTFAAuthentication(ctx, user.ID, "")
	case orgTFAEnforced:
		return h.startTFAEnrollment(ctx, user)
	}

	authData, err := h.generateAuthSession(ctx, user, false, "")
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, models.LoginResponse{
		Success:  true,
		Message:  "success",
		AuthData: *authData,
	})
}

// startTFAEnrollment issues an enrollment-only session when the organization enforces two factor authentication
// and the user has not enrolled yet
func (h *Handler) startTFAEnrollment(ctx echo.Context, user *ent.User) error {
	authData, err := h.generateAuthSession(ctx, user, true, "")
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, models.LoginResponse{
		Success:          true,
		TFASetupRequired: true,
		Message:          "2fa enrollment required",
		AuthData:         *authData,
	})
}

// generateAuthSession mints the session for the user. if the orgID is provided, the session is scoped there else it
// uses the default organization of the user. enrollment-only when tfaEnrollmentRequired is set
func (h *Handler) generateAuthSession(ctx echo.Context, user *ent.User, tfaEnrollmentRequired bool, orgID string) (*models.AuthData, error) {
	reqCtx := ctx.Request().Context()
	userCtx := setAuthenticatedContext(reqCtx, user)

	authData, err := h.AuthManager.GenerateUserAuthSessionWithOrg(userCtx, ctx.Response().Writer, user, orgID, tfaEnrollmentRequired)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return nil, err
	}

	if err := h.updateUserLastSeen(userCtx, user.ID, enums.AuthProviderCredentials); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to update last seen")

		return nil, err
	}

	return authData, nil
}
