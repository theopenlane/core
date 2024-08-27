package auth

import (
	"context"
	"net/http"
	"regexp"
	"time"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/sessions"
)

const (
	// Authorization is the key used in HTTP headers or cookies to represent the authorization token
	Authorization = "Authorization"
	// AccessTokenCookie is the key used in cookies to represent the access token
	AccessTokenCookie = "access_token"
	// RefreshTokenCookie is the key used in cookies to represent the refresh token
	RefreshTokenCookie = "refresh_token"
)

// used to extract the access token from the header
var (
	bearer = regexp.MustCompile(`^\s*[Bb]earer\s+([a-zA-Z0-9_\-\.]+)\s*$`)
)

// GetAccessToken retrieves the bearer token from the authorization header and parses it
// to return only the JWT access token component of the header. Alternatively, if the
// authorization header is not present, then the token is fetched from cookies. If the
// header is missing or the token is not available, an error is returned.
//
// NOTE: the authorization header takes precedence over access tokens in cookies.
func GetAccessToken(c echo.Context) (string, error) {
	// Attempt to get the access token from the header.
	if h := c.Request().Header.Get(Authorization); h != "" {
		match := bearer.FindStringSubmatch(h)
		if len(match) == 2 { //nolint:mnd
			return match[1], nil
		}

		return "", ErrParseBearer
	}

	// Attempt to get the access token from cookies.
	if cookie, err := c.Cookie(AccessTokenCookie); err == nil {
		// If the error is nil, that means we were able to retrieve the access token cookie
		if CookieExpired(cookie) {
			return "", ErrNoAuthorization
		}

		return cookie.Value, nil
	}

	return "", ErrNoAuthorization
}

// GetRefreshToken retrieves the refresh token from the cookies in the request. If the
// cookie is not present or expired then an error is returned.
func GetRefreshToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie(RefreshTokenCookie)
	if err != nil {
		return "", ErrNoRefreshToken
	}

	// ensure cookie is not expired
	if CookieExpired(cookie) {
		return "", ErrNoRefreshToken
	}

	return cookie.Value, nil
}

// AuthContextFromRequest creates a context from the echo request context, copying fields
// that may be required for forwarded requests. This method should be called by
// handlers which need to forward requests to other services and need to preserve data
// from the original request such as the user's credentials.
func AuthContextFromRequest(c echo.Context) (*context.Context, error) {
	req := c.Request()
	if req == nil {
		return nil, ErrNoRequest
	}

	// Add access token to context (from either header or cookie using Authenticate middleware)
	ctx := req.Context()
	if token := c.Get(ContextAccessToken.name); token != "" {
		ctx = context.WithValue(ctx, ContextAccessToken, token)
	}

	// Add request id to context
	if requestID := c.Get(ContextRequestID.name); requestID != "" {
		ctx = context.WithValue(ctx, ContextRequestID, requestID)
	} else if requestID := c.Request().Header.Get("X-Request-ID"); requestID != "" {
		ctx = context.WithValue(ctx, ContextRequestID, requestID)
	}

	return &ctx, nil
}

// SetAuthCookies is a helper function to set authentication cookies on a echo request.
// The access token cookie (access_token) is an http only cookie that expires when the
// access token expires. The refresh token cookie is not an http only cookie (it can be
// accessed by client-side scripts) and it expires when the refresh token expires. Both
// cookies require https and will not be set (silently) over http connections.
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, c sessions.CookieConfig) {
	sessions.SetCookie(w, accessToken, AccessTokenCookie, c)
	sessions.SetCookie(w, refreshToken, RefreshTokenCookie, c)
}

// ClearAuthCookies is a helper function to clear authentication cookies on a echo
// request to effectively logger out a user.
func ClearAuthCookies(w http.ResponseWriter) {
	sessions.RemoveCookie(w, AccessTokenCookie, *sessions.DefaultCookieConfig)
	sessions.RemoveCookie(w, RefreshTokenCookie, *sessions.DefaultCookieConfig)
}

// CookieExpired checks to see if a cookie is expired
func CookieExpired(cookie *http.Cookie) bool {
	// ensure cookie is not expired
	if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
		return true
	}

	// negative max age means to expire immediately
	if cookie.MaxAge < 0 {
		return true
	}

	return false
}
