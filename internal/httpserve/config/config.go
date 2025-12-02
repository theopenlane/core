package config

import (
	"crypto/tls"
	"net/http"
	"time"

	echo "github.com/theopenlane/echox"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/core/config"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/objects/objstore"
)

var (
	// DefaultConfigRefresh sets the default interval to refresh the config.
	DefaultConfigRefresh = 10 * time.Minute
	// DefaultTLSConfig is the default TLS config used when HTTPS is enabled
	DefaultTLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
)

// Config is the configuration for the http server
type Config struct {
	// add all the configuration settings for the server
	Settings config.Config
	// Routes contains the handler functions
	Routes []http.Handler
	// DefaultMiddleware to enable on the echo server used on all requests
	DefaultMiddleware []echo.MiddlewareFunc
	// GraphMiddleware to enable on the echo server used on graph requests
	GraphMiddleware []echo.MiddlewareFunc
	// Handler contains the required settings for REST handlers including ready checks and JWT keys
	Handler handlers.Handler
	// SessionConfig manages sessions for users
	SessionConfig *sessions.SessionConfig
	// StorageService manages objects for the server
	StorageService *objstore.Service
}

// Ensure that *Config implements ConfigProvider interface.
var _ Provider = &Config{}

// Get implements ConfigProvider.
func (c *Config) Get() (*Config, error) {
	return c, nil
}

// WithTLSDefaults sets tls default settings assuming a default cert and key file location.
func (c Config) WithTLSDefaults() Config {
	c.WithDefaultTLSConfig()

	return c
}

// WithDefaultTLSConfig sets the default TLS Configuration
func (c Config) WithDefaultTLSConfig() Config {
	c.Settings.Server.TLS.Enabled = true
	c.Settings.Server.TLS.Config = DefaultTLSConfig

	return c
}

// WithTLSCerts sets the TLS Cert and Key locations
func (c *Config) WithTLSCerts(certFile, certKey string) *Config {
	c.Settings.Server.TLS.CertFile = certFile
	c.Settings.Server.TLS.CertKey = certKey

	return c
}

// WithAutoCert generates a letsencrypt certificate, a valid host must be provided
func (c *Config) WithAutoCert(host string) *Config {
	autoTLSManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
		Cache:      autocert.DirCache("/var/www/.cache"),
		HostPolicy: autocert.HostWhitelist(host),
	}

	c.Settings.Server.TLS.Enabled = true
	c.Settings.Server.TLS.Config = DefaultTLSConfig

	c.Settings.Server.TLS.Config.GetCertificate = autoTLSManager.GetCertificate
	c.Settings.Server.TLS.Config.NextProtos = []string{acme.ALPNProto}

	return c
}
