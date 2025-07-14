package csrf

import (
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
	"github.com/theopenlane/iam/auth"
)

// Config defines configuration for the CSRF middleware wrapper.
type Config struct {
	// Enabled indicates whether CSRF protection is enabled.
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// Header specifies the header name to look for the CSRF token.
	Header string `json:"header" koanf:"header" default:"X-CSRF-Token"`
	// Cookie specifies the cookie name used to store the CSRF token.
	Cookie string `json:"cookie" koanf:"cookie" default:"ol.csrf-token"`
	// Secure sets the Secure flag on the CSRF cookie.
	Secure bool `json:"secure" koanf:"secure" default:"true"`
	// SameSite configures the SameSite attribute on the CSRF cookie. Valid
	// values are "Lax", "Strict", "None" and "Default".
	SameSite string `json:"sameSite" koanf:"sameSite" default:"Lax"`
	// CookieHTTPOnly indicates whether the CSRF cookie is HTTP only.
	CookieHTTPOnly bool `json:"cookieHttpOnly" koanf:"cookieHttpOnly" default:"false"`
	// CookieDomain specifies the domain for the CSRF cookie, default to no domain
	CookieDomain string `json:"cookieDomain" koanf:"cookieDomain" default:""`
	// CookiePath specifies the path for the CSRF cookie, default to "/"
	CookiePath string `json:"cookiePath" koanf:"cookiePath" default:"/"`
}

// NewConfig returns a Config populated with default values.
func NewConfig() *Config {
	return &Config{
		Enabled:        false,
		Header:         "X-CSRF-Token",
		Cookie:         "ol.csrf-token",
		Secure:         true,
		SameSite:       "Lax",
		CookieHTTPOnly: true,
		CookiePath:     "/",
	}
}

// csrfSkipperFunc is the function that determines if the csrf token check should be skipped
// due to the request being a PAT or API Token auth request
var csrfSkipperFunc = func(c echo.Context) bool {
	ac := auth.GetAuthTypeFromEchoContext(c)

	// only skip CSRF checks for API Token or PAT authentication
	return ac == auth.APITokenAuthentication || ac == auth.PATAuthentication
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
		Skipper:        csrfSkipperFunc,
		CookieHTTPOnly: conf.CookieHTTPOnly,
		CookiePath:     conf.CookiePath,
	}

	if conf.CookieDomain != "" {
		csrfConf.CookieDomain = conf.CookieDomain
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
