package cors_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/theopenlane/echox"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/middleware/cors"
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
	assert.NilError(t, err)

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
	assert.NilError(t, err)

	assert.Check(t, is.Equal("*", rec.Header().Get(allowOriginHeader)))
	assert.Check(t, is.Len(rec.Header().Get(allowCredentialsHeader), 0))
}

// TestPublicPathPreflight verifies a preflight request to a public path advertises the wildcard
// origin and the allowed methods so browsers permit the cross-origin call
func TestPublicPathPreflight(t *testing.T) {
	cors.RegisterPublicPath("/v1/unsubscribe")

	mw, err := cors.NewWithConfig(prefixConfig("https://console.example.com"))
	assert.NilError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/v1/unsubscribe", nil)
	req.Header.Set("Origin", "https://random-customer-domain.example.org")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NilError(t, mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c))

	assert.Check(t, is.Equal("*", rec.Header().Get(allowOriginHeader)))
	assert.Check(t, is.Contains(rec.Header().Get(allowMethodsHeader), http.MethodPost))
	assert.Check(t, is.Len(rec.Header().Get(allowCredentialsHeader), 0))
}

// TestPrefixPathEchoesAllowedOriginWithCredentials verifies non-public paths keep the existing
// behavior: a configured origin is echoed back and credentials remain allowed
func TestPrefixPathEchoesAllowedOriginWithCredentials(t *testing.T) {
	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/v1/other", "https://console.example.com")
	assert.NilError(t, err)

	assert.Check(t, is.Equal("https://console.example.com", rec.Header().Get(allowOriginHeader)))
	assert.Check(t, is.Equal("true", rec.Header().Get(allowCredentialsHeader)))
}

// TestPrefixPathRejectsUnknownOrigin verifies non-public paths reject an unconfigured origin with
// 401 (echox returns Unauthorized for disallowed non-preflight origins) and grant it nothing
func TestPrefixPathRejectsUnknownOrigin(t *testing.T) {
	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/v1/other", "https://random-customer-domain.example.org")

	assert.Assert(t, err != nil)

	var httpErr *echo.HTTPError
	assert.Assert(t, errors.As(err, &httpErr))
	assert.Check(t, is.Equal(http.StatusUnauthorized, httpErr.Code))

	assert.Check(t, is.Len(rec.Header().Get(allowOriginHeader), 0))
}

// TestUnmatchedPathSkipsCORS verifies a path outside every configured prefix and the public set
// passes through without CORS headers
func TestUnmatchedPathSkipsCORS(t *testing.T) {
	rec, err := serveCORS(t, prefixConfig("https://console.example.com"), http.MethodGet, "/metrics", "https://console.example.com")
	assert.NilError(t, err)

	assert.Check(t, is.Len(rec.Header().Get(allowOriginHeader), 0))
}
