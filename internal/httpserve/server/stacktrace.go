package server

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
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
		// "stackTrace": getStackTrace(),
		"query": c.QueryParams(),
		"url":   c.Request().URL.String(),
	}

	if err, ok := err.(stackTracer); ok {
		errorResponse["stackTrace"] = err.StackTrace()
	}

	log.Error().
		Err(err).
		Str("query", fmt.Sprintf("%v", c.QueryParams())).
		Str("url", c.Request().URL.String()).
		Msg("Error")

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == "HEAD" { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, errorResponse)
		}
	}

	_ = err
}
