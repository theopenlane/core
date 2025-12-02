package redirect

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// Config contains the types used in executing redirects via the redirect middleware
type Config struct {
	// Enabled indicates if the redirect middleware should be enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper `json:"-" koanf:"-"`
	// Redirects is a map of paths to redirect to
	Redirects map[string]string `json:"redirects" koanf:"redirects"`
	// Code is the HTTP status code to use for the redirect
	Code int `json:"code" koanf:"code"`
}

// DefaultConfig is the default configuration of the redirect middleware
var DefaultConfig = Config{
	Skipper:   middleware.DefaultSkipper,
	Redirects: map[string]string{},
	Code:      0,
}

// New creates a new middleware function with the default config
func New() echo.MiddlewareFunc {
	return NewWithConfig(DefaultConfig)
}

// NewWithConfig returns a new router middleware handler
func NewWithConfig(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if config.Code == 0 {
				config.Code = http.StatusMovedPermanently
			}

			req := c.Request()

			if target, ok := config.Redirects[req.URL.Path]; ok {
				return c.Redirect(config.Code, target)
			}

			return next(c)
		}
	}
}
