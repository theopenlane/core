package serveropts

import (
	"github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
)

// ServerOptions defines the options for the server
type ServerOptions struct {
	// ConfigProvider is the provider for the server configuration
	ConfigProvider serverconfig.Provider
	// Config is the server configuration settings
	Config serverconfig.Config
}

// NewServerOptions creates a new ServerOptions instance with the provided options and configuration location
func NewServerOptions(opts []ServerOption, cfgLoc string) *ServerOptions {
	// load koanf config
	c, err := config.Load(&cfgLoc)
	if err != nil {
		panic(err)
	}

	so := &ServerOptions{
		Config: serverconfig.Config{
			Settings: *c,
		},
	}

	for _, opt := range opts {
		opt.apply(so)
	}

	return so
}

// AddServerOptions applies server options after the initial setup
// this should be used when information is not available on NewServerOptions
func (so *ServerOptions) AddServerOptions(opts ...ServerOption) {
	for _, opt := range opts {
		opt.apply(so)
	}
}
