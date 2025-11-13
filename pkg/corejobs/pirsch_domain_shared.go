package corejobs

// PirschDomainConfig contains the configuration for the pirsch domain workers
type PirschDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the pirsch domain worker is enabled"`

	PirschClientID     string `koanf:"pirschClientID" json:"pirschClientID" jsonschema:"required description=the pirsch client id"`
	PirschClientSecret string `koanf:"pirschClientSecret" json:"pirschClientSecret" jsonschema:"required description=the pirsch client secret"`

	DatabaseHost string `koanf:"databaseHost" json:"databaseHost" jsonschema:"required description=the database host"`
}
