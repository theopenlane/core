package webauthn

import (
	"encoding/gob"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	ProviderName = "WEBAUTHN"
)

// ProviderConfig represents the configuration settings for a Webauthn Provider
type ProviderConfig struct {
	// Enabled is the provider enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// DisplayName is the site display name
	DisplayName string `json:"displayName" koanf:"displayName" jsonschema:"required" default:""`
	// RelyingPartyID is the relying party identifier
	// set to localhost for development, no port
	RelyingPartyID string `json:"relyingPartyId" koanf:"relyingPartyId" jsonschema:"required" default:"localhost"`
	// RequestOrigins the origin domain(s) for authentication requests
	// include the scheme and port
	RequestOrigins []string `json:"requestOrigins" koanf:"requestOrigins" jsonschema:"required"  default:"[http://localhost:3001]"`
	// MaxDevices is the maximum number of devices that can be associated with a user
	MaxDevices int `json:"maxDevices" koanf:"maxDevices" default:"10"`
	// EnforceTimeout at the Relying Party / Server. This means if enabled and the user takes too long that even if the browser does not
	// enforce a timeout, the server will
	EnforceTimeout bool `json:"enforceTimeout" koanf:"enforceTimeout" default:"true"`
	// Timeout is the timeout in seconds
	Timeout time.Duration `json:"timeout" koanf:"timeout" default:"60s"`
	// Debug enables debug mode
	Debug bool `json:"debug" koanf:"debug" default:"false"`
}

// NewWithConfig returns a configured Webauthn Provider
func NewWithConfig(config ProviderConfig) *webauthn.WebAuthn {
	if !config.Enabled {
		return nil
	}

	cfg := &webauthn.Config{
		RPID:          config.RelyingPartyID,
		RPOrigins:     config.RequestOrigins,
		RPDisplayName: config.DisplayName,
		Debug:         config.Debug,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    config.EnforceTimeout,
				Timeout:    config.Timeout,
				TimeoutUVD: config.Timeout,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    config.EnforceTimeout,
				Timeout:    config.Timeout,
				TimeoutUVD: config.Timeout,
			},
		},
	}

	return &webauthn.WebAuthn{Config: cfg}
}

func init() {
	// Register the webauthn.SessionData type with gob so it can be used in sessions
	gob.Register(webauthn.SessionData{})
}
