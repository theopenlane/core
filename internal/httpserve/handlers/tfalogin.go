package handlers

import (
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/pkg/metrics"
)

const (
	tfaPendingCookieName = "tfa_pending"
	tfaPendingKeyPrefix  = "tfa:pending:"
	// tfaPendingTTL is how long a login may wait for its second factor before it must start over
	tfaPendingTTL = 5 * time.Minute
)

// ValidateLoginTOTP completes a pending login by validating the second factor of the user tied to the pending cookie
func (h *Handler) ValidateLoginTOTP(ctx echo.Context) error {
	cookie, err := sessions.GetCookie(ctx.Request(), tfaPendingCookieName)
	if err != nil || cookie.Value == "" {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	subjectID, err := h.SessionConfig.RedisStore.GetSession(ctx.Request().Context(), tfaPendingKeyPrefix+cookie.Value)
	if err != nil || subjectID == "" {
		h.clearAuthFlowCookies(ctx.Response().Writer, tfaPendingCookieName)

		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	return h.validateTOTP(ctx, subjectID, cookie.Value)
}

// startTFAAuthentication parks the login behind a one time pending cookie until the second factor is validated
func (h *Handler) startTFAAuthentication(ctx echo.Context, userID string) error {
	state, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	reqCtx := ctx.Request().Context()

	// the pending state is consumed on completion so the cookie can only ever mint one session
	if err := h.SessionConfig.RedisStore.StoreSessionWithExpiration(reqCtx, tfaPendingKeyPrefix+state, userID, tfaPendingTTL); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("could not store tfa auth requirement in redis")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	cfg := *h.SessionConfig.CookieConfig
	cfg.MaxAge = int(tfaPendingTTL.Seconds())

	sessions.SetCookie(ctx.Response().Writer, state, tfaPendingCookieName, cfg)

	return h.Success(ctx, models.LoginResponse{
		Success:    true,
		TFAEnabled: true,
		Message:    "2fa validation required",
	})
}

// completeTFALogin responds to a proven second factor, issuing a full session
func (h *Handler) completeTFALogin(ctx echo.Context, state, subjectID string) error {
	reqCtx := ctx.Request().Context()

	switch {
	case state != "":
		return h.issueTFALoginSession(ctx, subjectID)
	case auth.IsTFAEnrollmentRequiredFromContext(reqCtx):
		accessToken, err := auth.GetBearerToken(ctx)
		if err != nil {
			return h.Unauthorized(ctx, err)
		}

		// the enrollment-only access token is retired before the full session replaces it
		if err := h.revokeToken(reqCtx, accessToken); err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		return h.issueTFALoginSession(ctx, subjectID)
	default:
		return h.Success(ctx, models.TFAResponse{
			Success: true,
		})
	}
}

// issueTFALoginSession mints the full session once the second factor is proven
func (h *Handler) issueTFALoginSession(ctx echo.Context, userID string) error {
	reqCtx := ctx.Request().Context()

	user, err := h.getUserDetailsByID(auth.WithInternalReadContext(reqCtx), userID)
	if err != nil {
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	if user.Edges.Setting.Status != enums.UserStatusActive || !user.Edges.Setting.EmailConfirmed {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("user_id", userID).Msg("user not active or email not verified, unable to complete login")

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	authData, err := h.generateAuthSession(ctx, user, false)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, models.LoginResponse{
		Success:    true,
		TFAEnabled: user.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *authData,
	})
}
