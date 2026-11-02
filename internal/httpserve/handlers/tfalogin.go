package handlers

import (
	"time"

	models "github.com/theopenlane/core/common/openapi"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/v2/pkg/metrics"
)

const (
	tfaPendingCookieName = "tfa_pending"
	tfaPendingKeyPrefix  = "tfa:pending:"
)

func (h *Handler) ValidateLoginTOTP(ctx echo.Context) error {
	cookie, err := sessions.GetCookie(ctx.Request(), tfaPendingCookieName)
	if err != nil || cookie.Value == "" {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	subjectID, err := h.RedisClient.Get(ctx.Request().Context(), tfaPendingKeyPrefix+cookie.Value).Result()
	if err != nil || subjectID == "" {
		h.clearAuthFlowCookies(ctx.Response().Writer, tfaPendingCookieName)

		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	return h.validateTOTP(ctx, subjectID, cookie.Value)
}

func (h *Handler) startTFAAuthentication(ctx echo.Context, userID string) error {

	var pendingTTL = 3 * time.Minute

	state, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	reqCtx := ctx.Request().Context()

	// we want to prevent replay of the request with the cookie to allow minting tokens multiple times and one time usage only.
	// encoding a valid_until in the cookie won't fix that
	if err := h.RedisClient.Set(ctx.Request().Context(), tfaPendingKeyPrefix+state, userID, pendingTTL).Err(); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("could not store tfa auth requirement in redis")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	cfg := *h.SessionConfig.CookieConfig
	cfg.MaxAge = int(pendingTTL.Seconds())

	sessions.SetCookie(ctx.Response().Writer, state, tfaPendingCookieName, cfg)

	return h.Success(ctx, models.LoginResponse{
		Reply: rout.Reply{
			Success: true,
		},
		TFAEnabled: true,
		Message:    "2fa validation required",
	})
}

func (h *Handler) completeTFALogin(ctx echo.Context, state, subjectID string) error {
	if state == "" {
		return h.Success(ctx, models.TFAResponse{
			Reply: rout.Reply{
				Success: true,
			},
		})
	}

	reqCtx := ctx.Request().Context()

	userID, err := h.RedisClient.GetDel(reqCtx, tfaPendingKeyPrefix+state).Result()
	if err != nil || userID != subjectID {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	h.clearAuthFlowCookies(ctx.Response().Writer, tfaPendingCookieName)

	user, err := h.getUserDetailsByID(auth.WithInternalReadContext(reqCtx), userID)
	if err != nil {
		return h.BadRequest(ctx, ErrLoginFailed)
	}

	authData, err := h.generateAuthSession(ctx, user)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, models.LoginResponse{
		Reply: rout.Reply{
			Success: true,
		},
		AuthData: *authData,
		Message:  "success",
	})
}
