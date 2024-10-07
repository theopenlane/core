package debug

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// BodyDump prints out the request body for debugging purpose but attempts to obfuscate sensitive fields within the requests
func BodyDump() echo.MiddlewareFunc {
	return middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		// Create a child logger for concurrency safety
		logger := log.Logger.With().Logger()

		// Add context fields for the request
		logger.UpdateContext(func(l zerolog.Context) zerolog.Context {
			return l.Str("user-agent", c.Request().Header.Get("User-Agent")).
				Str("request-id", c.Response().Header().Get(echo.HeaderXRequestID)).
				Str("request-uri", c.Request().RequestURI).
				Str("request-method", c.Request().Method).
				Str("request-protocol", c.Request().Proto).
				Str("client-ip", c.RealIP())
		})

		if len(reqBody) > 0 {
			contentType := c.Request().Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "multipart/form-data") {
				// Parse the multipart form
				if err := c.Request().ParseMultipartForm(32 << 20); err == nil { // nolint:mnd
					form := c.Request().MultipartForm
					if form != nil {
						hasFile := false

						for _, files := range form.File {
							if len(files) > 0 {
								hasFile = true
								break
							}
						}

						if hasFile {
							logger.Info().Msg("request contains a file, not logging request body")
						} else {
							logRequestBody(logger, reqBody)
						}
					}
				} else {
					logRequestBody(logger, reqBody)
				}
			} else {
				logRequestBody(logger, reqBody)
			}
		}

		if (c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPatch) && len(resBody) > 0 {
			var bodymap map[string]interface{}
			if err := json.Unmarshal(resBody, &bodymap); err == nil {
				bodymap = redactSecretFields(bodymap)

				resBody, _ = json.Marshal(bodymap)
			}

			logger.Info().Bytes("response body", resBody).Msg("response body")
		}
	})
}

func logRequestBody(logger zerolog.Logger, reqBody []byte) {
	var bodymap map[string]interface{}
	if err := json.Unmarshal(reqBody, &bodymap); err == nil {
		bodymap = redactSecretFields(bodymap)

		reqBody, _ = json.Marshal(bodymap)
	}

	logger.Info().Bytes("request body", reqBody).Msg("request body")
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
