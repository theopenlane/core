package config

// ConfigProvider serves as a common interface to read echo server configuration
type ConfigProvider interface {
	// GetConfig returns the server configuration
	GetConfig() (*Config, error)
}
