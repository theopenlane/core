package corejobs

// PirschDomainConfig contains the configuration for the pirsch domain workers
type PirschDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for pirsch domain management"`

	// Enabled indicates whether the pirsch domain worker is enabled
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the pirsch domain worker is enabled"`

	// PirschClientID is the client id for the pirsch API
	PirschClientID string `koanf:"pirschClientID" json:"pirschClientID" jsonschema:"required description=the pirsch client id"`
	// PirschClientSecret is the client secret for the pirsch API
	PirschClientSecret string `koanf:"pirschClientSecret" json:"pirschClientSecret" jsonschema:"required description=the pirsch client secret" sensitive:"true"`
}
