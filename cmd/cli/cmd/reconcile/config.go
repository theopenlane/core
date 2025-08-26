//go:build cli

package reconcile

// Config defines configuration for the reconcile job
type Config struct {
	// ConfigPath is the location of the core configuration file
	ConfigPath string
	// DryRun prints actions without making changes when true
	DryRun bool
}

// Option configures the job config
type Option func(*Config)

// WithConfigPath sets the config file path
func WithConfigPath(path string) Option {
	return func(c *Config) {
		c.ConfigPath = path
	}
}

// WithDryRun sets the dry run flag
func WithDryRun(d bool) Option {
	return func(c *Config) {
		c.DryRun = d
	}
}

// NewConfig creates a new Config with provided options
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		ConfigPath: "./config/.config.yaml",
		DryRun:     true,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
