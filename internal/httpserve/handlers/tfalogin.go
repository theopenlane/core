package handlers

import (
	"encoding/json"
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

type tfaAuthenticationRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id,omitempty"`
}

// ValidateLoginTOTP completes a pending login by validating the second factor of the user tied to the pending cookie
func (h *Handler) ValidateLoginTOTP(ctx echo.Context) error {
	cookie, err := sessions.GetCookie(ctx.Request(), tfaPendingCookieName)
	if err != nil || cookie.Value == "" {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	data, err := h.SessionConfig.RedisStore.GetSession(ctx.Request().Context(), tfaPendingKeyPrefix+cookie.Value)
	if err != nil || data == "" {
		h.clearAuthFlowCookies(ctx.Response().Writer, tfaPendingCookieName)

		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	var pendingAuthRequest tfaAuthenticationRequest
	if err := json.Unmarshal([]byte(data), &pendingAuthRequest); err != nil || pendingAuthRequest.UserID == "" {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	return h.validateTOTP(ctx, pendingAuthRequest.UserID, cookie.Value)
}

// startTFAAuthentication parks the login behind a one time pending cookie until the second factor is validated
// the orgID if provided is preserved as that would be the identifier for the org the user is to be logged into. In
// cases where the sso starts with the org level sso login url, we need to bring the user back into that org.
// Without tracking, we cannot generate a sesion for the specific org
func (h *Handler) startTFAAuthentication(ctx echo.Context, userID, orgID string) error {
	state, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	reqCtx := ctx.Request().Context()

	data := tfaAuthenticationRequest{
		OrganizationID: orgID,
		UserID:         userID,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// the pending state is consumed on completion so the cookie can only ever mint one session
	if err := h.SessionConfig.RedisStore.StoreSessionWithExpiration(reqCtx, tfaPendingKeyPrefix+state, string(buf), tfaPendingTTL); err != nil {
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
func (h *Handler) completeTFALogin(ctx echo.Context, state, subjectID, orgID string) error {
	reqCtx := ctx.Request().Context()

	switch {
	case state != "":
		return h.issueTFALoginSession(ctx, subjectID, orgID)
	case auth.IsTFAEnrollmentRequiredFromContext(reqCtx):
		accessToken, err := auth.GetBearerToken(ctx)
		if err != nil {
			return h.Unauthorized(ctx, err)
		}

		// the enrollment-only access token is retired before the full session replaces it
		if err := h.revokeToken(reqCtx, accessToken); err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		return h.issueTFALoginSession(ctx, subjectID, orgID)
	default:
		return h.Success(ctx, models.TFAResponse{
			Success: true,
		})
	}
}

// issueTFALoginSession mints the full session once the second factor is proven
func (h *Handler) issueTFALoginSession(ctx echo.Context, userID, orgID string) error {
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

	authData, err := h.generateAuthSession(ctx, user, false, orgID)
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
