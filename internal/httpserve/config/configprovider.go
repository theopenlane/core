package config

// Provider serves as a common interface to read echo server configuration
type Provider interface {
	// Get returns the server configuration
	Get() (*Config, error)
}
