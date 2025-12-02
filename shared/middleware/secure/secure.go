package secure

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// Config contains the types used in the mw middleware
type Config struct {
	// Enabled indicates if the secure middleware should be enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper `json:"-" koanf:"-"`
	// XSSProtection is the value to set the X-XSS-Protection header to - default is 1; mode=block
	XSSProtection string `json:"xssprotection" koanf:"xssprotection" default:"1; mode=block"`
	// ContentTypeNosniff is the value to set the X-Content-Type-Options header to - default is nosniff
	ContentTypeNosniff string `json:"contenttypenosniff" koanf:"contenttypenosniff" default:"nosniff"`
	// XFrameOptions is the value to set the X-Frame-Options header to - default is SAMEORIGIN
	XFrameOptions string `json:"xframeoptions" koanf:"xframeoptions" default:"SAMEORIGIN"`
	// HSTSPreloadEnabled is a boolean to enable HSTS preloading - default is false
	HSTSPreloadEnabled bool `json:"hstspreloadenabled" koanf:"hstspreloadenabled" default:"false"`
	// HSTSMaxAge is the max age to set the HSTS header to - default is 31536000
	HSTSMaxAge int `json:"hstsmaxage" koanf:"hstsmaxage" default:"31536000"`
	// ContentSecurityPolicy is the value to set the Content-Security-Policy header to - default is default-src 'self'
	ContentSecurityPolicy string `json:"contentsecuritypolicy" koanf:"contentsecuritypolicy" default:"default-src 'self'"`
	// ReferrerPolicy is the value to set the Referrer-Policy header to - default is same-origin
	ReferrerPolicy string `json:"referrerpolicy" koanf:"referrerpolicy" default:"same-origin"`
	// CSPReportOnly is a boolean to enable the Content-Security-Policy-Report-Only header - default is false
	CSPReportOnly bool `json:"cspreportonly" koanf:"cspreportonly" default:"false"`
}

// DefaultConfig struct is a populated config struct that can be referenced if the default konaf configurations are not available
var DefaultConfig = Config{
	Enabled:               true,
	Skipper:               middleware.DefaultSkipper,
	XSSProtection:         "1; mode=block",
	ContentTypeNosniff:    "nosniff",
	XFrameOptions:         "SAMEORIGIN",
	HSTSPreloadEnabled:    false,
	HSTSMaxAge:            31536000, //nolint:mnd
	ContentSecurityPolicy: "default-src 'self'",
	ReferrerPolicy:        "same-origin",
	CSPReportOnly:         false,
}

// Secure returns a secure middleware with default unless overridden via the config
func Secure(conf *Config) echo.MiddlewareFunc {
	if conf.Enabled {
		secureConfig := middleware.SecureConfig{
			XSSProtection:         conf.XSSProtection,
			ContentTypeNosniff:    conf.ContentTypeNosniff,
			XFrameOptions:         conf.XFrameOptions,
			HSTSPreloadEnabled:    conf.HSTSPreloadEnabled,
			HSTSMaxAge:            conf.HSTSMaxAge,
			ReferrerPolicy:        conf.ReferrerPolicy,
			CSPReportOnly:         conf.CSPReportOnly,
			ContentSecurityPolicy: conf.ContentSecurityPolicy,
			Skipper:               conf.Skipper,
		}

		return middleware.SecureWithConfig(secureConfig)
	}

	return nil
}
