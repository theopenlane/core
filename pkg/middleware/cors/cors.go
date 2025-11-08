package cors

import (
	"fmt"
	"strings"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// Config holds the cors configuration settings
type Config struct {
	// Enable or disable the CORS middleware
	Enabled bool               `json:"enabled" koanf:"enabled" default:"true"`
	Skipper middleware.Skipper `json:"-" koanf:"-"`
	// Prefixes is a map of prefixes to allowed origins
	Prefixes map[string][]string `json:"prefixes" koanf:"prefixes"`
	// AllowOrigins is a list of allowed origins
	AllowOrigins []string `json:"allowOrigins" koanf:"allowOrigins" domain:"inherit" domainPrefix:"https://console,https://docs,https://www"`
	// CookieInsecure sets the cookie to be insecure
	CookieInsecure bool `json:"cookieInsecure" koanf:"cookieInsecure"`
}

// DefaultConfig creates a default config
var DefaultConfig = Config{
	Skipper:  middleware.DefaultSkipper,
	Prefixes: nil,
}

// MustNew creates a new middleware function with the default config or panics if it fails
func MustNew(allowedOrigins []string) echo.MiddlewareFunc {
	DefaultConfig.Prefixes = map[string][]string{
		"/": allowedOrigins,
	}

	mw, err := NewWithConfig(DefaultConfig)
	if err != nil {
		panic("failed to create CORS middleware")
	}

	return mw
}

// NewWithConfig creates a new middleware function with the provided config
func NewWithConfig(config Config) (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}

	prefixes := make(map[string]echo.MiddlewareFunc)

	for prefix, origins := range config.Prefixes {
		if err := Validate(origins); err != nil {
			return nil, fmt.Errorf("CORS config for prefix %s is invalid: %w", prefix, err)
		}

		conf := middleware.CORSConfig{
			AllowOrigins:     origins,
			AllowMethods:     []string{"GET", "HEAD", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-CSRF-Token", "X-User-ID", "X-Organization-ID", "Accept", "Cache-Control"},
			ExposeHeaders:    []string{"Content-Length", "Cache-Control"},
			AllowCredentials: true,                            // Allow credentials to be sent with requests - this is important for CSRF to work
			MaxAge:           int((24 * time.Hour).Seconds()), //nolint:mnd
		}

		prefixes[prefix] = middleware.CORSWithConfig(conf)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			path := c.Request().URL.Path

			var (
				middlewareFunc echo.MiddlewareFunc
				maxPrefixLen   int
			)

			for prefix, h := range prefixes {
				if strings.HasPrefix(path, prefix) {
					if len(prefix) > maxPrefixLen {
						maxPrefixLen = len(prefix)
						middlewareFunc = h
					}
				}
			}

			if middlewareFunc != nil {
				handler := middlewareFunc(next)
				return handler(c)
			}

			return next(c)
		}
	}, nil
}

// DefaultSchemas is a list of default allowed schemas for CORS origins
var DefaultSchemas = []string{
	"http://",
	"https://",
}

// Validate checks a list of origins to see if they comply with the allowed origins
func Validate(origins []string) error {
	for _, origin := range origins {
		if !strings.Contains(origin, "*") && !validateAllowedSchemas(origin) {
			allowed := fmt.Sprintf(" origins must contain '*' or include %s", strings.Join(getAllowedSchemas(), ", or "))

			return newValidationError("bad origin", allowed)
		}
	}

	return nil
}

func validateAllowedSchemas(origin string) bool {
	allowedSchemas := getAllowedSchemas()

	for _, schema := range allowedSchemas {
		if strings.HasPrefix(origin, schema) {
			return true
		}
	}

	return false
}

func getAllowedSchemas() []string {
	allowedSchemas := DefaultSchemas

	return allowedSchemas
}
