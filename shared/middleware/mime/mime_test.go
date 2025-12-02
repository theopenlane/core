package mime_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/middleware/mime"
)

func TestNewWithConfig(t *testing.T) {
	// Create a new Echo instance
	e := echo.New()

	// Create a request with a URL path
	req := httptest.NewRequest(http.MethodGet, "/path/to/file.html", nil)

	// Create a response recorder to capture the response
	rec := httptest.NewRecorder()

	// Create a context from the request and response recorder
	c := e.NewContext(req, rec)

	// Create a middleware configuration
	config := mime.Config{
		MimeTypesFile:      "mime.types",
		DefaultContentType: "text/plain",
	}

	// Create the middleware using the configuration
	middleware := mime.NewWithConfig(config)

	// Create a handler function that will be called after the middleware
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	// Invoke the middleware with the handler function
	err := middleware(handler)(c)

	// Assert that the middleware did not return an error
	assert.NoError(t, err)

	// Assert that the response header has the expected content type
	assert.Equal(t, "text/plain", rec.Header().Get(echo.HeaderContentType))
}
