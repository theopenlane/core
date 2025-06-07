package csrf

import (
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// Config defines configuration for the CSRF middleware wrapper.
type Config struct {
	// Enabled indicates whether CSRF protection is enabled.
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Header specifies the header name to look for the CSRF token.
	Header string `json:"header" koanf:"header" default:"X-CSRF-Token"`
	// Cookie specifies the cookie name used to store the CSRF token.
	Cookie string `json:"cookie" koanf:"cookie" default:"csrf_token"`
	// Secure sets the Secure flag on the CSRF cookie.
	Secure bool `json:"secure" koanf:"secure"`
	// SameSite configures the SameSite attribute on the CSRF cookie. Valid
	// values are "Lax", "Strict", "None" and "Default".
	SameSite string `json:"sameSite" koanf:"sameSite" default:"Lax"`
}

// NewConfig returns a Config populated with default values.
func NewConfig() *Config {
	return &Config{
		Enabled:  true,
		Header:   "X-CSRF-Token",
		Cookie:   "csrf_token",
		Secure:   false,
		SameSite: "Lax",
	}
}

// Middleware creates the CSRF middleware from the provided config.
func Middleware(conf *Config) echo.MiddlewareFunc {
	if conf == nil {
		conf = NewConfig()
	}

	if !conf.Enabled {
		return nil
	}

	csrfConf := middleware.CSRFConfig{
		TokenLookup:    "header:" + conf.Header,
		CookieName:     conf.Cookie,
		CookieSecure:   conf.Secure,
		CookieSameSite: parseSameSite(conf.SameSite),
	}

	return middleware.CSRFWithConfig(csrfConf)
}

func parseSameSite(val string) http.SameSite {
	switch strings.ToLower(val) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
