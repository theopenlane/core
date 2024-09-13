package debug

import (
	"encoding/json"
	"net/http"

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
			var bodymap map[string]interface{}
			if err := json.Unmarshal(reqBody, &bodymap); err == nil {
				bodymap = redactSecretFields(bodymap)

				reqBody, _ = json.Marshal(bodymap)
			}

			logger.Info().Bytes("request body", reqBody).Msg("request body")
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
