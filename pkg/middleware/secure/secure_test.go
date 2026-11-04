package secure_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/theopenlane/echox"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/middleware/secure"
)

func TestSecureMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up the middleware
	mw := secure.Secure(&secure.DefaultConfig)

	// Invoke the middleware
	err := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})(c)

	// Assert that the middleware did not return an error
	assert.Check(t, err)

	// Assert that the response has the expected headers
	assert.Check(t, is.Equal("1; mode=block", rec.Header().Get("X-XSS-Protection")))
	assert.Check(t, is.Equal("nosniff", rec.Header().Get("X-Content-Type-Options")))
	assert.Check(t, is.Equal("SAMEORIGIN", rec.Header().Get("X-Frame-Options")))
	assert.Check(t, is.Equal("default-src 'self'", rec.Header().Get("Content-Security-Policy")))

	// Assert that the response body is "OK"
	assert.Check(t, is.Equal("OK", rec.Body.String()))
}

func TestSecureMiddlewareWithConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	conf := &secure.Config{
		Enabled:       true,
		XSSProtection: "mitb",
	}
	// Set up the middleware
	mw := secure.Secure(conf)

	// Invoke the middleware
	err := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})(c)

	assert.Check(t, err)
	assert.Check(t, is.Equal("mitb", rec.Header().Get("X-XSS-Protection")))
	assert.Check(t, is.Equal("OK", rec.Body.String()))
}
