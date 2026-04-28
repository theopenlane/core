package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	openlaneclient "github.com/theopenlane/go-client"

	"github.com/theopenlane/core/internal/integrations/cli/config"
	"github.com/theopenlane/core/internal/integrations/cli/openlane"
	"github.com/theopenlane/core/pkg/logx"
)

// healthCheckTimeout is the maximum time to wait for the server livez check
const healthCheckTimeout = 5 * time.Second

// ConnectClient verifies the server is reachable, then returns an authenticated
// Openlane client. When credentials mode is active (the default), login is
// attempted automatically using the configured email and password
func ConnectClient(ctx context.Context) (*openlaneclient.Client, error) {
	var cfg config.Config
	if err := Config.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	host := strings.TrimSpace(cfg.Openlane.Host)
	if host == "" {
		return nil, openlane.ErrHostRequired
	}

	if err := checkServer(ctx, host); err != nil {
		return nil, fmt.Errorf("server not reachable at %s: %w", host, err)
	}

	client, err := openlane.Connect(ctx, cfg.Openlane)
	if err != nil {
		return nil, err
	}

	logx.FromContext(ctx).Debug().Str("host", host).Msg("connected to openlane")

	return client, nil
}

// checkServer verifies the server is reachable by hitting the /livez endpoint
func checkServer(ctx context.Context, host string) error {
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, host+"/livez", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("livez returned %d", resp.StatusCode)
	}

	return nil
}
