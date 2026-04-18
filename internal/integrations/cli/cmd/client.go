package cmd

import (
	"context"

	openlaneclient "github.com/theopenlane/go-client"

	"github.com/theopenlane/core/internal/integrations/cli/config"
	"github.com/theopenlane/core/internal/integrations/cli/openlane"
)

// ConnectClient resolves the loaded koanf config into an OpenlaneConfig and
// returns an authenticated Openlane client. Subcommands should call this
// instead of building a client directly so auth mode and host resolution stay
// in one place.
func ConnectClient(ctx context.Context) (*openlaneclient.Client, error) {
	var cfg config.Config
	if err := Config.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return openlane.Connect(ctx, cfg.Openlane)
}
