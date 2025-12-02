package cachecontrol_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/middleware/cachecontrol"
)

var epoch = time.Unix(0, 0).Format(time.RFC1123)

func TestNewWithConfig(t *testing.T) {
	// Create a new Echo instance
	e := echo.New()

	// Create a request with a dummy handler
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a dummy handler
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	// Create a middleware with default config
	middleware := cachecontrol.NewWithConfig(cachecontrol.DefaultConfig)

	// Wrap the dummy handler with the middleware
	wrappedHandler := middleware(handler)

	// Invoke the wrapped handler
	err := wrappedHandler(c)

	// Assert that the handler returned no error
	assert.NoError(t, err)

	// Assert that the response has the expected status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Assert that the response headers have been set correctly
	expectedHeaders := map[string]string{
		"Cache-Control": "no-cache, private, max-age=0",
		"Pragma":        "no-cache",
		"Expires":       epoch,
	}
	for k, v := range expectedHeaders {
		assert.Equal(t, v, rec.Header().Get(k))
	}
}
