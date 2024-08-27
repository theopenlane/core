package debug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
	"go.uber.org/zap"
)

// BodyDump prints out the request body for debugging purpose but attempts to obfuscate sensitive fields within the requests
func BodyDump(l *zap.SugaredLogger) echo.MiddlewareFunc {
	return middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		secretFields := []string{"new_password", "old_password", "password", "access_token", "refresh_token"}

		if len(reqBody) > 0 {
			var bodymap map[string]interface{}
			if err := json.Unmarshal(reqBody, &bodymap); err == nil {
				for i := 0; i < len(secretFields); i++ {
					if _, ok := bodymap[secretFields[i]]; ok {
						bodymap[secretFields[i]] = "********"
					}
				}

				reqBody, _ = json.Marshal(bodymap)

				var reqMethod string

				var methodColor, resetColor string

				req := c.Request()
				reqMethod = req.Method
				// for request method
				switch reqMethod {
				case "GET":
					methodColor = blue
				case "POST":
					methodColor = cyan
				case "PUT":
					methodColor = yellow
				case "DELETE":
					methodColor = red
				case "PATCH":
					methodColor = green
				case "HEAD":
					methodColor = magenta
				case "OPTIONS":
					methodColor = white
				default:
					methodColor = reset
				}
				// reset to return to the normal terminal color variables (kinda default)
				resetColor = reset

				name := "REQUEST"
				fmt.Printf("\n[%s] %v | %8s | %10s |%s %-7s %s %s\n",
					name, // request
					time.Now().Format("2006/01/02 - 15:04:05"), // TIMESTAMP for route access
					req.Proto,                          // protocol
					c.RealIP(),                         // client IP
					methodColor, reqMethod, resetColor, // request method
					req.URL, // request URI (path)
				)
			}

			l.Infof("Request Body: %v\n", string(reqBody))
		}

		if (c.Request().Method == "PATCH" || c.Request().Method == "POST") && len(resBody) > 0 {
			var bodymap map[string]interface{}
			if err := json.Unmarshal(resBody, &bodymap); err == nil {
				for i := 0; i < len(secretFields); i++ {
					if _, ok := bodymap[secretFields[i]]; ok {
						bodymap[secretFields[i]] = "********"
					}
				}

				resBody, _ = json.Marshal(bodymap)

				var resStatus int

				var statusColor, resetColor string

				res := c.Response()
				resStatus = res.Status

				switch {
				case resStatus >= http.StatusOK && resStatus < http.StatusMultipleChoices:
					statusColor = green
				case resStatus >= http.StatusMultipleChoices && resStatus < http.StatusBadRequest:
					statusColor = white
				case resStatus >= http.StatusBadRequest && resStatus < http.StatusInternalServerError:
					statusColor = yellow
				default:
					statusColor = red
				}
				// reset to return to the normal terminal color variables (kinda default)
				resetColor = reset

				name := "RESPONSE"
				fmt.Printf("\n[%s] %v |%s %3d %s| %s\n",
					name, // response
					time.Now().Format("2006/01/02 - 15:04:05"), // TIMESTAMP for route access
					statusColor, resStatus, resetColor, // response status
					res.Header(),
				)
			}

			l.Infof("Response Body: %v\n", string(resBody))
		}
	})
}
