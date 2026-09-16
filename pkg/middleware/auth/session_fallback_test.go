//go:build test

package auth_test

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox"
	"github.com/theopenlane/utils/ulids"

	iamauth "github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/middleware/auth"
)

func TestSessionFallbackUserID(t *testing.T) {
	t.Run("resolves the caller subject", func(t *testing.T) {
		subjectID := ulids.New().String()
		ctx := iamauth.WithCaller(context.Background(), &iamauth.Caller{SubjectID: subjectID})

		userID, ok := auth.SessionFallbackUserID(ctx)
		assert.True(t, ok)
		assert.Equal(t, subjectID, userID)
	})

	t.Run("declines without a caller", func(t *testing.T) {
		userID, ok := auth.SessionFallbackUserID(context.Background())
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("declines an impersonated caller", func(t *testing.T) {
		ctx := iamauth.WithCaller(context.Background(), &iamauth.Caller{
			SubjectID:     ulids.New().String(),
			Impersonation: &iamauth.ImpersonationContext{ImpersonatorID: ulids.New().String()},
		})

		userID, ok := auth.SessionFallbackUserID(ctx)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("declines a subject that is not a user id", func(t *testing.T) {
		ctx := iamauth.WithCaller(context.Background(), &iamauth.Caller{SubjectID: "anon_trustcenter_" + ulids.New().String()})

		userID, ok := auth.SessionFallbackUserID(ctx)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}

// newSessionMiddlewareConfig builds the session config the way serveropts wires it for graph routes
func newSessionMiddlewareConfig(t *testing.T, client *redis.Client, opts ...sessions.Option) sessions.SessionConfig {
	t.Helper()

	cc := sessions.NewCookieConfig(true)
	cc.Name = sessions.DefaultCookieName

	hashKey := make([]byte, 32)
	blockKey := make([]byte, 32)

	_, err := rand.Read(hashKey)
	require.NoError(t, err)

	_, err = rand.Read(blockKey)
	require.NoError(t, err)

	sm := sessions.NewCookieStore[map[string]any](cc, hashKey, blockKey)

	sc := sessions.NewSessionConfig(sm, append([]sessions.Option{sessions.WithPersistence(client)}, opts...)...)
	sc.CookieConfig = cc

	return sc
}

// graphSessionConfig mirrors the production graph route session config with the caller fallback
func graphSessionConfig(t *testing.T, client *redis.Client) sessions.SessionConfig {
	t.Helper()

	return newSessionMiddlewareConfig(t, client,
		sessions.WithSkipperFunc(auth.SessionSkipperFunc),
		sessions.WithFallbackUserID(auth.SessionFallbackUserID),
	)
}

func jwtCaller(subjectID string) *iamauth.Caller {
	return &iamauth.Caller{SubjectID: subjectID, AuthenticationType: iamauth.JWTAuthentication}
}

// runSessionMiddleware sends one request through the session middleware as the given caller
func runSessionMiddleware(t *testing.T, sc sessions.SessionConfig, caller *iamauth.Caller, cookies ...*http.Cookie) (*httptest.ResponseRecorder, error) {
	t.Helper()

	ctx := context.Background()
	if caller != nil {
		ctx = iamauth.WithCaller(ctx, caller)
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/query", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	c := echox.New().NewContext(req, rec)

	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	return rec, sessions.LoadAndSaveWithConfig(sc)(handler)(c)
}

func sessionCookies(rec *httptest.ResponseRecorder, name string) []*http.Cookie {
	var found []*http.Cookie

	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			found = append(found, c)
		}
	}

	return found
}

func sessionIDFromCookie(t *testing.T, sc sessions.SessionConfig, cookie *http.Cookie) string {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	session, err := sc.SessionManager.Get(req, sc.CookieConfig.Name)
	require.NoError(t, err)

	return sc.SessionManager.GetSessionIDFromCookie(session)
}

// mintedCookie creates a persisted session for the user and returns its cookie
func mintedCookie(t *testing.T, sc sessions.SessionConfig, userID string) *http.Cookie {
	t.Helper()

	rec := httptest.NewRecorder()
	_, err := sc.CreateAndStoreSession(context.Background(), rec, userID)
	require.NoError(t, err)

	cookies := sessionCookies(rec, sc.CookieConfig.Name)
	require.Len(t, cookies, 1)

	return cookies[0]
}

func TestGraphSessionMiddlewareFallback(t *testing.T) {
	ctx := context.Background()

	t.Run("mints a session for a jwt caller with no session cookie", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		userID := ulids.New().String()

		rec, err := runSessionMiddleware(t, sc, jwtCaller(userID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		cookies := sessionCookies(rec, sc.CookieConfig.Name)
		require.Len(t, cookies, 1)

		sessionID := sessionIDFromCookie(t, sc, cookies[0])

		stored, err := client.Get(ctx, sessionID).Result()
		require.NoError(t, err)
		assert.Equal(t, userID, stored)

		ttl, err := client.TTL(ctx, sessionID).Result()
		require.NoError(t, err)
		assert.InDelta(t, sc.CookieConfig.MaxAge, ttl.Seconds(), 1)
	})

	t.Run("replaces a session that idled out of the store", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		userID := ulids.New().String()

		stale := mintedCookie(t, sc, userID)
		staleID := sessionIDFromCookie(t, sc, stale)
		require.NoError(t, client.Del(ctx, staleID).Err())

		rec, err := runSessionMiddleware(t, sc, jwtCaller(userID), stale)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		cookies := sessionCookies(rec, sc.CookieConfig.Name)
		require.Len(t, cookies, 1)
		assert.NotEqual(t, staleID, sessionIDFromCookie(t, sc, cookies[0]))
	})

	t.Run("replaces a session cookie it cannot decode", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		userID := ulids.New().String()

		garbage := &http.Cookie{Name: sc.CookieConfig.Name, Value: "not-a-session"}

		rec, err := runSessionMiddleware(t, sc, jwtCaller(userID), garbage)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, sessionCookies(rec, sc.CookieConfig.Name), 1)
	})

	t.Run("survives a stale duplicate cookie shadowing a live one", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		userID := ulids.New().String()

		stale := mintedCookie(t, sc, userID)
		require.NoError(t, client.Del(ctx, sessionIDFromCookie(t, sc, stale)).Err())
		live := mintedCookie(t, sc, userID)

		rec, err := runSessionMiddleware(t, sc, jwtCaller(userID), stale, live)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, sessionCookies(rec, sc.CookieConfig.Name), 1)
	})

	t.Run("minted session is loaded on the next request without the fallback", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		userID := ulids.New().String()

		first, err := runSessionMiddleware(t, sc, jwtCaller(userID))
		require.NoError(t, err)

		minted := sessionCookies(first, sc.CookieConfig.Name)
		require.Len(t, minted, 1)
		mintedID := sessionIDFromCookie(t, sc, minted[0])

		strict := newSessionMiddlewareConfig(t, client, sessions.WithSkipperFunc(auth.SessionSkipperFunc))
		strict.SessionManager = sc.SessionManager

		second, err := runSessionMiddleware(t, strict, jwtCaller(userID), minted[0])
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, second.Code)

		slid := sessionCookies(second, sc.CookieConfig.Name)
		require.Len(t, slid, 1)

		slidID := sessionIDFromCookie(t, sc, slid[0])
		assert.NotEqual(t, mintedID, slidID)

		stored, err := client.Get(ctx, slidID).Result()
		require.NoError(t, err)
		assert.Equal(t, userID, stored)
	})

	t.Run("rejects an impersonated caller with no session", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		caller := jwtCaller(ulids.New().String())
		caller.Impersonation = &iamauth.ImpersonationContext{ImpersonatorID: ulids.New().String()}

		rec, err := runSessionMiddleware(t, sc, caller)
		require.ErrorIs(t, err, sessions.ErrInvalidSession)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Empty(t, sessionCookies(rec, sc.CookieConfig.Name))
	})

	t.Run("loads the impersonator's own session for an impersonated caller", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)
		impersonatorID := ulids.New().String()
		own := mintedCookie(t, sc, impersonatorID)

		caller := jwtCaller(ulids.New().String())
		caller.Impersonation = &iamauth.ImpersonationContext{ImpersonatorID: impersonatorID}

		rec, err := runSessionMiddleware(t, sc, caller, own)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		slid := sessionCookies(rec, sc.CookieConfig.Name)
		require.Len(t, slid, 1)

		stored, err := client.Get(ctx, sessionIDFromCookie(t, sc, slid[0])).Result()
		require.NoError(t, err)
		assert.Equal(t, impersonatorID, stored)
	})

	t.Run("skips sessions for non jwt callers", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := graphSessionConfig(t, client)

		for _, caller := range []*iamauth.Caller{
			{SubjectID: ulids.New().String(), AuthenticationType: iamauth.PATAuthentication},
			{SubjectID: ulids.New().String(), AuthenticationType: iamauth.APITokenAuthentication},
			iamauth.NewTrustCenterCaller(ulids.New().String(), "anon_trustcenter_"+ulids.New().String(), "Anonymous User", ""),
		} {
			rec, err := runSessionMiddleware(t, sc, caller)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Empty(t, sessionCookies(rec, sc.CookieConfig.Name))
		}
	})

	t.Run("still rejects a stale session without the fallback", func(t *testing.T) {
		client := coreutils.NewRedisClient()
		defer client.Close()

		sc := newSessionMiddlewareConfig(t, client, sessions.WithSkipperFunc(auth.SessionSkipperFunc))
		userID := ulids.New().String()

		stale := mintedCookie(t, sc, userID)
		require.NoError(t, client.Del(ctx, sessionIDFromCookie(t, sc, stale)).Err())

		rec, err := runSessionMiddleware(t, sc, jwtCaller(userID), stale)
		require.ErrorIs(t, err, sessions.ErrInvalidSession)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
