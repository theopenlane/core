package serveropts

import (
	"github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
)

type ServerOptions struct {
	ConfigProvider serverconfig.ConfigProvider
	Config         serverconfig.Config
}

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

// AddServerOptions applies a server option after the initial setup
// this should be used when information is not available on NewServerOptions
func (so *ServerOptions) AddServerOptions(opt ServerOption) {
	opt.apply(so)
}
