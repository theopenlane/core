//go:build test

package handlers_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	coreutils "github.com/theopenlane/core/internal/testutils"
)

// newSessionConfig builds a session config the way serveropts.WithSessionManager does for a
// deployed server: a secure, named cookie config is shared between the cookie store and the session
// config so that cookies written by CreateAndStoreSession round-trip back through Get. A fresh
// NewCookieConfig(true) is used rather than mutating the shared DefaultCookieConfig global
func newSessionConfig(t *testing.T, client *redis.Client) sessions.SessionConfig {
	t.Helper()

	cc := sessions.NewCookieConfig(true)
	cc.Name = sessions.DefaultCookieName

	hashKey := make([]byte, 32)
	blockKey := make([]byte, 32)

	_, err := rand.Read(hashKey)
	assert.NoError(t, err)

	_, err = rand.Read(blockKey)
	assert.NoError(t, err)

	sm := sessions.NewCookieStore[map[string]any](cc, hashKey, blockKey)

	sc := sessions.NewSessionConfig(sm, sessions.WithPersistence(client))
	sc.CookieConfig = cc

	return sc
}

// TestLogoutHandler verifies that logout revokes the access and refresh tokens, deletes the
// server-side session, and clears the auth cookies
func TestLogoutHandler(t *testing.T) {
	client := coreutils.NewRedisClient()
	defer client.Close()

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute)
	assert.NoError(t, err)

	tm.WithBlacklist(tokens.NewRedisTokenBlacklist(client, "token:blacklist:"))
	assert.True(t, tm.RevocationEnabled())

	sc := newSessionConfig(t, client)

	h := &handlers.Handler{
		TokenManager:  tm,
		SessionConfig: &sc,
		RedisClient:   client,
	}

	// create the token pair the caller will log out with
	claims := &tokens.Claims{UserID: "user-123", OrgID: "org-456"}

	access, refresh, err := tm.CreateTokenPair(claims)
	assert.NoError(t, err)

	// the access and refresh token in a pair share a jwt id, so one blacklist entry covers both
	accessClaims, err := tokens.ParseUnverifiedTokenClaims(access)
	assert.NoError(t, err)

	jti := accessClaims.ID
	assert.NotEmpty(t, jti)

	// create a persisted session and capture its cookie so the handler can delete it
	sessionRec := httptest.NewRecorder()
	_, err = sc.CreateAndStoreSession(context.Background(), sessionRec, claims.UserID)
	assert.NoError(t, err)

	// build the logout request with the bearer access token, refresh token body, and session cookie
	body := fmt.Sprintf(`{"refresh_token":%q}`, refresh)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/logout", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access)

	for _, c := range sessionRec.Result().Cookies() {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	// resolve the session id and confirm both the token and session are live beforehand
	sessionForReq, err := sc.SessionManager.Get(req, sc.CookieConfig.Name)
	assert.NoError(t, err)

	sessionID := sc.SessionManager.GetSessionIDFromCookie(sessionForReq)
	assert.NotEmpty(t, sessionID)

	_, err = sc.RedisStore.GetSession(context.Background(), sessionID)
	assert.NoError(t, err)

	revoked, err := tm.IsTokenRevoked(context.Background(), jti)
	assert.NoError(t, err)
	assert.False(t, revoked)

	// perform the logout
	err = h.LogoutHandler(ctx, &handlers.OpenAPIContext{Operation: &openapi3.Operation{}})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// the shared jwt id is now revoked, so both the access and refresh tokens are rejected
	revoked, err = tm.IsTokenRevoked(context.Background(), jti)
	assert.NoError(t, err)
	assert.True(t, revoked)

	// the server-side session was deleted
	_, err = sc.RedisStore.GetSession(context.Background(), sessionID)
	assert.Error(t, err)

	// the auth cookies were cleared on the response
	cleared := map[string]bool{}

	for _, c := range rec.Result().Cookies() {
		if c.MaxAge < 0 || c.Value == "" {
			cleared[c.Name] = true
		}
	}

	assert.True(t, cleared[auth.AccessTokenCookie], "access token cookie should be cleared")
	assert.True(t, cleared[auth.RefreshTokenCookie], "refresh token cookie should be cleared")
}

// TestLogoutHandlerWithoutCredentials verifies that logout succeeds and is idempotent when no
// tokens or session are presented
func TestLogoutHandlerWithoutCredentials(t *testing.T) {
	client := coreutils.NewRedisClient()
	defer client.Close()

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute)
	assert.NoError(t, err)

	tm.WithBlacklist(tokens.NewRedisTokenBlacklist(client, "token:blacklist:"))

	sc := newSessionConfig(t, client)

	h := &handlers.Handler{
		TokenManager:  tm,
		SessionConfig: &sc,
		RedisClient:   client,
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/logout", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	err = h.LogoutHandler(ctx, &handlers.OpenAPIContext{Operation: &openapi3.Operation{}})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
