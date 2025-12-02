package secure_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/middleware/secure"
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
	assert.NoError(t, err)

	// Assert that the response has the expected headers
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))

	// Assert that the response body is "OK"
	assert.Equal(t, "OK", rec.Body.String())
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

	assert.NoError(t, err)
	assert.Equal(t, "mitb", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "OK", rec.Body.String())
}
