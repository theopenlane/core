package corejobs

import (
	"time"
)

// CustomDomainConfig contains the configuration for the custom domain workers
type CustomDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the custom domain worker is enabled"`

	CloudflareAPIKey string `koanf:"cloudflareApiKey" json:"cloudflareApiKey" jsonschema:"required description=the cloudflare api key"`

	DatabaseHost     string        `koanf:"databaseHost" json:"databaseHost" jsonschema:"required description=the database host"`
	ValidateInterval time.Duration `koanf:"validateInterval" json:"validateInterval" jsonschema:"required,default=5m description=the interval to validate custom domains"`
}
