package corejobs

import (
	"time"
)

// CustomDomainConfig contains the configuration for the custom domain workers
type CustomDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for watermarking"`

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the custom domain worker is enabled"`

	CloudflareAPIKey string `koanf:"cloudflareapikey" json:"cloudflareapikey" jsonschema:"required description=the cloudflare api key" sensitive:"true"`

	ValidateInterval time.Duration `koanf:"validateinterval" json:"validateinterval" jsonschema:"required,default=5m description=the interval to validate custom domains"`
}
