package debug

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"

	"github.com/theopenlane/shared/objects"
)

// BodyDump prints out the request body for debugging purpose but attempts to obfuscate sensitive fields within the requests
func BodyDump() echo.MiddlewareFunc {
	return middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		r := c.Request()
		w := c.Response()
		ctx := r.Context()

		// Create a child logger for concurrency safety
		logger := log.Logger.With().Logger()

		// Add context fields for the request
		logger.UpdateContext(func(l zerolog.Context) zerolog.Context {
			return l.Str("request-id", w.Header().Get(echo.HeaderXRequestID))
		})

		// Log the request body if it is not empty and the content type is not multipart/form-data
		if shouldLogBody(ctx, logger, reqBody) {
			logRequestBody(logger, reqBody)
		}

		if (c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPatch) && len(resBody) > 0 {
			var bodymap map[string]interface{}
			if err := json.Unmarshal(resBody, &bodymap); err == nil {
				bodymap = redactSecretFields(bodymap)

				resBody, _ = json.Marshal(bodymap)
			}

			logger.Info().Bytes("RESPONSE_BODY", resBody).Msg("response_body")
		}
	})
}

// shouldLogBody determines if the request body should be logged based on the content type and the presence of files in the request
func shouldLogBody(ctx context.Context, logger zerolog.Logger, reqBody []byte) bool {
	// If the request body is empty, there is nothing to log
	if len(reqBody) == 0 {
		return false
	}

	files, _ := objects.FilesFromContext(ctx)
	if len(files) > 0 {
		logger.Info().Msg("request contains a file, not logging request body")

		return false
	}

	// if we can json unmarshal the request body, it is not a file upload
	var bodymap map[string]interface{}
	if err := json.Unmarshal(reqBody, &bodymap); err == nil {
		return true
	}

	// default to not logging the request body
	logger.Info().Msg("request cannot be unmarshalled, not logging request body")

	return false
}

// logRequestBody logs the request body to the logger
func logRequestBody(logger zerolog.Logger, reqBody []byte) {
	var bodymap map[string]interface{}
	if err := json.Unmarshal(reqBody, &bodymap); err == nil {
		bodymap = redactSecretFields(bodymap)

		reqBody, _ = json.Marshal(bodymap)
	}

	logger.Info().Bytes("REQUEST_BODY", reqBody).Msg("request_body")
}

// redactSecretFields redacts sensitive fields from the request body
func redactSecretFields(bodymap map[string]interface{}) map[string]interface{} {
	secretFields := []string{"new_password", "old_password", "password", "access_token", "refresh_token"}

	for i := 0; i < len(secretFields); i++ {
		if _, ok := bodymap[secretFields[i]]; ok {
			bodymap[secretFields[i]] = "********"
		}
	}

	return bodymap
}
