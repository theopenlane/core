package server

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/logx"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// CustomHTTPErrorHandler is a custom error handler that logs the error and returns a JSON response
func CustomHTTPErrorHandler(c echo.Context, err error) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	// Create a custom error response
	errorResponse := map[string]interface{}{
		"error": err.Error(),
		"query": c.QueryParams(),
		"url":   c.Request().URL.String(),
	}

	if err, ok := err.(stackTracer); ok {
		errorResponse["stackTrace"] = err.StackTrace()
	}

	logx.FromContext(c.Request().Context()).Error().
		Err(err).
		Str("query", fmt.Sprintf("%v", c.QueryParams())).
		Str("url", c.Request().URL.String()).
		Msgf("Error handling %s to %s", c.Request().Method, c.Request().URL.String())

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, errorResponse)
		}
	}

	_ = err
}
