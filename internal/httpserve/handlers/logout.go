package handlers

import (
	"context"
	"errors"
	"time"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/logx"
)

// LogoutHandler revokes the caller's access and refresh tokens and deletes their server-side
// session so that logout takes effect on the server rather than only clearing client state. The
// endpoint is public so that a caller holding an expired access token can still log out. Cookies
// are only cleared once the server-side revocation has succeeded so that a failed logout is
// retried by the client rather than silently leaving valid credentials in place
func (h *Handler) LogoutHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleLogoutRequest, models.ExampleLogoutSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// revoke the access token when one is presented so it is rejected before its natural expiry. Any
	// failure to read it, whether absent or malformed, means there is nothing to revoke here and must
	// not block logout, so the specific error is irrelevant
	if accessToken, err := auth.GetBearerToken(ctx); err == nil {
		if err := h.revokeToken(reqCtx, accessToken); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("unable to revoke access token on logout")

			return h.InternalServerError(ctx, err, openapi)
		}
	}

	// resolve the refresh token, preferring the bound request body which is the declared contract
	// for this endpoint and falling back to the cookie for browser clients that set it. The only
	// error GetRefreshToken returns is ErrNoRefreshToken, meaning none was presented, which leaves
	// refreshToken empty and skips revocation below
	refreshToken := req.RefreshToken
	if refreshToken == "" {
		if cookieToken, err := auth.GetRefreshToken(ctx); err == nil {
			refreshToken = cookieToken
		}
	}

	if refreshToken != "" {
		if err := h.revokeToken(reqCtx, refreshToken); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("unable to revoke refresh token on logout")

			return h.InternalServerError(ctx, err, openapi)
		}
	}

	// destroy the server-side session and expire its cookie so the session middleware rejects
	// subsequent requests; this is the inverse of the login flow's CreateAndStoreSession
	if err := h.SessionConfig.DestroySession(reqCtx, ctx.Response().Writer, ctx.Request()); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to destroy session on logout")

		return h.InternalServerError(ctx, err, openapi)
	}

	// clear the auth token cookies using the same config they were set with, once the server-side
	// revocation has succeeded
	auth.ClearAuthCookies(ctx.Response().Writer, *h.SessionConfig.CookieConfig)

	out := &models.LogoutReply{
		Reply:   rout.Reply{Success: true},
		Message: "logged out successfully",
	}

	return h.Success(ctx, out, openapi)
}

// revokeToken records the token's id on the blacklist for the remainder of its lifetime so it can no
// longer be used. A token that cannot be parsed or has already expired carries nothing to revoke and
// returns nil. A blacklist that is not configured is tolerated and returns nil, since the token
// manager has already logged the warning and logout should still clear the session and cookies; any
// other revocation failure is returned so the caller can fail the request rather than report a
// logout that did not take effect
func (h *Handler) revokeToken(ctx context.Context, token string) error {
	claims, err := h.TokenManager.Parse(token)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("skipping revocation of an unparsable token on logout")

		return nil
	}

	if claims.ID == "" || claims.ExpiresAt == nil {
		return nil
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	if err := h.TokenManager.RevokeToken(ctx, claims.ID, ttl); err != nil && !errors.Is(err, tokens.ErrRevocationNotConfigured) {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to revoke token on logout")

		return err
	}

	return nil
}
