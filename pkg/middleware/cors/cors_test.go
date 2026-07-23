package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/middleware/cors"
)

const (
	allowOriginHeader      = "Access-Control-Allow-Origin"
	allowCredentialsHeader = "Access-Control-Allow-Credentials"
	allowMethodsHeader     = "Access-Control-Allow-Methods"
)

// serveCORS runs a request through the CORS middleware built from config, returning the recorder
// and the handler chain error so rejection behavior can be asserted
func serveCORS(t *testing.T, config cors.Config, method, path, origin string) (*httptest.ResponseRecorder, error) {
	t.Helper()

	mw, err := cors.NewWithConfig(config)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Origin", origin)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	return rec, mw(handler)(c)
}

func prefixConfig(origins ...string) cors.Config {
	return cors.Config{
		Prefixes: map[string][]string{"/v1": origins},
	}
}

// TestPublicPathAllowsAnyOriginWithoutCredentials verifies a registered public path answers any
// origin with a wildcard and never allows credentials, regardless of the configured prefix origins
func TestPublicPathAllowsAnyOriginWithoutCredentials(t *testing.T) {
	cors.RegisterPublicPath("/v1/subscribe/verify")

	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/v1/subscribe/verify", "https://random-customer-domain.example.org")
	require.NoError(t, err)

	assert.Equal(t, "*", rec.Header().Get(allowOriginHeader))
	assert.Empty(t, rec.Header().Get(allowCredentialsHeader))
}

// TestPublicPathPreflight verifies a preflight request to a public path advertises the wildcard
// origin and the allowed methods so browsers permit the cross-origin call
func TestPublicPathPreflight(t *testing.T) {
	cors.RegisterPublicPath("/v1/unsubscribe")

	mw, err := cors.NewWithConfig(prefixConfig("https://console.example.com"))
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/v1/unsubscribe", nil)
	req.Header.Set("Origin", "https://random-customer-domain.example.org")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c))

	assert.Equal(t, "*", rec.Header().Get(allowOriginHeader))
	assert.Contains(t, rec.Header().Get(allowMethodsHeader), http.MethodPost)
	assert.Empty(t, rec.Header().Get(allowCredentialsHeader))
}

// TestPrefixPathEchoesAllowedOriginWithCredentials verifies non-public paths keep the existing
// behavior: a configured origin is echoed back and credentials remain allowed
func TestPrefixPathEchoesAllowedOriginWithCredentials(t *testing.T) {
	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/v1/other", "https://console.example.com")
	require.NoError(t, err)

	assert.Equal(t, "https://console.example.com", rec.Header().Get(allowOriginHeader))
	assert.Equal(t, "true", rec.Header().Get(allowCredentialsHeader))
}

// TestPrefixPathRejectsUnknownOrigin verifies non-public paths reject an unconfigured origin with
// 401 (echox returns Unauthorized for disallowed non-preflight origins) and grant it nothing
func TestPrefixPathRejectsUnknownOrigin(t *testing.T) {
	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/v1/other", "https://random-customer-domain.example.org")

	require.Error(t, err)

	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)

	assert.Empty(t, rec.Header().Get(allowOriginHeader))
}

// TestUnmatchedPathSkipsCORS verifies a path outside every configured prefix and the public set
// passes through without CORS headers
func TestUnmatchedPathSkipsCORS(t *testing.T) {
	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/metrics", "https://console.example.com")
	require.NoError(t, err)

	assert.Empty(t, rec.Header().Get(allowOriginHeader))
}
